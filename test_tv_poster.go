package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/service"
)

func main() {
	// Get database path
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".media-manager", "media.db")

	// Initialize database
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize TMDb service
	tmdb := service.NewTMDbService(database)

	// Test titles from the database
	testTitles := []string{
		"Shoresy S05E01 Keep It Simple 1080p AMZN WEB-DL DDP5 1 H 264-NTb",
		"Adventure Time Fionna and Cake S02E10 1080p WEB h264-DOLORES",
		"The Kardashians S07E10 1080p WEB h264-GRACE",
		"All Creatures Great and Small 2020 S06E00 Christmas Special 1080p HDTV H264-DARKFLiX",
		"Die My Love 2025", // Movie
	}

	fmt.Println("Testing TV Show and Movie Detection:")
	fmt.Println("=====================================\n")

	for _, title := range testTitles {
		fmt.Printf("Title: %s\n", title)

		// Extract media info
		info := tmdb.ExtractMediaInfo(title)
		fmt.Printf("  Name: %s\n", info.Name)
		fmt.Printf("  Year: %d\n", info.Year)
		fmt.Printf("  Type: ")
		if info.IsTV {
			fmt.Printf("TV Show (S%02dE%02d)\n", info.Season, info.Episode)
		} else {
			fmt.Printf("Movie\n")
		}

		// Fetch poster
		posterURL, tmdbID, err := tmdb.FetchPosterForTask(title)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  TMDb ID: %d\n", tmdbID)
			fmt.Printf("  Poster: %s\n", posterURL)
		}
		fmt.Println()
	}
}
