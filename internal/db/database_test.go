package db

import (
	"github.com/user/media-manager/pkg/models"
	"os"
	"testing"
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
