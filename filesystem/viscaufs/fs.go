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
	_ fs.NodeSymlinker = (*Node)(nil)
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
	childPath := filepath.Join(n.Path, name)
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

	childInode := n.NewInode(ctx, child, fs.StableAttr{
		Mode: resp.Attributes.Mode,
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

	entries := make([]fuse.DirEntry, 0, len(resp.Entries)+2)

	entries = append(entries, fuse.DirEntry{
		Name: ".",
		Mode: syscall.S_IFDIR,
		Ino:  n.StableAttr().Ino,
	})

	_, inode := n.Parent()
	if inode != nil {
		entries = append(entries, fuse.DirEntry{
			Name: "..",
			Mode: syscall.S_IFDIR,
			Ino:  inode.StableAttr().Ino,
		})
	}

	for _, entry := range resp.Entries {
		// Get the base name from the full path
		name := filepath.Base(entry.Path)
		if name == "" || name == "." || name == ".." {
			continue
		}

		var mode uint32
		if entry.Attributes != nil {
			mode = entry.Attributes.Mode
		}

		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: mode,
			Ino:  entry.Attributes.Inode,
		})
	}

	return fs.NewListDirStream(entries), 0
}

func (n *Node) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	// todo implement

	return nil, syscall.ENOSYS
}
