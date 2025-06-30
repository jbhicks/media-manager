package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	DatabasePath  string
	ThumbnailDir  string
	ThumbnailSize int
	MediaDirs     []string
}

func (c *Config) GetThumbnailDir() string {
	return c.ThumbnailDir
}

func (c *Config) GetThumbnailSize() int {
	return c.ThumbnailSize
}

func NewConfig(mediaDir string) *Config {
	homeDir, _ := os.UserHomeDir()
	fmt.Printf("[DEBUG] config.go: Received mediaDir: %s\n", mediaDir)
	configDir := filepath.Join(homeDir, ".media-manager")

	// Ensure config directory exists
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(filepath.Join(configDir, "thumbnails"), 0755)

	cfg := &Config{
		DatabasePath:  filepath.Join(configDir, "media.db"),
		ThumbnailDir:  filepath.Join(configDir, "thumbnails"),
		ThumbnailSize: 300,
		MediaDirs:     []string{mediaDir},
	}
	fmt.Printf("[DEBUG] config.go: Config.MediaDirs: %v\n", cfg.MediaDirs)

	// Load from environment variables
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}

	if thumbDir := os.Getenv("THUMBNAIL_DIR"); thumbDir != "" {
		cfg.ThumbnailDir = thumbDir
	}

	if thumbSize := os.Getenv("THUMBNAIL_SIZE"); thumbSize != "" {
		if size, err := strconv.Atoi(thumbSize); err == nil {
			cfg.ThumbnailSize = size
		}
	}

	return cfg
}
