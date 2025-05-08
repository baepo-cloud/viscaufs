package imgservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/alphadose/haxmap"
	"github.com/baepo-cloud/viscaufs-server/internal/config"
	"github.com/baepo-cloud/viscaufs-server/internal/helper"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"github.com/google/go-containerregistry/pkg/name"
	img "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/nrednav/cuid2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service handles container image operations
type Service struct {
	basePath        string
	db              *gorm.DB
	fsIndexService  types.FileSystemIndexService
	pendingDownload *haxmap.Map[string, struct{}] // set of image digest
	logger          *slog.Logger
}

var _ types.ImageService = (*Service)(nil)

// NewService creates a new image service
func NewService(cfg *config.Config, db *gorm.DB, fsIndexSvc types.FileSystemIndexService) (*Service, error) {
	if err := os.MkdirAll(filepath.Join(cfg.ImageDir, "layers"), 0755); err != nil {
		return nil, err
	}

	return &Service{
		basePath:        cfg.ImageDir,
		db:              db,
		fsIndexService:  fsIndexSvc,
		pendingDownload: haxmap.New[string, struct{}](),
		logger:          slog.New(slog.NewTextHandler(log.Writer(), nil)).With("service", "image"),
	}, nil
}

type ImageWrapper struct {
	Reference      name.Reference
	Image          img.Image
	Manifest       *img.Manifest
	Digest         string
	Layers         []img.Layer
	LayersDigests  []string
	ExistingLayers []types.Layer
	ImageModel     *types.Image
}

// Download downloads an image and its layers
func (s *Service) Download(refId string) (string, error) {
	image, err := s.buildImageWrapper(refId)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve image: %w", err)
	}

	logger := s.logger.With(
		"reference", image.Reference.String(),
		"digest", image.Digest,
		"component", "downloading")

	// Check if the image already findImageByDigestID
	imageModel, err := s.findImageByDigestID(image.Digest)
	if err != nil {
		return "", fmt.Errorf("failed to check image existence: %w", err)
	}

	if imageModel != nil && len(image.Layers) == len(image.LayersDigests) {
		logger.Info("image already exists in local storage")
		s.fsIndexService.BuildImageIndex(imageModel, s.createDigestToPositionMap(image.LayersDigests))
		return imageModel.Digest, types.ErrImageAlreadyPresent
	}

	if imageModel == nil {
		imageModel, err = s.upsertImage(image)
		if err != nil {
			return "", fmt.Errorf("failed to upsert image: %w", err)
		}
	}

	image.ImageModel = imageModel

	err = s.tryAcquireDownload(image, logger)
	if err != nil {
		return "", err
	}

	// todo: this is dangerous
	go func() {
		s.downloadLayersReverseOrder(image, imageModel, logger)
	}()

	return imageModel.Digest, nil
}

func (s *Service) tryAcquireDownload(image *ImageWrapper, logger *slog.Logger) error {
	_, ok := s.pendingDownload.Get(image.Digest)
	if ok {
		logger.Info("image already downloading")
		return types.ErrImageDownloadAlreadyAcquired
	}

	s.pendingDownload.Set(image.Digest, struct{}{})
	return nil
}

func (s *Service) downloadLayersReverseOrder(imgWrapper *ImageWrapper, imageModel *types.Image, logger *slog.Logger) {
	filesystemIndexer := s.fsIndexService.CreateImageIndexChannel(imgWrapper.Digest)

	for _, items := range helper.BackwardByBatches(imgWrapper.Layers, 1) {
		var (
			wg                        sync.WaitGroup
			serializedFSIndexByDigest = haxmap.New[string, []byte]()
		)

		for _, layer := range items {
			wg.Add(1)
			go func(layerIndex int) {
				defer wg.Done()
				layerDigest := imgWrapper.LayersDigests[layerIndex]

				layerModel, err := s.downloadLayer(filepath.Join(s.basePath, "layers"), layerIndex, imgWrapper, imageModel)
				if err != nil {
					logger.Error("failed to download layer",
						slog.String("layer_digest", layerDigest),
						slog.String("error", err.Error()))
				}

				serializedFSIndexByDigest.Set(layerDigest, layerModel.FsIndex)
			}(layer.Index)
		}

		wg.Wait()

		for _, layer := range items {
			position := uint8(layer.Index)
			digest := imgWrapper.LayersDigests[layer.Index]
			serializedFSIndex, _ := serializedFSIndexByDigest.Get(digest)

			filesystemIndexer <- types.FileSystemIndexLayer{
				Digest:         digest,
				Position:       position,
				SerializedData: serializedFSIndex,
			}
		}
	}

	logger.Info("image download completed")
	close(filesystemIndexer)
	s.pendingDownload.Del(imgWrapper.Digest)
}

