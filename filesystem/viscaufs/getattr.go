package viscaufs

import (
	"context"
	"log/slog"
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func (n *Node) Getattr(ctx context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	slog.Info("getattr", "path", n.path)

	resp, err := n.fs.Client.GetAttr(ctx, &fspb.GetAttrRequest{
		Path:        n.path,
		ImageDigest: n.fs.ImageDigest,
	})

	if err != nil {
		return syscall.ENOENT
	}

	AttrFromProto(&out.Attr, resp.Attributes)

	return 0
}
