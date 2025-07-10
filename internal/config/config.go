package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	DatabasePath           string
	ThumbnailDir           string
	ThumbnailSize          int
	MediaDirs              []string
	MainContentSplitOffset float32
	SidebarSplitOffset     float32
	WindowWidth            float32 // New field for window width
	WindowHeight           float32 // New field for window height
	WindowX                float32 // New field for window X position
	WindowY                float32 // New field for window Y position
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
		DatabasePath:           filepath.Join(configDir, "media.db"),
		ThumbnailDir:           filepath.Join(configDir, "thumbnails"),
		ThumbnailSize:          300,
		MediaDirs:              []string{mediaDir},
		MainContentSplitOffset: 0.25,
		SidebarSplitOffset:     0.95,
		WindowWidth:            0, // Initialize with 0, meaning no saved size
		WindowHeight:           0, // Initialize with 0, meaning no saved size
		WindowX:                0, // Initialize with 0, meaning no saved position
		WindowY:                0, // Initialize with 0, meaning no saved position
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

func GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".media-manager")
	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig(mediaDir string) (*Config, error) {
	fmt.Println("[DEBUG] Attempting to load config...")
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get config file path: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Config file path: %s\n", configFilePath)

	cfg := &Config{
		DatabasePath:           filepath.Join(filepath.Dir(configFilePath), "media.db"),
		ThumbnailDir:           filepath.Join(filepath.Dir(configFilePath), "thumbnails"),
		ThumbnailSize:          300,
		MediaDirs:              []string{mediaDir},
		MainContentSplitOffset: 0.25,
		SidebarSplitOffset:     0.95,
		WindowWidth:            0, // Initialize with 0, meaning no saved size
		WindowHeight:           0, // Initialize with 0, meaning no saved size
		WindowX:                0, // Initialize with 0, meaning no saved position
		WindowY:                0, // Initialize with 0, meaning no saved position
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			fmt.Printf("[DEBUG] config.json not found at %s, creating default config.\n", configFilePath)
			return cfg, nil
		}
		fmt.Printf("[DEBUG] Failed to read config file %s: %v\n", configFilePath, err)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	fmt.Printf("[DEBUG] Read %d bytes from config file.\n", len(data))
	err = json.Unmarshal(data, cfg)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to unmarshal config data: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	fmt.Printf("[DEBUG] Config loaded: %+v\n", cfg)
	// Ensure MediaDirs is set from the command line argument if provided
	if mediaDir != "" {
		cfg.MediaDirs = []string{mediaDir}
		fmt.Printf("[DEBUG] MediaDirs overridden by command line argument: %v\n", cfg.MediaDirs)
	}

	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	fmt.Println("[DEBUG] Attempting to save config...")
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get config file path: %v\n", err)
		return err
	}
	fmt.Printf("[DEBUG] Saving config to: %s\n", configFilePath)

	// Ensure config directory exists
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("[DEBUG] Failed to marshal config: %v\n", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to write config file: %v\n", err)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("[DEBUG] Config saved to %s\n", configFilePath)
	return nil
}
