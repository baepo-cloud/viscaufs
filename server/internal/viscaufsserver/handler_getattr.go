package viscaufsserver

import (
	"context"
	"os"
	"time"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) GetAttr(ctx context.Context, request *fspb.GetAttrRequest) (*fspb.GetAttrResponse, error) {
	if request.Path == "/" {
		// Create hardcoded attributes for root directory
		rootAttrs := &fspb.FileAttributes{
			Inode:     1,                 // Root directory typically has inode 1
			Size:      4096,              // Standard size for directories
			Blocks:    8,                 // Typical number of blocks for 4096 bytes
			Atime:     time.Now().Unix(), // Current time
			Mtime:     time.Now().Unix(), // Current time
			Ctime:     time.Now().Unix(), // Current time
			Atimensec: 0,
			Mtimensec: 0,
			Ctimensec: 0,
			Mode:      uint32(0755) | uint32(os.ModeDir), // Directory with standard permissions
			Nlink:     2,                                 // . and .. entries at minimum
			Uid:       0,                                 // Root user
			Gid:       0,                                 // Root group
			Rdev:      0,                                 // Not a device file
			Blksize:   4096,                              // Standard block size
		}

		return &fspb.GetAttrResponse{
			File: &fspb.File{
				Path:       request.Path,
				Attributes: rootAttrs,
			},
		}, nil
	}

	lookup := s.FSIndexerService.Lookup(ctx, request.ImageDigest, request.Path)
	if lookup == nil {
		return nil, status.Error(codes.NotFound, "path not found")
	}

	proto := lookup.ToProto()
	return &fspb.GetAttrResponse{
		File: &fspb.File{
			Path:          request.Path,
			Attributes:    proto.Attributes,
			SymlinkTarget: proto.SymlinkTarget,
		},
	}, nil
}
