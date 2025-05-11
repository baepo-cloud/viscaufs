package fsindex

import (
	"syscall"

	art "github.com/alexisvisco/go-adaptive-radix-tree/v2"
	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
)

// Node represents a node in our filesystem index
type Node struct {
	Path          string
	Attributes    FileAttributes
	LayerPosition uint8
	SymlinkTarget *string
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

func (f *Node) IsDirectory() bool {
	return f.Attributes.Mode&syscall.S_IFMT == syscall.S_IFDIR
}

func (f *Node) IsSymlink() bool {
	return f.Attributes.Mode&syscall.S_IFMT == syscall.S_IFLNK
}

// Index represents our optimized filesystem index
type Index struct {
	// Adaptive Radix Tree for fast lookup
	Trie         art.Tree
	withoutFiles map[string]struct{}
	withoutDirs  map[string]struct{}

	// IsComplete indicates if the index is complete (image is fully loaded or partially loaded)
	IsComplete bool
}

func (f *Node) ToProto() *fspb.FSIndexNode {
	return &fspb.FSIndexNode{
		Path:          f.Path,
		Attributes:    f.FileAttributesToProto(),
		LayerPosition: uint32(f.LayerPosition),
		SymlinkTarget: f.SymlinkTarget,
	}
}

func (f *Node) FileAttributesToProto() *fspb.FileAttributes {
	return &fspb.FileAttributes{
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
	}
}

func FSFileAttrFromProto(attr *fspb.FileAttributes) FileAttributes {
	return FileAttributes{
		Inode:     attr.Inode,
		Size:      attr.Size,
		Blocks:    attr.Blocks,
		Atime:     attr.Atime,
		Mtime:     attr.Mtime,
		Ctime:     attr.Ctime,
		Atimensec: attr.Atimensec,
		Mtimensec: attr.Mtimensec,
		Ctimensec: attr.Ctimensec,
		Mode:      attr.Mode,
		Nlink:     attr.Nlink,
		Rdev:      attr.Rdev,
		Blksize:   attr.Blksize,
	}
}

func FSNodeFromProto(node *fspb.FSIndexNode) *Node {
	return &Node{
		Path:          node.Path,
		Attributes:    FSFileAttrFromProto(node.Attributes),
		LayerPosition: uint8(node.LayerPosition),
		SymlinkTarget: node.SymlinkTarget,
	}
}
