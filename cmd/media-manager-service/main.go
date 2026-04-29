package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/service"
)

func main() {
	log.Println("[SERVICE] Starting Media Manager Service...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("[SERVICE] No .env file found, using environment variables")
	} else {
		log.Println("[SERVICE] Loaded .env file")
	}

	// Get the media directory from args or use current directory
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Load configuration
	cfg, err := config.LoadConfig(dir)
	if err != nil {
		log.Fatalf("[SERVICE] Failed to load config: %v", err)
	}

	log.Printf("[SERVICE] Using database: %s", cfg.DatabasePath)

	// Initialize database
	database, err := db.NewDatabase(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("[SERVICE] Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create and start the service
	mediaService, err := service.NewMediaService(cfg, database)
	if err != nil {
		log.Fatalf("[SERVICE] Failed to create service: %v", err)
	}

	if err := mediaService.Start(); err != nil {
		log.Fatalf("[SERVICE] Failed to start service: %v", err)
	}

	log.Println("[SERVICE] Media Manager Service is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("[SERVICE] Received shutdown signal")

	mediaService.Stop()
	log.Println("[SERVICE] Service stopped")
}
