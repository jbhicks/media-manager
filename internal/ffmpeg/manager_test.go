package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadAndCache(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	binDir := filepath.Join(homeDir, ".media-manager", "bin")
	os.RemoveAll(binDir)

	t.Log("Testing binary download on first initialization...")
	err = Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ffmpegPath, err := GetFFmpegPath()
	if err != nil {
		t.Fatalf("GetFFmpegPath failed: %v", err)
	}

	ffprobePath, err := GetFFprobePath()
	if err != nil {
		t.Fatalf("GetFFprobePath failed: %v", err)
	}

	if _, err := os.Stat(ffmpegPath); os.IsNotExist(err) {
		t.Errorf("ffmpeg binary not found at: %s", ffmpegPath)
	}

	if _, err := os.Stat(ffprobePath); os.IsNotExist(err) {
		t.Errorf("ffprobe binary not found at: %s", ffprobePath)
	}

	t.Logf("✓ ffmpeg: %s", ffmpegPath)
	t.Logf("✓ ffprobe: %s", ffprobePath)

	cmd, err := NewFFmpegCommand("-version")
	if err != nil {
		t.Fatalf("NewFFmpegCommand failed: %v", err)
	}

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run ffmpeg -version: %v", err)
	}

	t.Logf("FFmpeg version output:\n%s", string(output[:200]))
}
