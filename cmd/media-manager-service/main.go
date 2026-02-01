package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/service"
)

func main() {
	log.Println("[SERVICE] Starting Media Manager service...")

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("[SERVICE] Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("[SERVICE] Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create and start the media service
	mediaService, err := service.NewMediaService(cfg, database)
	if err != nil {
		log.Fatalf("[SERVICE] Failed to create media service: %v", err)
	}

	if err := mediaService.Start(); err != nil {
		log.Fatalf("[SERVICE] Failed to start media service: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("[SERVICE] Service is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("[SERVICE] Shutting down...")
	mediaService.Stop()
	log.Println("[SERVICE] Service stopped")
}
