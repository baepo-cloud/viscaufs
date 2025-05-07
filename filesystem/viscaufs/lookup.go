package viscaufs

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Lookup implements path resolution for directories
func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Construct the full path for the child
	childPath := n.path
	if childPath != "/" {
		childPath += "/"
	}
	childPath += name

	childPath = filepath.Clean(childPath)
	childPath = strings.ReplaceAll(childPath, "\\", "/") // Ensure forward slashes

	slog.Info("lookup", "path", childPath)

	resp, err := n.fs.Client.GetAttr(ctx, &fspb.GetAttrRequest{
		Path:        childPath,
		ImageDigest: n.fs.ImageDigest,
	})

	if err != nil {
		return nil, syscall.ENOENT
	}

	AttrFromProto(&out.Attr, resp.Attributes)

	child := &Node{
		fs:   n.fs,
		path: childPath,
	}

	var mode uint32 = syscall.S_IFREG
	if (resp.Attributes.Mode & uint32(syscall.S_IFDIR)) != 0 {
		mode = syscall.S_IFDIR
	} else if (resp.Attributes.Mode & uint32(syscall.S_IFLNK)) != 0 {
		mode = syscall.S_IFLNK
	}

	childInode := n.NewInode(ctx, child, fs.StableAttr{
		Mode: mode,
		Ino:  resp.Attributes.Inode,
	})

	return childInode, 0
}
