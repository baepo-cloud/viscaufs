package types

import (
	"time"

	"github.com/baepo-cloud/viscaufs-server/internal/helper"
)

// Image represents the images table
type Image struct {
	ID           string
	Repository   string
	Identifier   string
	Digest       string
	LayersCount  int
	LayerDigests helper.SQLiteStringArray
	Manifest     string
	FsIndex      []byte

	CreatedAt time.Time

	Layers []*Layer `gorm:"many2many:image_layers;joinForeignKey:ImageID;joinReferences:LayerID"`
}

func (i Image) FindLayerByDigest(digest string) *Layer {
	for _, layer := range i.Layers {
		if layer.Digest == digest {
			return layer
		}
	}
	return nil
}

// Layer represents the layers table
type Layer struct {
	ID        string
	Digest    string
	FsIndex   []byte
	CreatedAt time.Time

	Images []Image `gorm:"many2many:image_layers;joinForeignKey:LayerID;joinReferences:ImageID"`
}

// ImageLayer represents the image_layers table (join table)
type ImageLayer struct {
	ImageID   string
	LayerID   string
	Position  int
	CreatedAt time.Time

	Image Image `gorm:"foreignKey:ImageID;references:ID"`
	Layer Layer `gorm:"foreignKey:LayerID;references:ID"`
}
