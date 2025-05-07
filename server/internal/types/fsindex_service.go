package types

import (
	"github.com/baepo-cloud/viscaufs-server/internal/fsindex"
)

type (
	FileSystemIndexLayer struct {
		Digest         string
		Position       uint8
		SerializedData []byte
	}

	FileSystemIndexService interface {
		CreateImageIndexChannel(imageDigest string) chan<- FileSystemIndexLayer
		BuildImageIndex(inspect *Image, digestToPosition map[string]uint8)
		BuildLayerIndex(path, layerDigest string) ([]byte, error)

		Lookup(imageDigest, path string) *fsindex.FSNode
		LookupByPrefix(imageDigest, path string) []*fsindex.FSNode
		Ready(imageDigest string) bool
	}
)
