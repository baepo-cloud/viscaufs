package fsindex

import (
	"path/filepath"
	"strings"

	art "github.com/plar/go-adaptive-radix-tree/v2"
)

// JoinFSIndex merges two FSIndex instances.
// It takes the previous layer and adds the new layer on top of it, overriding existing nodes.
// It will remove every directory and file marked as "without" in the old layer.
// This implements the overlay filesystem semantics used in container images.
//
// Example to merge 3 layers:
// LAYER 0
// LAYER 1
// LAYER 2
//
// JoinFSIndex(LAYER 1, LAYER 2, 1, true) = LAYER 1 -> the merged layer
// JoinFSIndex(LAYER 0, LAYER 1, 0, false) = LAYER 0 -> the merged layer
func JoinFSIndex(currentLayerFSIndex, applyLayerFSIndex *FSIndex, currentLayerPosition uint8, firstJoin bool) {
	currentLayerFSIndex.Trie.ForEach(func(node art.Node) (cont bool) {
		fsNode, ok := node.Value().(*Node)
		if ok {
			fsNode.LayerPosition = currentLayerPosition
		}

		return true
	})

	if firstJoin {
		applyLayerFSIndex.Trie.ForEach(func(node art.Node) (cont bool) {
			fsNode, ok := node.Value().(*Node)
			if ok {
				fsNode.LayerPosition = currentLayerPosition + 1
			}

			return true
		})
	}

	// Process file whiteouts (.wh. files)
	for filePath := range applyLayerFSIndex.withoutFiles {
		// For a path like "a/b/c/.wh.file.json", we need to:
		// 1. Get the directory path: "a/b/c/"
		// 2. Get the filename: ".wh.file.json"
		// 3. Remove the ".wh." prefix from the filename: "file.json"
		// 4. Combine them back: "a/b/c/file.json"

		dirPath := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)

		// Make sure we're actually dealing with a whiteout file
		if !strings.HasPrefix(fileName, ".wh.") {
			continue
		}

		// Remove the ".wh." prefix from the filename
		realFileName := strings.TrimPrefix(fileName, ".wh.")

		// Construct the real path that should be removed
		var realPath string
		if dirPath == "." {
			realPath = realFileName
		} else {
			realPath = dirPath + "/" + realFileName
		}

		// Remove the target file from the previous layer
		currentLayerFSIndex.Trie.Delete(art.Key(realPath))

		// Find all children of this path if it was a directory and remove them
		prefixKey := art.Key(realPath + "/")
		var keysToDelete []art.Key

		// Find all keys that need to be deleted
		currentLayerFSIndex.Trie.ForEachPrefix(prefixKey, func(node art.Node) bool {
			keysToDelete = append(keysToDelete, node.Key())
			return true
		})

		// Delete all identified keys
		for _, key := range keysToDelete {
			currentLayerFSIndex.Trie.Delete(key)
		}
	}

	// Process directory opaque whiteouts (.wh..wh.)
	for dirPath := range applyLayerFSIndex.withoutDirs {
		// For a path like "a/b/c/.wh..wh.opq", we need to:
		// 1. Get the directory path: "a/b/c/"
		// 2. This directory should have its contents removed

		// Get the parent directory that needs to be made opaque
		opaqueDir := filepath.Dir(dirPath)

		if opaqueDir == "." {
			opaqueDir = ""
		}

		// For opaque directories, we keep the directory itself but remove all its contents
		if opaqueDir != "" {
			prefixKey := art.Key(opaqueDir + "/")
			var keysToDelete []art.Key

			// Find all keys that need to be deleted
			currentLayerFSIndex.Trie.ForEachPrefix(prefixKey, func(node art.Node) bool {
				nodePath := string(node.Key())
				// Don't delete the directory itself
				if nodePath != opaqueDir && strings.HasPrefix(nodePath, string(prefixKey)) {
					keysToDelete = append(keysToDelete, node.Key())
				}
				return true
			})

			// Delete all identified keys
			for _, key := range keysToDelete {
				currentLayerFSIndex.Trie.Delete(key)
			}
		}
	}

	// Add or override files from the new layer
	applyLayerFSIndex.Trie.ForEach(func(node art.Node) bool {
		fsNode, ok := node.Value().(*Node)
		if ok {
			fileName := filepath.Base(fsNode.Path)

			// Skip whiteout files and opaque dir markers as they've already been processed
			if strings.HasPrefix(fileName, ".wh.") {
				return true
			}

			// Add or override the file/directory in the previous layer
			currentLayerFSIndex.Trie.Insert(node.Key(), fsNode)
		}
		return true
	})
}