// downloadLayer downloads and extracts a single layer
func (s *Service) downloadLayer(basePath string, position int, imgWrapper *ImageWrapper, model *types.Image) (*types.Layer, error) {
	digest := imgWrapper.LayersDigests[position]
	layer := imgWrapper.Layers[position]

	layerPath := filepath.Join(basePath, digest)
	contentPath := filepath.Join(layerPath, "content")
	if err := os.MkdirAll(contentPath, 0755); err != nil {
		return nil, err
	}

	layerModel := imgWrapper.ImageModel.FindLayerByDigest(digest)
	if layerModel != nil {
		slog.Info("layer already findImageByDigestID and is valid", slog.String("digest", digest))
		return layerModel, nil
	}

	// Download the layer as tar file
	tarFilePath := filepath.Join(contentPath, cuid2.Generate()+".tar")
	r, err := layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	file, err := os.Create(tarFilePath)
	if err != nil {
		return nil, err
	}

	// Copy contents to file
	if _, err = io.Copy(file, r); err != nil {
		file.Close()
		return nil, err
	}

	// Close the file before extraction
	file.Close()

	// Extract the tar file using the tar command
	cmd := exec.Command("tar", "-xf", tarFilePath, "-C", contentPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract layer: %w", err)
	}

	slog.Info("layer extracted", slog.String("layer_digest", digest))

	// Delete the tar file after extraction
	if err := os.Remove(tarFilePath); err != nil {
		return nil, fmt.Errorf("failed to remove tar file: %w", err)
	}

	// build the layer file system index
	serializedFSIndex, err := s.fsIndexService.BuildLayerIndex(contentPath, digest)

	// Upsert the layer into the database
	layerModel, err = s.upsertLayer(*model, layer, serializedFSIndex, position)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert layer: %w", err)
	}

	return layerModel, nil
}

// findImageByDigestID checks if the image findImageByDigestID in the local storage and if all layers are present and valid
func (s *Service) findImageByDigestID(digestID string) (*types.Image, error) {
	var image types.Image
	err := s.db.
		Preload("Layers").
		Where("digest = ?", digestID).First(&image).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check image existence: %w", err)
	}

	return &image, nil
}

func (s *Service) upsertImage(imgWrapper *ImageWrapper) (*types.Image, error) {
	manifestBytes, err := json.Marshal(imgWrapper.Manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	imageModel := types.Image{
		ID:           cuid2.Generate(),
		Repository:   imgWrapper.Reference.Context().Name(),
		Identifier:   imgWrapper.Reference.Identifier(),
		LayersCount:  len(imgWrapper.Layers),
		LayerDigests: imgWrapper.LayersDigests,
		Manifest:     string(manifestBytes),
		Digest:       imgWrapper.Digest,
	}

	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "digest"}},
		DoNothing: true,
	}).Create(&imageModel).Error; err != nil {
		return nil, fmt.Errorf("failed to upsert image: %w", err)
	}

	return &imageModel, nil
}

func (s *Service) upsertLayer(i types.Image, layer img.Layer, fsIndex []byte, position int) (*types.Layer, error) {
	digest, err := layer.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get layer digest: %w", err)
	}

	layerModel := types.Layer{
		ID:      cuid2.Generate(),
		Digest:  digest.String(),
		FsIndex: fsIndex,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "digest"}},
			DoNothing: true,
		}).Create(&layerModel).Error; err != nil {
			return fmt.Errorf("failed to upsert layer: %w", err)
		}

		err = s.db.Create(&types.ImageLayer{
			ImageID:  i.ID,
			LayerID:  layerModel.ID,
			Position: position,
		}).Error

		if err != nil {
			return fmt.Errorf("failed to upsert layer image relation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to associate layer with image: %w", err)
	}

	return &layerModel, nil
}

func (s *Service) buildImageWrapper(id string) (*ImageWrapper, error) {
	reference, err := name.ParseReference(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	image, err := remote.Image(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %w", err)
	}

	manifest, err := image.Manifest()
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}

	digest, err := image.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get image digest: %w", err)
	}

	layersDigests := make([]string, len(layers))
	for i, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer digest: %w", err)
		}
		layersDigests[i] = digest.String()
	}

	return &ImageWrapper{
		Reference:     reference,
		Image:         image,
		Layers:        layers,
		LayersDigests: layersDigests,
		Manifest:      manifest,
		Digest:        digest.String(),
	}, nil
}

func (s *Service) createDigestToPositionMap(layers []string) map[string]uint8 {
	digestToPosition := make(map[string]uint8)
	for i, digest := range layers {
		digestToPosition[digest] = uint8(i)
	}
	return digestToPosition
}
