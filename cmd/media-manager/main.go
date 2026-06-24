package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/user/media-manager/internal/app"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
)

type launchTarget struct {
	path        string
	forceParent bool
}

type launchRequest struct {
	dir         string
	launchFiles []string
	explicitDir bool
}

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

func run(runner func(string, []string)) {
	request := resolveLaunchRequest()
	dir := request.dir
	explicitDir := request.explicitDir
	launchFiles := request.launchFiles

	if len(launchFiles) == 1 {
		log.Printf("Opening parent directory for file: %s -> %s", launchFiles[0], dir)
	} else if len(launchFiles) > 1 {
		log.Printf("Opening parent directory for %d selected files: %s", len(launchFiles), dir)
	}

	// Load config first to check for saved MediaDirs
	cfg, _ := config.LoadConfig("")

	// If no explicit directory was provided, prefer the saved selected folder.
	if !explicitDir && cfg != nil {
		if cfg.SelectedFolder != "" {
			if info, err := os.Stat(cfg.SelectedFolder); err == nil && info.IsDir() {
				dir = cfg.SelectedFolder
				log.Printf("Opening selected folder from config: %s", dir)
			}
		} else if len(cfg.MediaDirs) > 0 {
			lastDir := cfg.MediaDirs[len(cfg.MediaDirs)-1]
			if lastDir != "" {
				dir = lastDir
				log.Printf("Opening last directory from config: %s", dir)
			}
		}
	}

	if explicitDir || (cfg == nil || len(cfg.MediaDirs) == 0) {
		log.Printf("Opening directory: %s", dir)
	}

	// Reload config with the final directory
	cfg, _ = config.LoadConfig(dir)
	if cfg != nil {
		log.Printf("[DEBUG] main.go: Using DB path: %s", cfg.DatabasePath)
	} else {
		log.Printf("[DEBUG] main.go: Could not load config, DB path unknown.")
	}
	if runner != nil {
		runner(dir, launchFiles)
	}
}

func resolveLaunchRequest() launchRequest {
	explicitTargets := collectLaunchTargets(os.Args[1:])
	if len(explicitTargets) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
		return launchRequest{dir: cwd, explicitDir: false}
	}

	dirs := make([]string, 0, len(explicitTargets))
	launchFiles := make([]string, 0, len(explicitTargets))

	for _, target := range explicitTargets {
		info, err := os.Stat(target.path)
		if err != nil {
			if target.forceParent {
				dirs = append(dirs, filepath.Dir(target.path))
				continue
			}
			dirs = append(dirs, target.path)
			continue
		}

		if target.forceParent {
			dirs = append(dirs, filepath.Dir(target.path))
			continue
		}

		if info.IsDir() {
			dirs = append(dirs, target.path)
			continue
		}

		parentDir := filepath.Dir(target.path)
		dirs = append(dirs, parentDir)
		launchFiles = append(launchFiles, target.path)
	}

	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
		return launchRequest{dir: cwd, explicitDir: false}
	}

	selectedDir := dirs[0]
	for _, dir := range dirs[1:] {
		if !pathsEqual(selectedDir, dir) {
			log.Printf("[WARN] Multiple directories selected; using first: %s", selectedDir)
			break
		}
	}

	filteredFiles := make([]string, 0, len(launchFiles))
	for _, file := range launchFiles {
		if pathsEqual(filepath.Dir(file), selectedDir) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return launchRequest{dir: selectedDir, launchFiles: filteredFiles, explicitDir: true}
}

func collectLaunchTargets(args []string) []launchTarget {
	targets := make([]launchTarget, 0)
	forceParentNext := false

	for _, arg := range args {
		if arg == "--open-parent" {
			forceParentNext = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		targets = append(targets, launchTarget{path: arg, forceParent: forceParentNext})
		forceParentNext = false
	}

	return targets
}

func pathsEqual(a, b string) bool {
	if strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) {
		return true
	}
	return false
}

func runApp(dir string, launchFiles []string) {
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
	application, err := app.NewMediaManagerApp(dir, launchFiles)
	if err != nil {
		log.Fatalf("Failed to create application!: %v", err)
	}

	// Setup file watcher BEFORE showing window to avoid threading issues on Windows
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

	// Run the application on the main thread
	// This is a blocking call that returns when the window is closed
	application.Run()
}
