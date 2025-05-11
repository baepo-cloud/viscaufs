package filehandlerservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/baepo-cloud/viscaufs/common/fsindex"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alphadose/haxmap"
	"github.com/baepo-cloud/viscaufs-server/internal/config"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"github.com/nrednav/cuid2"
	"gorm.io/gorm"
)

// FileHandle represents information about an open file
type fileHandle struct {
	RelativePath string
	AbsolutePath string
	File         *os.File
	Flag         uint32
}

// Service handles container image operations
type Service struct {
	basePath        string
	db              *gorm.DB
	fsIndexService  types.FileSystemIndexService
	pendingFileOpen *haxmap.Map[string, fileHandle]
	logger          *slog.Logger
}

// NewService creates a new image service
func NewService(cfg *config.Config, db *gorm.DB, fsIndexSvc types.FileSystemIndexService) (*Service, error) {
	return &Service{
		basePath:        cfg.ImageDir,
		db:              db,
		fsIndexService:  fsIndexSvc,
		pendingFileOpen: haxmap.New[string, fileHandle](),
		logger:          slog.New(slog.NewTextHandler(log.Writer(), nil)).With("service", "file_handler"),
	}, nil
}

// OpenFile opens a file for reading
func (s *Service) OpenFile(ctx context.Context, params types.OpenFileParams) (string, error) {
	var (
		image types.Image
		uid   = cuid2.Generate()
		node  *fsindex.Node
	)

	node = s.fsIndexService.Lookup(ctx, params.ImageDigest, params.Path)
	if node == nil {
		return "", types.ErrFileNotFound
	}

	err := s.db.Where("digest = ?", params.ImageDigest).First(&image).Error
	if err != nil {
		return "", fmt.Errorf("failed to find image: %w", err)
	}

	layerDigest := image.LayerDigests[node.LayerPosition]

	// check if the file exists on the appropriate path
	filePath := filepath.Join(s.basePath, "layers", layerDigest, "content", params.Path)
	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("file not found: %w", err)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	//// Combine all write-related flags
	//writeFlags := os.O_WRONLY | os.O_RDWR | os.O_APPEND | os.O_CREATE | os.O_EXCL | os.O_TRUNC
	//
	//// Remove write flags because we only want to read
	//newFlag := int(params.Flags) &^ writeFlags

	// open the file
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	fh := fileHandle{
		RelativePath: params.Path,
		Flag:         params.Flags,
		AbsolutePath: filePath,
		File:         file,
	}

	s.pendingFileOpen.Set(uid, fh)

	return uid, nil
}

// ReleaseFile releases the file handle
func (s *Service) ReleaseFile(uid string) error {
	fh, ok := s.pendingFileOpen.Get(uid)
	if !ok {
		return fmt.Errorf("file handle not found: %s", uid)
	}

	if err := fh.File.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	s.pendingFileOpen.Del(uid)
	return nil
}

func (s *Service) ReadFile(uid string, offset int64, length uint32) ([]byte, error) {
	fh, ok := s.pendingFileOpen.Get(uid)
	if !ok {
		return nil, fmt.Errorf("file handle not found: %s", uid)
	}

	data := make([]byte, length)
	n, err := fh.File.ReadAt(data, offset)

	if (err == nil || err == io.EOF) && n > 0 {
		return data[:n], nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return []byte{}, nil
}
