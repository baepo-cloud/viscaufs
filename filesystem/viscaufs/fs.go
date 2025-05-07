package viscaufs

import (
	"context"
	"fmt"
	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// FS represents our FUSE filesystem
type FS struct {
	Client      fspb.FuseServiceClient
	ImageDigest string
}

// Node directly implements FS interfaces
type Node struct {
	fs.Inode
	FS   *FS
	Path string
}

// Ensure interfaces are implemented
var (
	_ fs.NodeGetattrer = (*Node)(nil)
	_ fs.NodeLookuper  = (*Node)(nil)
	_ fs.NodeReaddirer = (*Node)(nil)
)

// Getattr implementation
func (n *Node) Getattr(ctx context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fmt.Fprintf(os.Stderr, "DEBUG: Getattr called for Path: %s\n", n.Path)

	resp, err := n.FS.Client.GetAttr(ctx, &fspb.GetAttrRequest{
		Path:        n.Path,
		ImageDigest: n.FS.ImageDigest,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: GetAttr error: %v\n", err)
		return syscall.ENOENT
	}

	AttrFromProto(&out.Attr, resp.Attributes)
	return 0
}

// Lookup implementation
func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	childPath := n.Path
	childPath = filepath.Join(n.Path, name)
	if !strings.HasPrefix(childPath, "/") {
		childPath = "/" + childPath
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Lookup called for Path: %s\n", childPath)

	resp, err := n.FS.Client.GetAttr(ctx, &fspb.GetAttrRequest{
		Path:        childPath,
		ImageDigest: n.FS.ImageDigest,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Lookup error: %v\n", err)
		return nil, syscall.ENOENT
	}

	AttrFromProto(&out.Attr, resp.Attributes)

	child := &Node{
		FS:   n.FS,
		Path: childPath,
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

// Readdir implementation
func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Fprintf(os.Stderr, "DEBUG: Readdir called for Path: %s\n", n.Path)

	resp, err := n.FS.Client.ReadDir(ctx, &fspb.ReadDirRequest{
		Path:        n.Path,
		ImageDigest: n.FS.ImageDigest,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: ReadDir error: %v\n", err)
		return nil, syscall.EIO
	}

	entries := []fuse.DirEntry{}
	for _, entry := range resp.Entries {
		name := entry.Path
		if name == "" || name == "." || name == ".." {
			continue
		}

		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: entry.Attributes.Mode,
			Ino:  entry.Attributes.Inode,
		})
	}

	return fs.NewListDirStream(entries), 0
}
