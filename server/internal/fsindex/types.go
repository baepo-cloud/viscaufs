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
	Blocks    uint64
	Atime     uint64
	Mtime     uint64
	Ctime     uint64
	Atimensec uint32
	Mtimensec uint32
	Ctimensec uint32
	Mode      uint32
	Nlink     uint32
	Owner     struct {
		Uid uint32
		Gid uint32
	}
	Rdev    uint32
	Blksize uint32
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
			Size:      uint64(f.Attributes.Size),
			Blocks:    f.Attributes.Blocks,
			Atime:     f.Attributes.Atime,
			Mtime:     f.Attributes.Mtime,
			Ctime:     f.Attributes.Ctime,
			Atimensec: uint64(f.Attributes.Atimensec),
			Mtimensec: uint64(f.Attributes.Mtimensec),
			Ctimensec: uint64(f.Attributes.Ctimensec),
			Mode:      f.Attributes.Mode,
			Nlink:     f.Attributes.Nlink,
			Uid:       f.Attributes.Owner.Uid,
			Gid:       f.Attributes.Owner.Gid,
			Rdev:      uint64(f.Attributes.Rdev),
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
			Size:      int64(node.Attributes.Size),
			Blocks:    node.Attributes.Blocks,
			Atime:     node.Attributes.Atime,
			Mtime:     node.Attributes.Mtime,
			Ctime:     node.Attributes.Ctime,
			Atimensec: uint32(node.Attributes.Atimensec),
			Mtimensec: uint32(node.Attributes.Mtimensec),
			Ctimensec: uint32(node.Attributes.Ctimensec),
			Mode:      node.Attributes.Mode,
			Nlink:     node.Attributes.Nlink,
			Owner: struct {
				Uid uint32
				Gid uint32
			}{
				Uid: node.Attributes.Uid,
				Gid: node.Attributes.Gid,
			},
			Rdev:    uint32(node.Attributes.Rdev),
			Blksize: node.Attributes.Blksize,
		},
		LayerPosition: uint8(node.LayerPosition),
	}
}
