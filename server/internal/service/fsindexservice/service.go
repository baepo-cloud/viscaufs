package fsindexservice

import (
	"fmt"
	"log"
	"log/slog"
	"slices"
	"time"

	"github.com/alphadose/haxmap"
	"github.com/baepo-cloud/viscaufs-server/internal/fsindex"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"gorm.io/gorm"
)

type Service struct {
	layerDigestToFSIndex *haxmap.Map[string, *fsindex.FSIndex]
	imageDigestToFSIndex *haxmap.Map[string, *fsindex.FSIndex]

	db     *gorm.DB
	logger *slog.Logger
}

// NewService creates a new fsindex service.
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:                   db,
		logger:               slog.New(slog.NewTextHandler(log.Writer(), nil)).With("service", "fsindex"),
		layerDigestToFSIndex: haxmap.New[string, *fsindex.FSIndex](),
		imageDigestToFSIndex: haxmap.New[string, *fsindex.FSIndex](),
	}
}

func (s *Service) CreateImageIndexChannel(imageDigest string) chan<- types.FileSystemIndexLayer {
	layersChan := make(chan types.FileSystemIndexLayer, 16)

	go func() {
		var imageFSIndex *fsindex.FSIndex
		firstTimeJoin := true
		now := time.Now()
		for layer := range layersChan {
			nowLayer := time.Now()
			var currentFsIndex *fsindex.FSIndex
			layerFSIndex, ok := s.layerDigestToFSIndex.Get(layer.Digest)
			if ok {
				currentFsIndex = layerFSIndex
			} else {
				currentFsIndex, _ = fsindex.Deserialize(layer.SerializedData)
			}

			if imageFSIndex == nil {
				imageFSIndex = currentFsIndex
			} else {
				fsindex.JoinFSIndex(currentFsIndex, imageFSIndex, layer.Position, firstTimeJoin)
				firstTimeJoin = false
				imageFSIndex = currentFsIndex
			}

			slog.Info("layer indexed",
				slog.String("image_digest", imageDigest),
				slog.Int("layer_position", int(layer.Position)),
				slog.Duration("duration", time.Since(nowLayer)),
				slog.String("layer_digest", layer.Digest))

			s.imageDigestToFSIndex.Set(imageDigest, imageFSIndex)
		}

		if imageFSIndex != nil {
			serializeFSIndex, err := imageFSIndex.Serialize()
			if err != nil {
				s.logger.Error("failed to serialize image fs index", slog.String("image_digest", imageDigest), slog.Any("error", err))
				return
			}
			err = s.db.Model(&types.Image{}).Where("digest = ?", imageDigest).Update("fs_index", serializeFSIndex).Error
			if err != nil {
				s.logger.Error("failed to update image fs index", slog.String("image_digest", imageDigest), slog.Any("error", err))
			}
			s.logger.Info("entire image indexed", slog.String("image_digest", imageDigest), slog.Duration("duration", time.Since(now)))
		}
	}()

	return layersChan
}

func (s *Service) Ready(imageDigest string) bool {
	_, ok := s.imageDigestToFSIndex.Get(imageDigest)
	if ok {
		return true
	}

	imageModel := &types.Image{}
	err := s.db.Model(&types.Image{}).Where("digest = ?", imageDigest).First(imageModel).Error
	if err != nil {
		return false
	}

	if imageModel.FsIndex != nil {
		deserializeFSIndex, _ := fsindex.Deserialize(imageModel.FsIndex)
		s.imageDigestToFSIndex.Set(imageDigest, deserializeFSIndex)
		return true
	}

	return false
}

func (s *Service) BuildImageIndex(img *types.Image, digestToPosition map[string]uint8) {
	if img.FsIndex != nil {
		fi, _ := fsindex.Deserialize(img.FsIndex)
		s.imageDigestToFSIndex.Set(img.Digest, fi)
		return
	}

	indexer := s.CreateImageIndexChannel(img.Digest)
	for _, layer := range slices.Backward(img.Layers) {
		s.logger.Info("image filesystem index create",
			slog.String("image_digest", img.Digest),
			slog.Int("layer_position", int(digestToPosition[layer.Digest])),
			slog.String("layer_digest", layer.Digest))

		indexer <- types.FileSystemIndexLayer{
			Digest:         layer.Digest,
			Position:       digestToPosition[layer.Digest],
			SerializedData: layer.FsIndex,
		}
	}
	close(indexer)
}

func (s *Service) BuildLayerIndex(path, layerDigest string) ([]byte, error) {
	index := fsindex.NewFSIndex()
	err := index.BuildIndex(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build file sytem index: %w", err)
	}

	serializedFileSystemIndex, err := index.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize file system index: %w", err)
	}
	s.layerDigestToFSIndex.Set(layerDigest, index)

	return serializedFileSystemIndex, nil
}

func (s *Service) Lookup(imageDigest, path string) *fsindex.FSNode {
	imageFSIndex, ok := s.imageDigestToFSIndex.Get(imageDigest)
	if !ok {
		return nil
	}

	node, err := imageFSIndex.LookupPath(path)
	if node == nil || err != nil {
		return nil
	}

	return node
}

func (s *Service) LookupByPrefix(imageDigest, path string) []*fsindex.FSNode {
	imageFSIndex, ok := s.imageDigestToFSIndex.Get(imageDigest)
	if !ok {
		return nil
	}

	nodes := imageFSIndex.LookupPrefixSearch(path)
	if nodes == nil {
		return nil
	}

	return nodes
}
