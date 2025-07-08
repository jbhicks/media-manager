package scanner

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/user/media-manager/internal/db"
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

func (s *MediaScanner) ScanDirectory(dirPath string) error {
	fmt.Printf("[DEBUG] Scanning directory: %s\n", dirPath)
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
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

		// Create media file entry
		mediaFile := &models.MediaFile{
			Path:		 path,
			Filename: info.Name(),
			Size:		 info.Size(),
			ModTime:	 info.ModTime(),
			FileType: s.getFileType(path),
			MimeType: s.getMimeType(path),
		}

		// Save to database
		fmt.Printf("[DEBUG] Saving media file to DB: %s\n", path)
		err = s.database.CreateMediaFile(mediaFile)
		if err != nil {
			fmt.Printf("Error saving file %s: %v\n", path, err)
		}

		return nil
	})
}

func (s *MediaScanner) isMediaFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Supported image formats
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}

	// Supported video formats
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return true
		}
	}

	return false
}

func (s *MediaScanner) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "image"
		}
	}

	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return "video"
		}
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
