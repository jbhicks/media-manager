package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// cleanupOrphanedPreviews removes preview GIFs that do not correspond to any current video file.
func cleanupOrphanedPreviews(mediaDirs []string) {
	homeDir, _ := os.UserHomeDir()
	previewDir := filepath.Join(homeDir, ".media-manager", "previews")
	gifFiles, err := os.ReadDir(previewDir)
	if err != nil {
		fmt.Printf("[WARN] Could not read preview directory: %v\n", err)
		return
	}

	// Build a set of all current video base names (without extension)
	videoBaseNames := make(map[string]struct{})
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v", ".3gp"}
	for _, dir := range mediaDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(file.Name()))
			for _, vext := range videoExts {
				if ext == vext {
					base := strings.TrimSuffix(file.Name(), ext)
					videoBaseNames[base] = struct{}{}
				}
			}
		}
	}

	// Remove any GIFs that do not match a video base name
	for _, gifFile := range gifFiles {
		if gifFile.IsDir() || !strings.HasSuffix(gifFile.Name(), ".gif") {
			continue
		}
		gifBase := strings.TrimSuffix(gifFile.Name(), ".gif")
		if _, ok := videoBaseNames[gifBase]; !ok {
			gifPath := filepath.Join(previewDir, gifFile.Name())
			fmt.Printf("[CLEANUP] Removing orphaned preview: %s\n", gifPath)
			os.Remove(gifPath)
		}
	}
}
