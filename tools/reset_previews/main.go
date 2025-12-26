package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/user/media-manager/pkg/models"
	"gorm.io/gorm"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".media-manager", "media.db")
	fmt.Printf("Connecting to database: %s\n", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	fmt.Println("Resetting preview_path for all media files...")
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.MediaFile{}).Update("preview_path", "")
	if result.Error != nil {
		log.Fatalf("failed to reset preview paths: %v", result.Error)
	}

	fmt.Printf("Successfully reset %d preview paths.\n", result.RowsAffected)
}
