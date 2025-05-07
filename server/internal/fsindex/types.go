package fsindex

import (
	"syscall"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	art "github.com/plar/go-adaptive-radix-tree/v2"
)

// FSNode represents a node in our filesystem index
type FSNode struct {
	Path          string
	Attributes    FileAttributes
	LayerPosition uint8
}

type FileAttributes struct {
	Inode     uint64
	Size      int64
	Blocks    int64
	Atime     int64
	Mtime     int64
	Ctime     int64
	Atimensec int64
	Mtimensec int64
	Ctimensec int64
	Mode      uint32
	Nlink     uint64
	Owner     struct {
		Uid uint32
		Gid uint32
	}
	Rdev    uint64
	Blksize int64
}

// FSIndex represents our optimized filesystem index
type FSIndex struct {
	// Adaptive Radix Tree for fast lookup
	trie         art.Tree
	withoutFiles map[string]struct{}
	withoutDirs  map[string]struct{}
}

func (f *FSNode) IsDirectory() bool {
	return f.Attributes.Mode&syscall.S_IFMT == syscall.S_IFDIR
}

func (f *FSNode) ToProto() *fspb.FSNode {
	return &fspb.FSNode{
		Path: f.Path,
		Attributes: &fspb.FileAttributes{
			Inode:     f.Attributes.Inode,
			Size:      f.Attributes.Size,
			Blocks:    f.Attributes.Blocks,
			Atime:     f.Attributes.Atime,
			Mtime:     f.Attributes.Mtime,
			Ctime:     f.Attributes.Ctime,
			Atimensec: f.Attributes.Atimensec,
			Mtimensec: f.Attributes.Mtimensec,
			Ctimensec: f.Attributes.Ctimensec,
			Mode:      f.Attributes.Mode,
			Nlink:     f.Attributes.Nlink,
			Uid:       f.Attributes.Owner.Uid,
			Gid:       f.Attributes.Owner.Gid,
			Rdev:      f.Attributes.Rdev,
			Blksize:   f.Attributes.Blksize,
		},
		LayerPosition: uint32(f.LayerPosition),
	}
}

func FSNodeFromProto(node *fspb.FSNode) *FSNode {
	return &FSNode{
		Path: node.Path,
		Attributes: FileAttributes{
			Inode:     node.Attributes.Inode,
			Size:      node.Attributes.Size,
			Blocks:    node.Attributes.Blocks,
			Atime:     node.Attributes.Atime,
			Mtime:     node.Attributes.Mtime,
			Ctime:     node.Attributes.Ctime,
			Atimensec: node.Attributes.Atimensec,
			Mtimensec: node.Attributes.Mtimensec,
			Ctimensec: node.Attributes.Ctimensec,
			Mode:      node.Attributes.Mode,
			Nlink:     node.Attributes.Nlink,
			Owner: struct {
				Uid uint32
				Gid uint32
			}{
				Uid: node.Attributes.Uid,
				Gid: node.Attributes.Gid,
			},
			Rdev:    node.Attributes.Rdev,
			Blksize: node.Attributes.Blksize,
		},
		LayerPosition: uint8(node.LayerPosition),
	}
}
