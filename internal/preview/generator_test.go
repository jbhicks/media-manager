package preview

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"runtime"
)

func getProjectRoot() (string, error) {
	_, b, _, _ := runtime.Caller(0)
	// dir is the directory of the current file (generator_test.go)
	dir := filepath.Dir(b)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir { // Reached root
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func TestGenerateVideoThumbnail(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tempDir := t.TempDir()
	videoPath := filepath.Join(tempDir, "test.mp4")
	thumbnailPath := filepath.Join(tempDir, "thumbnail.jpg")

	// Create a dummy video file
	if err := os.WriteFile(videoPath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("Failed to create dummy video file: %v", err)
	}

	fmt.Printf("[DEBUG] Testing thumbnail generation for: %s\n", videoPath)
	fmt.Printf("[DEBUG] Thumbnail output path: %s\n", thumbnailPath)

	err = GenerateThumbnail(videoPath, thumbnailPath)
	if err != nil {
		t.Fatalf("Failed to generate thumbnail for video: %v", err)
	}

	if _, err = os.Stat(thumbnailPath); os.IsNotExist(err) {
		t.Errorf("Thumbnail file not created: %s", thumbnailPath)
	}
}

func TestGenerateImageThumbnail(t *testing.T) {
	tempDir := t.TempDir()
	imagePath := filepath.Join(tempDir, "test.png")
	thumbnailPath := filepath.Join(tempDir, "thumbnail.jpg")

	// Create a dummy png file
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	f, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("Failed to create dummy image file: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Failed to encode dummy image: %v", err)
	}

	fmt.Printf("[DEBUG] Testing thumbnail generation for: %s\n", imagePath)
	fmt.Printf("[DEBUG] Thumbnail output path: %s\n", thumbnailPath)

	err = GenerateThumbnail(imagePath, thumbnailPath)
	if err != nil {
		t.Fatalf("Failed to generate thumbnail for image: %v", err)
	}

	if _, err = os.Stat(thumbnailPath); os.IsNotExist(err) {
		t.Errorf("Thumbnail file not created: %s", thumbnailPath)
	}
}

func TestGenerateThumbnailUnsupportedFile(t *testing.T) {
	tempDir := t.TempDir()
	unsupportedPath := filepath.Join(tempDir, "unsupported.txt")
	thumbnailPath := filepath.Join(tempDir, "unsupported.jpg")

	// Create a dummy unsupported file
	if err := os.WriteFile(unsupportedPath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("Failed to create dummy unsupported file: %v", err)
	}

	err := GenerateThumbnail(unsupportedPath, thumbnailPath)
	if err == nil {
		t.Errorf("Expected error for unsupported file type, got nil")
	}
}