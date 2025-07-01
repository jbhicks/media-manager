package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateVideoThumbnail(t *testing.T) {
	tempThumbDir := t.TempDir()
	videoPath := "media/big_buck_bunny.mp4"
	thumbnailName := "big_buck_bunny_thumbnail.jpg"
	thumbnailPath := filepath.Join(tempThumbDir, thumbnailName)

	fmt.Printf("[DEBUG] Testing thumbnail generation for: %s\n", videoPath)
	fmt.Printf("[DEBUG] Thumbnail output path: %s\n", thumbnailPath)

	err := GenerateThumbnail(videoPath, thumbnailPath)
	if err != nil {
		t.Fatalf("Failed to generate thumbnail for video: %v", err)
	}

	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		t.Errorf("Thumbnail file not created: %s", thumbnailPath)
	}
}

func TestGenerateImageThumbnail(t *testing.T) {
	tempThumbDir := t.TempDir()
	imagePath := "media/big_buck_bunny.jpg"
	thumbnailName := "big_buck_bunny_thumbnail.jpg"
	thumbnailPath := filepath.Join(tempThumbDir, thumbnailName)

	fmt.Printf("[DEBUG] Testing thumbnail generation for: %s\n", imagePath)
	fmt.Printf("[DEBUG] Thumbnail output path: %s\n", thumbnailPath)

	err := GenerateThumbnail(imagePath, thumbnailPath)
	if err != nil {
		t.Fatalf("Failed to generate thumbnail for image: %v", err)
	}

	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		t.Errorf("Thumbnail file not created: %s", thumbnailPath)
	}
}


func TestGenerateThumbnailUnsupportedFile(t *testing.T) {
	tempThumbDir := t.TempDir()
	unsupportedPath := "README.md"
	thumbnailPath := filepath.Join(tempThumbDir, "unsupported.jpg")

	err := GenerateThumbnail(unsupportedPath, thumbnailPath)
	if err == nil {
		t.Errorf("Expected error for unsupported file type, got nil")
	}
}
