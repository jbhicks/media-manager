package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
)

func deleteGifFiles(dir string) int {
	pattern := filepath.Join(dir, "*.gif")
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("[WARN] Failed to glob GIFs in %s: %v", dir, err)
		return 0
	}
	count := 0
	for _, file := range files {
		err := os.Remove(file)
		if err == nil {
			log.Printf("[INFO] Deleted animated preview: %s", file)
			count++
		} else {
			log.Printf("[WARN] Failed to delete %s: %v", file, err)
		}
	}
	return count
}

func main() {
	// Load config to get thumbnail directories
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Printf("[WARN] Could not load config, using default thumbnail dirs: %v", err)
	}

	// Default thumbnail dirs
	thumbDirs := []string{"./thumbnails"}
	if cfg != nil && cfg.ThumbnailDir != "" {
		thumbDirs = append([]string{cfg.ThumbnailDir}, thumbDirs...)
	}

	totalDeleted := 0
	for _, dir := range thumbDirs {
		deleted := deleteGifFiles(dir)
		totalDeleted += deleted
	}
	log.Printf("[INFO] Deleted %d animated preview GIFs from thumbnail directories.", totalDeleted)

	dbPath := "media-manager.db"
	if envPath := os.Getenv("MEDIA_MANAGER_DB"); envPath != "" {
		dbPath = envPath
	}
	log.Printf("[DEBUG] clear-previews: Using DB path: %s", dbPath)
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	err = database.ClearAllPreviews()
	if err != nil {
		log.Fatalf("Failed to clear previews: %v", err)
	}
	log.Println("Successfully cleared all preview paths in database.")
}
