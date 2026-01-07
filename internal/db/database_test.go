package db

import (
	"os"
	"testing"

	"github.com/user/media-manager/pkg/models"
)

func TestClearAllPreviews(t *testing.T) {
	dbPath := "test_clear_previews.db"
	defer os.Remove(dbPath)
	database, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Insert test records
	files := []models.MediaFile{
		{Path: "file1.mp4", PreviewPath: "preview1.gif"},
		{Path: "file2.mp4", PreviewPath: "preview2.gif"},
		{Path: "file3.mp4", PreviewPath: ""}, // already empty
	}
	for _, f := range files {
		err := database.GetDB().Create(&f).Error
		if err != nil {
			t.Fatalf("Failed to insert test record: %v", err)
		}
	}

	// Clear previews
	err = database.ClearAllPreviews()
	if err != nil {
		t.Fatalf("ClearAllPreviews failed: %v", err)
	}

	// Check all preview paths are empty
	var results []models.MediaFile
	err = database.GetDB().Find(&results).Error
	if err != nil {
		t.Fatalf("Failed to query records: %v", err)
	}
	for _, f := range results {
		if f.PreviewPath != "" {
			t.Errorf("PreviewPath not cleared for %s: got '%s'", f.Path, f.PreviewPath)
		}
	}
}

func TestGetMediaFileByPath(t *testing.T) {
	dbPath := "test_get_media_file.db"
	defer os.Remove(dbPath)
	database, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Create test tags
	tag1 := models.Tag{Name: "studio:Test Studio", Color: "#FF0000"}
	tag2 := models.Tag{Name: "actress:Test Actress", Color: "#00FF00"}
	err = database.GetDB().Create(&tag1).Error
	if err != nil {
		t.Fatalf("Failed to create test tag1: %v", err)
	}
	err = database.GetDB().Create(&tag2).Error
	if err != nil {
		t.Fatalf("Failed to create test tag2: %v", err)
	}

	// Create test media file with tags
	mediaFile := models.MediaFile{
		Path:          "test/path/video.mp4",
		Filename:      "video.mp4",
		FriendlyTitle: "Test Studio - Test Video - 2024-01-01",
		Tags:          []models.Tag{tag1, tag2},
	}
	err = database.GetDB().Create(&mediaFile).Error
	if err != nil {
		t.Fatalf("Failed to create test media file: %v", err)
	}

	// Test GetMediaFileByPath
	retrieved, err := database.GetMediaFileByPath("test/path/video.mp4")
	if err != nil {
		t.Fatalf("GetMediaFileByPath failed: %v", err)
	}

	// Verify basic fields
	if retrieved.Path != "test/path/video.mp4" {
		t.Errorf("Path mismatch: got %s, expected %s", retrieved.Path, "test/path/video.mp4")
	}
	if retrieved.Filename != "video.mp4" {
		t.Errorf("Filename mismatch: got %s, expected %s", retrieved.Filename, "video.mp4")
	}
	if retrieved.FriendlyTitle != "Test Studio - Test Video - 2024-01-01" {
		t.Errorf("FriendlyTitle mismatch: got %s, expected %s", retrieved.FriendlyTitle, "Test Studio - Test Video - 2024-01-01")
	}

	// Verify tags are loaded (this was the bug!)
	if len(retrieved.Tags) != 2 {
		t.Errorf("Tags not loaded: got %d tags, expected 2", len(retrieved.Tags))
	}

	// Check tag names
	tagNames := make(map[string]bool)
	for _, tag := range retrieved.Tags {
		tagNames[tag.Name] = true
	}
	if !tagNames["studio:Test Studio"] {
		t.Error("Missing studio tag")
	}
	if !tagNames["actress:Test Actress"] {
		t.Error("Missing actress tag")
	}
}

func TestGetMediaFileByPathNotFound(t *testing.T) {
	dbPath := "test_get_media_file_not_found.db"
	defer os.Remove(dbPath)
	database, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	// Test with non-existent path
	_, err = database.GetMediaFileByPath("nonexistent/path.mp4")
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}
