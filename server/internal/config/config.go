package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Addr                   string
	SqliteDir              string
	ImageDir               string
	ImageServiceNumWorkers int
}

func ParseConfig() *Config {
	defaultConfig := &Config{
		Addr:                   ":8080",
		SqliteDir:              "db",
		ImageDir:               "images",
		ImageServiceNumWorkers: 8,
	}

	addr := os.Getenv("ADDR")
	if addr != "" {
		defaultConfig.Addr = addr
	}

	sqliteDir := os.Getenv("SQLITE_DIR")
	if sqliteDir != "" {
		defaultConfig.SqliteDir = sqliteDir
	}

	imageDir := os.Getenv("IMAGE_DIR")
	if imageDir != "" {
		defaultConfig.ImageDir = imageDir
	}

	imageServiceNumWorkers := os.Getenv("IMAGE_SERVICE_NUM_WORKERS")
	if imageServiceNumWorkers != "" {
		numWorkers, err := strconv.Atoi(imageServiceNumWorkers)
		if err == nil {
			defaultConfig.ImageServiceNumWorkers = numWorkers
		}
	} else {
		defaultConfig.ImageServiceNumWorkers = 8
	}

	return defaultConfig
}
