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
	"time"
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
	if err == nil {
		t.Errorf("Expected error for GenerateThumbnail with video path, got nil")
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

func TestSceneDetection(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tempDir := t.TempDir()
	videoPath := filepath.Join(tempDir, "test.mp4")

	// Create a test video with distinct scenes (color changes)
	// Scene 1: black (0-2s), Scene 2: white (2-4s), Scene 3: red (4-6s)
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "color=c=black:s=320x240:d=2",
		"-f", "lavfi", "-i", "color=c=white:s=320x240:d=2",
		"-f", "lavfi", "-i", "color=c=red:s=320x240:d=2",
		"-filter_complex", "[0:v][1:v][2:v]concat=n=3:v=1:a=0",
		"-c:v", "libx264", "-t", "6", videoPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test video with scenes: %v", err)
	}

	scenes, err := detectScenes(videoPath, 0.3)
	if err != nil {
		t.Fatalf("Scene detection failed: %v", err)
	}

	// Should detect at least 2 scene changes (black->white, white->red)
	if len(scenes) < 2 {
		t.Errorf("Expected at least 2 scenes, got %d", len(scenes))
	}

	fmt.Printf("[DEBUG] Detected scenes: %+v\n", scenes)
}

func TestSelectRepresentativeScenes(t *testing.T) {
	// Test with many scenes - should pick top N by score
	scenes := []SceneTimestamp{
		{Time: 1.0, Score: 0.5},
		{Time: 2.0, Score: 0.9},
		{Time: 3.0, Score: 0.3},
		{Time: 4.0, Score: 0.8},
		{Time: 5.0, Score: 0.6},
	}

	duration := 10 * time.Second
	selected := selectRepresentativeScenes(scenes, duration, 3)

	if len(selected) != 3 {
		t.Errorf("Expected 3 scenes, got %d", len(selected))
	}

	// Should select scenes with highest scores: 2.0 (0.9), 4.0 (0.8), 5.0 (0.6)
	// Sorted by time: 2.0, 4.0, 5.0
	expected := []float64{2.0, 4.0, 5.0}
	for i, exp := range expected {
		if selected[i] != exp {
			t.Errorf("Scene %d: expected %.1f, got %.1f", i, exp, selected[i])
		}
	}
}

func TestEvenlyDistributedTimestamps(t *testing.T) {
	duration := 100 * time.Second
	timestamps := evenlyDistributedTimestamps(duration, 4)

	if len(timestamps) != 4 {
		t.Fatalf("Expected 4 timestamps, got %d", len(timestamps))
	}

	// Should be at 20%, 40%, 60%, 80% (10%, 40%, 70%, 90% would also be valid)
	// Our implementation uses (i+1)/(count+1), so: 1/5=20%, 2/5=40%, 3/5=60%, 4/5=80%
	expected := []float64{20.0, 40.0, 60.0, 80.0}
	for i, exp := range expected {
		if timestamps[i] != exp {
			t.Errorf("Timestamp %d: expected %.1f, got %.1f", i, exp, timestamps[i])
		}
	}
}

func TestGenerateSmartPreview(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tempDir := t.TempDir()
	videoPath := filepath.Join(tempDir, "test.mp4")
	gifPath := filepath.Join(tempDir, "preview.gif")

	// Create test video
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "color=c=blue:s=320x240:d=5",
		"-c:v", "libx264", videoPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test video: %v", err)
	}

	opts := DefaultPreviewOptions()
	err := GenerateSmartPreview(videoPath, gifPath, opts)
	if err != nil {
		t.Fatalf("GenerateSmartPreview failed: %v", err)
	}

	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		t.Errorf("GIF preview not created: %s", gifPath)
	}
}

func TestGenerateSceneMosaic(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping test")
	}

	tempDir := t.TempDir()
	videoPath := filepath.Join(tempDir, "test.mp4")
	mosaicPath := filepath.Join(tempDir, "mosaic.jpg")

	// Create test video
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "color=c=green:s=320x240:d=5",
		"-c:v", "libx264", videoPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test video: %v", err)
	}

	opts := DefaultPreviewOptions()
	opts.UseMosaic = true

	err := GenerateSmartPreview(videoPath, mosaicPath, opts)
	if err != nil {
		t.Fatalf("GenerateSmartPreview (mosaic) failed: %v", err)
	}

	if _, err := os.Stat(mosaicPath); os.IsNotExist(err) {
		t.Errorf("Mosaic preview not created: %s", mosaicPath)
	}
}

func TestDefaultPreviewOptions(t *testing.T) {
	opts := DefaultPreviewOptions()

	if opts.SceneThreshold != 0.4 {
		t.Errorf("Expected SceneThreshold 0.4, got %.2f", opts.SceneThreshold)
	}
	if opts.MaxScenes != 4 {
		t.Errorf("Expected MaxScenes 4, got %d", opts.MaxScenes)
	}
	if opts.FPS != 12 {
		t.Errorf("Expected FPS 12, got %d", opts.FPS)
	}
	if opts.Width != 216 || opts.Height != 216 {
		t.Errorf("Expected size 216x216, got %dx%d", opts.Width, opts.Height)
	}
	if opts.UseGPU {
		t.Error("Expected UseGPU to be false by default")
	}
}
