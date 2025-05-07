package viscaufs

import (
	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func AttrFromProto(attr *fuse.Attr, pbAttr *fspb.FileAttributes) {
	attr.Mode = pbAttr.Mode
	attr.Size = pbAttr.Size
	attr.Blocks = pbAttr.Blocks

	// Time values
	attr.Atime = pbAttr.Atime
	attr.Mtime = pbAttr.Mtime
	attr.Ctime = pbAttr.Ctime
	attr.Atimensec = uint32(pbAttr.Atimensec)
	attr.Mtimensec = uint32(pbAttr.Mtimensec)
	attr.Ctimensec = uint32(pbAttr.Ctimensec)

	// Other attributes
	attr.Nlink = pbAttr.Nlink
	attr.Uid = pbAttr.Uid
	attr.Gid = pbAttr.Gid
	attr.Rdev = uint32(pbAttr.Rdev)
	attr.Blksize = pbAttr.Blksize

	// Set inode number if available
	if pbAttr.Inode > 0 {
		attr.Ino = pbAttr.Inode
	}
}
