package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/user/media-manager/internal/app"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
)

func main() {
	// Check for dev-reset flag
	resetAll := false
	for _, arg := range os.Args[1:] {
		if arg == "--dev-reset" || arg == "--reset-all" {
			resetAll = true
			break
		}
	}
	if resetAll {
		log.Printf("[DEV] --dev-reset flag detected: clearing cache and database before startup.")
		clearCacheAndDb()
	}
	run(runApp)
}

// clearCacheAndDb deletes the thumbnail cache and database file for a full reset
func clearCacheAndDb() {
	cfgPath := os.ExpandEnv("$HOME/.media-manager/config.json")
	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		log.Printf("[DEV] Could not open config file for reset: %v", err)
		return
	}
	defer cfgFile.Close()
	var cfg struct {
		DatabasePath string
		ThumbnailDir string
	}
	if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		log.Printf("[DEV] Could not decode config for reset: %v", err)
		return
	}
	if cfg.ThumbnailDir != "" {
		log.Printf("[DEV] Removing thumbnail cache: %s", cfg.ThumbnailDir)
		os.RemoveAll(cfg.ThumbnailDir)
	}
	if cfg.DatabasePath != "" {
		log.Printf("[DEV] Removing database: %s", cfg.DatabasePath)
		os.Remove(cfg.DatabasePath)
	}
}

func run(runner func(string)) {
	dir := getDirectoryFromArgs()
	log.Printf("Opening directory: %s", dir)
	cfg, _ := config.LoadConfig(dir)
	if cfg != nil {
		log.Printf("[DEBUG] main.go: Using DB path: %s", cfg.DatabasePath)
	} else {
		log.Printf("[DEBUG] main.go: Could not load config, DB path unknown.")
	}
	if runner != nil {
		runner(dir)
	}
}

func getDirectoryFromArgs() string {
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if !strings.HasPrefix(arg, "-") {
				return arg
			}
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	return cwd
}

func runApp(dir string) {
	if os.Getenv("CLEAR_DB_ON_START") == "true" {
		// Load config to get the correct database path
		cfg, err := config.LoadConfig(dir)
		if err != nil {
			log.Fatalf("Failed to load config for clearing previews: %v", err)
		}
		database, err := db.NewDatabase(cfg.DatabasePath)
		if err != nil {
			log.Fatalf("Failed to open database for clearing previews: %v", err)
		}
		if err := database.ClearAllPreviews(); err != nil {
			log.Fatalf("Failed to clear previews: %v", err)
		}
		if err := database.Close(); err != nil {
			log.Printf("Warning: failed to close database: %v", err)
		}
	}
	log.Printf("[DEBUG] main.go: Passing dir to app: %s", dir)
	application, err := app.NewMediaManagerApp(dir)
	if err != nil {
		log.Fatalf("Failed to create application!: %v", err)
	}
	application.Run()

	// Setup file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Watcher events channel closed.")
					return
				}
				log.Printf("FSNotify Event: Name=%s, Op=%s", event.Name, event.Op)
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
					log.Printf("Detected relevant file system change for %s, triggering rescan.", event.Name)
					application.RescanMediaDirectory()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Keep main goroutine alive
	<-make(chan struct{})
}
