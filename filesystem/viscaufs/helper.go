package viscaufs

import (
	"github.com/baepo-cloud/viscaufs/common/fsindex"
	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func AttrFromProto(attr *fuse.Attr, pbAttr *fspb.FileAttributes) {
	attr.Mode = pbAttr.Mode
	attr.Size = uint64(pbAttr.Size)
	attr.Blocks = uint64(pbAttr.Blocks)

	// Time values
	attr.Atime = uint64(pbAttr.Atime)
	attr.Mtime = uint64(pbAttr.Mtime)
	attr.Ctime = uint64(pbAttr.Ctime)
	attr.Atimensec = uint32(pbAttr.Atimensec)
	attr.Mtimensec = uint32(pbAttr.Mtimensec)
	attr.Ctimensec = uint32(pbAttr.Ctimensec)

	// Other attributes
	attr.Nlink = uint32(pbAttr.Nlink)
	attr.Uid = pbAttr.Uid
	attr.Gid = pbAttr.Gid
	attr.Rdev = uint32(pbAttr.Rdev)
	attr.Blksize = uint32(pbAttr.Blksize)

	// Set inode number if available
	if pbAttr.Inode > 0 {
		attr.Ino = pbAttr.Inode
	}
}

func AttrFromFSIndex(attr *fuse.Attr, fsindexAttr fsindex.FileAttributes) {
	attr.Mode = fsindexAttr.Mode
	attr.Size = uint64(fsindexAttr.Size)
	attr.Blocks = uint64(fsindexAttr.Blocks)

	// Time values
	attr.Atime = uint64(fsindexAttr.Atime)
	attr.Mtime = uint64(fsindexAttr.Mtime)
	attr.Ctime = uint64(fsindexAttr.Ctime)
	attr.Atimensec = uint32(fsindexAttr.Atimensec)
	attr.Mtimensec = uint32(fsindexAttr.Mtimensec)
	attr.Ctimensec = uint32(fsindexAttr.Ctimensec)

	// Other attributes
	attr.Nlink = uint32(fsindexAttr.Nlink)
	attr.Uid = fsindexAttr.Owner.Uid
	attr.Gid = fsindexAttr.Owner.Gid
	attr.Rdev = uint32(fsindexAttr.Rdev)
	attr.Blksize = uint32(fsindexAttr.Blksize)

	// Set inode number if available
	if fsindexAttr.Inode > 0 {
		attr.Ino = fsindexAttr.Inode
	}
}
