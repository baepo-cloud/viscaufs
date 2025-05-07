package viscaufs

import (
	"context"
	"log/slog"
	"path/filepath"
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Readdir implements reading directory entries
func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	// Skip if not a directory
	if !n.IsDir() {
		return nil, syscall.ENOTDIR
	}

	slog.Info("readdir", "path", n.path)

	resp, err := n.fs.Client.ReadDir(ctx, &fspb.ReadDirRequest{
		Path:        n.path,
		ImageDigest: n.fs.ImageDigest,
	})

	if err != nil {
		return nil, syscall.EIO
	}

	entries := make([]fuse.DirEntry, 0, len(resp.Entries))
	offset := uint64(0)
	for _, entry := range resp.Entries {
		name := filepath.Base(entry.Path)

		if name == "" || name == "." || name == ".." {
			continue
		}

		entries = append(entries, fuse.DirEntry{
			Mode: entry.Attributes.Mode,
			Name: name,
			Ino:  entry.Attributes.Inode,
			Off:  offset,
		})

		offset++
	}

	return fs.NewListDirStream(entries), 0
}
