package viscaufs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// FS represents our FUSE filesystem
type FS struct {
	Client      fspb.FuseServiceClient
	ImageDigest string
	MountPath   string
	Cache       *Cache
}

// Node directly implements FS interfaces
type Node struct {
	fs.Inode
	FS            *FS
	Path          string
	SymlinkTarget *string
}

type FileHandle struct {
	Uid string
}

// Ensure interfaces are implemented
var (
	_ fs.NodeGetattrer  = (*Node)(nil)
	_ fs.NodeLookuper   = (*Node)(nil)
	_ fs.NodeReaddirer  = (*Node)(nil)
	_ fs.NodeReadlinker = (*Node)(nil)
	_ fs.NodeOpener     = (*Node)(nil)
	_ fs.NodeReader     = (*Node)(nil)
	_ fs.NodeReleaser   = (*Node)(nil)
)

// Getattr implementation
func (n *Node) Getattr(ctx context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fmt.Fprintf(os.Stderr, "DEBUG: Getattr called for Path: %s\n", n.Path)

	file, err := n.FS.Cache.GetOrFetchAttr(n.Path, func() (*fspb.GetAttrResponse, error) {
		return n.FS.Client.GetAttr(ctx, &fspb.GetAttrRequest{
			Path:        n.Path,
			ImageDigest: n.FS.ImageDigest,
		})
	})

	if err != nil {
		slog.Info("getattr: error", "path", n.Path, "err", err)
		return syscall.ENOENT
	}

	AttrFromProto(&out.Attr, file.Attributes)

	return 0
}

// Lookup implementation
func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	childPath := filepath.Join(n.Path, name)
	if !strings.HasPrefix(childPath, "/") {
		childPath = "/" + childPath
	}

	file, err := n.FS.Cache.GetOrFetchAttr(childPath, func() (*fspb.GetAttrResponse, error) {
		return n.FS.Client.GetAttr(ctx, &fspb.GetAttrRequest{
			Path:        childPath,
			ImageDigest: n.FS.ImageDigest,
		})
	})

	if err != nil {
		slog.Info("lookup: error", "path", n.Path, "err", err)
		return nil, syscall.ENOENT
	}

	AttrFromProto(&out.Attr, file.Attributes)

	child := &Node{
		FS:            n.FS,
		Path:          childPath,
		SymlinkTarget: file.SymlinkTarget,
	}

	childInode := n.NewPersistentInode(ctx, child, fs.StableAttr{
		Mode: file.Attributes.Mode,
		Ino:  file.Attributes.Inode,
	})

	return childInode, 0
}

// Readdir implementation
func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Fprintf(os.Stderr, "DEBUG: Readdir called for Path: %s\n", n.Path)

	files, err := n.FS.Cache.GetOrFetchDir(n.Path, func() (*fspb.ReadDirResponse, error) {
		return n.FS.Client.ReadDir(ctx, &fspb.ReadDirRequest{
			Path:        n.Path,
			ImageDigest: n.FS.ImageDigest,
		})
	})

	if err != nil {
		slog.Error("readdir: error", "path", n.Path, "err", err)
		return nil, syscall.EIO
	}

	entries := make([]fuse.DirEntry, 0, len(files)+2)

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

	for _, entry := range files {
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

func (n *Node) Readlink(_ context.Context) ([]byte, syscall.Errno) {
	// Check if this node is actually a symlink
	if n.StableAttr().Mode&syscall.S_IFMT != syscall.S_IFLNK {
		return nil, syscall.EINVAL
	}

	fmt.Sprintf("ASKING CACHE FOR SYMLINK: %s", n.Path)

	if n.SymlinkTarget == nil {
		return nil, syscall.ENOENT
	}

	return []byte(filepath.Join(n.FS.MountPath, *n.SymlinkTarget)), 0
}

func (n *Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	resp, err := n.FS.Client.Open(ctx, &fspb.OpenRequest{
		Path:        n.Path,
		Flags:       flags,
		ImageDigest: n.FS.ImageDigest,
	})

	if err != nil {
		slog.Error("open: error", "path", n.Path, "err", err)
		return nil, 0, syscall.EIO
	}

	handle := &FileHandle{
		Uid: resp.Uid,
	}

	return handle, fuse.FOPEN_KEEP_CACHE, 0
}

func (n *Node) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fh, ok := f.(*FileHandle)
	if !ok {
		return nil, syscall.EINVAL
	}

	resp, err := n.FS.Client.Read(ctx, &fspb.ReadRequest{
		Uid:    fh.Uid,
		Offset: off,
		Size:   uint32(len(dest)),
	})

	if err != nil {
		slog.Error("read: error", "path", n.Path, "err", err)
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(resp.Data), 0
}

func (n *Node) Release(ctx context.Context, f fs.FileHandle) syscall.Errno {
	fh, ok := f.(*FileHandle)
	if !ok {
		return syscall.EINVAL
	}

	_, err := n.FS.Client.Release(ctx, &fspb.ReleaseRequest{
		Uid: fh.Uid,
	})

	if err != nil {
		slog.Error("release: error", "path", n.Path, "err", err)
		return syscall.EIO
	}

	return 0
}
