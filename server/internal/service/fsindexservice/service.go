package fsindexservice

import (
	"context"
	"fmt"
	"github.com/baepo-cloud/viscaufs-common/fsindex"
	"log"
	"log/slog"
	"slices"
	"time"

	"github.com/alphadose/haxmap"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"gorm.io/gorm"
)

type Service struct {
	layerDigestToFSIndex *haxmap.Map[string, *fsindex.Index]
	imageDigestToFSIndex *haxmap.Map[string, *fsindex.Index]

	db     *gorm.DB
	logger *slog.Logger
}

// NewService creates a new fsindex service.
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:                   db,
		logger:               slog.New(slog.NewTextHandler(log.Writer(), nil)).With("service", "fsindex"),
		layerDigestToFSIndex: haxmap.New[string, *fsindex.Index](),
		imageDigestToFSIndex: haxmap.New[string, *fsindex.Index](),
	}
}

func (s *Service) CreateImageIndexChannel(imageDigest string) chan<- types.FileSystemIndexLayer {
	layersChan := make(chan types.FileSystemIndexLayer, 16)

	go func() {
		var imageFSIndex *fsindex.Index
		firstTimeJoin := true
		now := time.Now()
		for layer := range layersChan {
			nowLayer := time.Now()
			var currentFsIndex *fsindex.Index
			layerFSIndex, ok := s.layerDigestToFSIndex.Get(layer.Digest)
			if ok {
				currentFsIndex = layerFSIndex
			} else {
				currentFsIndex, _ = fsindex.Deserialize(layer.SerializedData, false)
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

			if layer.Position == 0 {
				imageFSIndex.IsComplete = true
			}
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
		deserializeFSIndex, _ := fsindex.Deserialize(imageModel.FsIndex, true)
		s.imageDigestToFSIndex.Set(imageDigest, deserializeFSIndex)
		return true
	}

	return false
}

func (s *Service) BuildImageIndex(img *types.Image, digestToPosition map[string]uint8) {
	if img.FsIndex != nil {
		fi, _ := fsindex.Deserialize(img.FsIndex, true)
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

// Lookup attempts to lookup a path in the filesystem index
// and retries if the index is still being built, unless context is done
func (s *Service) Lookup(ctx context.Context, imageDigest, path string) *fsindex.Node {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Proceed with lookup
		}

		imageFSIndex, ok := s.imageDigestToFSIndex.Get(imageDigest)
		if !ok {
			if !s.Ready(imageDigest) {
				return nil
			}
			continue
		}

		node, err := imageFSIndex.LookupPath(path)
		if node != nil && err == nil {
			return node
		}

		if imageFSIndex.IsComplete {
			return nil
		}

		// Index is still building, wait a bit and retry
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// LookupByPrefix attempts to lookup paths with a prefix in the filesystem index
// and retries if the index is still being built, unless context is done
func (s *Service) LookupByPrefix(ctx context.Context, imageDigest, path string) []*fsindex.Node {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Proceed with lookup
		}

		imageFSIndex, ok := s.imageDigestToFSIndex.Get(imageDigest)
		if !ok {
			if !s.Ready(imageDigest) {
				return nil
			}
			continue
		}

		nodes := imageFSIndex.LookupPrefixSearch(path)
		if nodes != nil && len(nodes) > 0 {
			return nodes
		}

		fmt.Println("FINISHED?", imageFSIndex.IsComplete)
		if imageFSIndex.IsComplete {
			return nil
		}

		// Index is still building, wait a bit and retry
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(100 * time.Millisecond):
		}
	}
}
