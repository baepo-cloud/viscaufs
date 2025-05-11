package fsindex

import (
	"bytes"
	"path/filepath"

	art "github.com/alexisvisco/go-adaptive-radix-tree/v2"
)

// numberOfSlashesFromPrefix counts how many slashes are in 'current' after the 'key' prefix.
func numberOfSlashesFromPrefix(prefix, current art.Key) int {
	// Remove trailing slash from prefix if present (unless it's root)
	cleanPrefix := prefix
	if len(prefix) > 1 && prefix[len(prefix)-1] == filepath.Separator {
		cleanPrefix = prefix[:len(prefix)-1]
	}

	cleanCurrent := current
	if len(current) > 1 && current[len(current)-1] == filepath.Separator {
		cleanCurrent = current[:len(current)-1]
	}

	if bytes.Equal(cleanPrefix, cleanCurrent) {
		return 0
	}

	if !bytes.HasPrefix(cleanCurrent, cleanPrefix) {
		return -1
	}

	trimmed := bytes.TrimPrefix(cleanCurrent, cleanPrefix)
	paths := bytes.Split(trimmed, []byte{filepath.Separator})
	count := len(paths)
	for _, path := range paths {
		if bytes.Equal(path, []byte{}) {
			count--
		}
	}

	return count
}
