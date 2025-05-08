package fsindex

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	art "github.com/plar/go-adaptive-radix-tree/v2"
)

// NewFSIndex creates a new filesystem index
func NewFSIndex() *Index {
	return &Index{
		Trie:         art.New(),
		withoutFiles: make(map[string]struct{}),
		withoutDirs:  make(map[string]struct{}),
	}
}

// BuildIndex builds the index from a root directory
func (idx *Index) BuildIndex(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %q: %w", path, err)
		}

		relPath, relErr := filepath.Rel(rootDir, path)
		if relErr != nil {
			return fmt.Errorf("failed to compute relative path for %q: %w", path, relErr)
		}

		if relPath == "." {
			return nil
		}

		// Create the node first
		node := &Node{
			Path:       cleanPath(relPath),
			Attributes: collectFileAttributes(info),
		}

		// If it's a symlink, read the target
		if (info.Mode() & os.ModeSymlink) != 0 {
			target, err := os.Readlink(path)
			if err == nil {
				target = cleanPath(target)
				node.SymlinkTarget = &target
			} else {
				slog.Error("unable to read symlink target", "path", path, "error", err.Error())
			}
		}

		// Add the node to the index
		idx.Trie.Insert(art.Key(node.Path), node)

		if strings.Contains(relPath, ".wh.") {
			idx.withoutFiles[relPath] = struct{}{}
		}

		if strings.Contains(relPath, ".wh..wh.") {
			idx.withoutDirs[relPath] = struct{}{}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	return nil
}
func collectFileAttributes(info os.FileInfo) FileAttributes {
	stat := info.Sys().(*syscall.Stat_t)

	return FileAttributes{
		Inode:     stat.Ino,
		Size:      info.Size(),
		Blocks:    stat.Blocks,
		Atime:     stat.Atim.Sec,
		Atimensec: stat.Atim.Nsec,
		Mtime:     stat.Mtim.Sec,
		Mtimensec: stat.Mtim.Nsec,
		Ctime:     stat.Ctim.Sec,
		Ctimensec: stat.Ctim.Nsec,
		Mode:      stat.Mode,
		Nlink:     stat.Nlink,
		Owner: struct {
			Uid uint32
			Gid uint32
		}{
			Uid: stat.Uid,
			Gid: stat.Gid,
		},
		Rdev:    stat.Rdev,
		Blksize: stat.Blksize,
	}
}

// LookupPath looks up a path in the index
func (idx *Index) LookupPath(path string) (*Node, error) {
	path = filepath.ToSlash(filepath.Clean(path))

	value, found := idx.Trie.Search(art.Key(path))
	if !found {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	node, ok := value.(*Node)
	if !ok {
		return nil, fmt.Errorf("invalid node type for path: %s", path)
	}

	return node, nil
}

// LookupPrefixSearch performs a prefix search on the index
// Only returns immediate children (depth 1) of the given prefix
func (idx *Index) LookupPrefixSearch(prefix string) []*Node {
	prefix = filepath.ToSlash(filepath.Clean(prefix))

	// Ensure prefix ends with a slash if it's not empty
	// This ensures we're looking for children of the directory
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var results []*Node
	prefixKey := art.Key(prefix)
	prefixLen := len(prefix)

	idx.Trie.ForEachPrefix(prefixKey, func(node art.Node) bool {
		nodePath := string(node.Key())

		// Skip the prefix node itself
		if nodePath == prefix {
			return true
		}

		// Calculate the relative path from the prefix
		relPath := nodePath[prefixLen:]

		// Only include direct children (no additional slashes in the relative path)
		if !strings.Contains(relPath, "/") {
			if fsNode, ok := node.Value().(*Node); ok {
				results = append(results, fsNode)
			}
		}

		return true
	})

	return results
}

// GetStats returns statistics about the index
func (idx *Index) GetStats() map[string]interface{} {
	var totalFiles, totalDirs int

	idx.Trie.ForEach(func(node art.Node) (cont bool) {
		val := node.Value()
		if val == nil {
			return true
		}
		if fsNode, ok := val.(*Node); ok {
			if fsNode.IsDirectory() {
				totalDirs++
			} else {
				totalFiles++
			}
		}
		return true
	})

	return map[string]interface{}{
		"total_files":       totalFiles,
		"total_directories": totalDirs,
	}
}

func cleanPath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	if path != "" && !strings.HasSuffix(path, "/") {
		path = "/" + path
	}
	return filepath.Clean(path)
}

// addPath adds a relative path to the index (for testing purposes)
func (idx *Index) addPath(relPath string, info os.FileInfo) {
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + relPath
	}

	node := &Node{
		Path:       relPath,
		Attributes: collectFileAttributes(info),
	}

	idx.Trie.Insert(art.Key(relPath), node)

	if strings.Contains(relPath, ".wh.") {
		idx.withoutFiles[relPath] = struct{}{}
	}

	if strings.Contains(relPath, ".wh..wh.") {
		idx.withoutDirs[relPath] = struct{}{}
	}
}
