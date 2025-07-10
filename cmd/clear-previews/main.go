package main

import (
	"log"
	"os"

	"github.com/user/media-manager/internal/db"
)

func main() {
	dbPath := "media-manager.db"
	if envPath := os.Getenv("MEDIA_MANAGER_DB"); envPath != "" {
		dbPath = envPath
	}
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	err = database.ClearAllPreviews()
	if err != nil {
		log.Fatalf("Failed to clear previews: %v", err)
	}
	log.Println("Successfully cleared all previews.")
}
