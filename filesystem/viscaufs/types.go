package viscaufs

import (
	"context"
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
)

// FS represents our FUSE filesystem
type FS struct {
	Client      fspb.FuseServiceClient
	ImageDigest string
}

// Node represents a file or directory in the filesystem
type Node struct {
	fs.Inode
	fs   *FS
	path string
}

// Root node of the filesystem
type Root struct {
	fs.Inode
	FileSystem *FS
}

// OnMount is called when the filesystem is mounted
func (r *Root) OnMount(ctx context.Context) {
	// Create the root directory node
	r.NewPersistentInode(
		ctx,
		&Node{
			fs:   r.FileSystem,
			path: "/",
		},
		fs.StableAttr{Mode: syscall.S_IFDIR},
	)
}
