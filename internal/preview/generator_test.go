package preview

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
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
	var err error
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tempDir := t.TempDir()
	videoPath := filepath.Join(tempDir, "test.mp4")
	thumbnailPath := filepath.Join(tempDir, "thumbnail.jpg")

	// Create a minimal valid video file using ffmpeg
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "color=c=black:s=320x240:d=2", "-c:v", "libx264", "-t", "2", videoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test video file with ffmpeg: %v", err)
	}

	fmt.Printf("[DEBUG] Testing thumbnail generation for: %s\n", videoPath)
	fmt.Printf("[DEBUG] Thumbnail output path: %s\n", thumbnailPath)

	err = GenerateThumbnail(videoPath, thumbnailPath)
	if err != nil {
		t.Fatalf("GenerateThumbnail for video returned error: %v", err)
	}

	if _, err = os.Stat(thumbnailPath); err == nil {
		t.Errorf("Thumbnail file should NOT be created for video: %s", thumbnailPath)
	}
}

func TestGenerateImageThumbnail(t *testing.T) {
	var err error
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

func TestRemove1x1Frames(t *testing.T) {
	tempDir := t.TempDir()

	// Create a 1x1 JPEG frame
	frame1x1Path := filepath.Join(tempDir, "frame_1.jpg")
	img1x1 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	f1, err := os.Create(frame1x1Path)
	if err != nil {
		t.Fatalf("Failed to create 1x1 frame: %v", err)
	}
	if err := jpeg.Encode(f1, img1x1, nil); err != nil {
		f1.Close()
		t.Fatalf("Failed to encode 1x1 frame: %v", err)
	}
	f1.Close()

	// Create a 10x10 JPEG frame
	frame10x10Path := filepath.Join(tempDir, "frame_2.jpg")
	img10x10 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f2, err := os.Create(frame10x10Path)
	if err != nil {
		t.Fatalf("Failed to create 10x10 frame: %v", err)
	}
	if err := jpeg.Encode(f2, img10x10, nil); err != nil {
		f2.Close()
		t.Fatalf("Failed to encode 10x10 frame: %v", err)
	}
	f2.Close()

	framePaths := []string{frame1x1Path, frame10x10Path}
	filtered, err := remove1x1Frames(framePaths)
	if err != nil {
		t.Fatalf("remove1x1Frames returned error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 frame after removal, got %d", len(filtered))
	}
	if filtered[0] != frame10x10Path {
		t.Errorf("Expected remaining frame to be %s, got %s", frame10x10Path, filtered[0])
	}
	if _, err := os.Stat(frame1x1Path); !os.IsNotExist(err) {
		t.Errorf("1x1 frame was not deleted")
	}
}
