package scanner

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/pkg/models"
)

type MediaScanner struct {
	database *db.Database
	watcher  *fsnotify.Watcher
}

func NewMediaScanner(database *db.Database) (*MediaScanner, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &MediaScanner{
		database: database,
		watcher:  watcher,
	}, nil
}

// ScanDirectory walks a directory and persists media files to the database.
// It does not block on ffprobe metadata; dimensions and duration are filled in
// asynchronously after the initial record is created so the UI can render quickly.
func (s *MediaScanner) ScanDirectory(dirPath string) error {
	fmt.Printf("[DEBUG] Scanning directory: %s\n", dirPath)
	var mu sync.Mutex
	var firstErr error
	var wg sync.WaitGroup

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[DEBUG] Error walking path %s: %v\n", path, err)
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Check if file is a supported media type
		if !s.isMediaFile(path) {
			fmt.Printf("[DEBUG] Skipping non-media file: %s\n", path)
			return nil
		}

		mediaFile := &models.MediaFile{
			Path:     path,
			Filename: info.Name(),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			FileType: s.getFileType(path),
			MimeType: s.getMimeType(path),
		}
		fmt.Printf("[DEBUG] Saving media file to DB: %s\n", path)
		err = s.database.CreateMediaFile(mediaFile)
		if err != nil {
			fmt.Printf("Error saving file %s: %v\n", path, err)
			return nil
		}

		// Backfill metadata asynchronously so ffprobe doesn't block the scan.
		wg.Add(1)
		go func(filePath string, mf *models.MediaFile) {
			defer wg.Done()
			width, height, duration, metaErr := preview.GetMetadata(filePath)
			if metaErr != nil {
				fmt.Printf("[DEBUG] Metadata extraction failed for %s: %v\n", filePath, metaErr)
				return
			}
			updates := map[string]interface{}{
				"width":    width,
				"height":   height,
				"duration": duration,
			}
			if updErr := s.database.UpdateMediaFileFields(filePath, updates); updErr != nil {
				fmt.Printf("[WARN] Failed to update metadata for %s: %v\n", filePath, updErr)
				mu.Lock()
				if firstErr == nil {
					firstErr = updErr
				}
				mu.Unlock()
			}
		}(path, mediaFile)

		return nil
	})

	wg.Wait()
	if err != nil {
		return err
	}
	return firstErr
}

func (s *MediaScanner) isMediaFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return preview.IsImageFile(ext) || preview.IsVideoFile(ext)
}

func (s *MediaScanner) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if preview.IsImageFile(ext) {
		return "image"
	}
	if preview.IsVideoFile(ext) {
		return "video"
	}
	return "unknown"
}

func (s *MediaScanner) getMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

func (s *MediaScanner) StartWatching(directories []string) error {
	for _, dir := range directories {
		err := s.watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("failed to watch directory %s: %w", dir, err)
		}
	}

	go s.watchLoop()
	return nil
}

func (s *MediaScanner) watchLoop() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// Handle new file creation
				if s.isMediaFile(event.Name) {
					s.handleNewFile(event.Name)
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				// Handle file deletion
				s.handleFileRemoval(event.Name)
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)
		}
	}
}

func (s *MediaScanner) handleNewFile(filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return
	}

	mediaFile := &models.MediaFile{
		Path:     filePath,
		Filename: info.Name(),
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		FileType: s.getFileType(filePath),
		MimeType: s.getMimeType(filePath),
	}

	err = s.database.CreateMediaFile(mediaFile)
	if err != nil {
		fmt.Printf("Error saving new file %s: %v\n", filePath, err)
	}
}

func (s *MediaScanner) handleFileRemoval(filePath string) {
	// TODO: Implement file removal from database
	fmt.Printf("File removed: %s\n", filePath)
}

func (s *MediaScanner) Close() error {
	return s.watcher.Close()
}
