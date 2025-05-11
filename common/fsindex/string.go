package fsindex

import (
	"fmt"
	"github.com/baepo-cloud/viscaufs/common/humanize"
	"path/filepath"
	"sort"
	"strings"

	art "github.com/alexisvisco/go-adaptive-radix-tree/v2"
)

// String returns a string representation of the Index as a tree
func (idx *Index) String() string {
	var sb strings.Builder
	// Get all paths from the Trie
	type nodeInfo struct {
		path  string
		depth int
		node  *Node
		last  bool // whether this is the last child at its level
	}

	// Collect all paths and sort them
	var paths []nodeInfo
	idx.Trie.ForEach(func(node art.NodeKV) (cont bool) {
		path := string(node.Key())
		fsNode, ok := node.Value().(*Node)
		if ok {
			depth := strings.Count(path, "/")
			if path == "" {
				depth = 0
			}
			paths = append(paths, nodeInfo{path: path, depth: depth, node: fsNode})
		}
		return true
	})

	// Sort paths by their string value which will naturally organize the tree structure
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].path < paths[j].path
	})

	// Mark last nodes at each level
	if len(paths) > 0 {
		// Find the last node at each depth
		depthMap := make(map[int][]int) // depth -> indices at that depth
		for i, p := range paths {
			depthMap[p.depth] = append(depthMap[p.depth], i)
		}

		// Mark the last index at each depth
		for _, indices := range depthMap {
			if len(indices) > 0 {
				lastIdx := indices[len(indices)-1]
				paths[lastIdx].last = true
			}
		}
	}

	// Track which levels have remaining children
	activeLevels := make(map[int]bool)

	// Print the tree
	for i, info := range paths {
		path := info.path
		depth := info.depth
		node := info.node

		// Skip the root node if it exists
		if path == "" {
			continue
		}

		// Get the directory name or file name
		name := filepath.Base(path)

		// Update active levels
		for j := 0; j < depth; j++ {
			// Check if there are more nodes at this level
			hasMore := false
			for k := i + 1; k < len(paths); k++ {
				if paths[k].depth >= j {
					if paths[k].depth == j {
						hasMore = true
						break
					}
				} else {
					break
				}
			}
			activeLevels[j] = hasMore
		}

		// Build the line prefix based on tree structure
		for j := 0; j < depth; j++ {
			if j == depth-1 {
				if info.last {
					sb.WriteString("└── ")
				} else {
					sb.WriteString("├── ")
				}
			} else {
				if activeLevels[j] {
					sb.WriteString("│   ")
				} else {
					sb.WriteString("    ")
				}
			}
		}

		// Add node type indicator
		if node.IsDirectory() {
			sb.WriteString("📁 ")
		} else {
			sb.WriteString("📄 ")
		}

		// Write the name and add file size for non-directories
		link := "none"
		if node.SymlinkTarget != nil {
			link = *node.SymlinkTarget
		}
		if node.IsDirectory() {
			sb.WriteString(fmt.Sprintf("%s (L: %d, S: %s)", name, node.LayerPosition, link))
		} else {
			size := humanize.Bytes(node.Attributes.Size)
			sb.WriteString(fmt.Sprintf("%s (%s, L: %d, S: %s)", name, size, node.LayerPosition, link))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
