package fsindex

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	art "github.com/plar/go-adaptive-radix-tree/v2"
	protobuf "google.golang.org/protobuf/proto"
)

const (
	currentVersion = uint32(1)
)

// Serialize serializes the FSIndex into a FlatBuffer byte array
func (idx *FSIndex) Serialize() ([]byte, error) {
	proto := &fspb.FSIndex{
		Version: currentVersion,
		Paths:   make([]*fspb.FSNode, 0),
	}

	idx.trie.ForEach(func(node art.Node) bool {
		fsNode, ok := node.Value().(*FSNode)
		if !ok {
			return true
		}

		proto.Paths = append(proto.Paths, fsNode.ToProto())
		return true
	})

	data, err := protobuf.Marshal(proto)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize FSIndex: %w", err)
	}

	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Deserialize deserializes the byte array into a FSIndex
func Deserialize(data []byte) (*FSIndex, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer r.Close()

	decompressedData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	proto := &fspb.FSIndex{}
	if err := protobuf.Unmarshal(decompressedData, proto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal FSIndex: %w", err)
	}

	if proto.Version != currentVersion {
		return nil, fmt.Errorf("unsupported FSIndex version: %d", proto.Version)
	}

	idx := &FSIndex{
		trie: art.New(),
	}

	for _, nodeProto := range proto.Paths {
		node := FSNodeFromProto(nodeProto)
		idx.trie.Insert(art.Key(node.Path), node)
	}

	return idx, nil
}
