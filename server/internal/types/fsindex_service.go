package types

import (
	"context"
	"github.com/baepo-cloud/viscaufs/common/fsindex"
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

		Lookup(ctx context.Context, imageDigest, path string) *fsindex.Node
		LookupByPrefix(ctx context.Context, imageDigest, path string) []*fsindex.Node
		Ready(imageDigest string) bool
	}
)
