package ffmpeg

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

//go:embed binaries/ffmpeg-darwin-arm64
var ffmpegDarwinARM64 []byte

//go:embed binaries/ffprobe-darwin-arm64
var ffprobeDarwinARM64 []byte

var (
	ffmpegPath  string
	ffprobePath string
	once        sync.Once
	initErr     error
)

func Initialize() error {
	once.Do(func() {
		initErr = extractBinaries()
	})
	return initErr
}

func extractBinaries() error {
	tmpDir := os.TempDir()
	appDir := filepath.Join(tmpDir, "media-manager-ffmpeg")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	var ffmpegData, ffprobeData []byte
	var ffmpegName, ffprobeName string

	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			ffmpegData = ffmpegDarwinARM64
			ffprobeData = ffprobeDarwinARM64
			ffmpegName = "ffmpeg"
			ffprobeName = "ffprobe"
		case "amd64":
			return fmt.Errorf("macOS Intel binaries not yet embedded - please add them")
		default:
			return fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
	case "linux":
		return fmt.Errorf("Linux binaries not yet embedded - please add them")
	case "windows":
		return fmt.Errorf("Windows binaries not yet embedded - please add them")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	ffmpegPath = filepath.Join(appDir, ffmpegName)
	ffprobePath = filepath.Join(appDir, ffprobeName)

	if _, err := os.Stat(ffmpegPath); err == nil {
		fmt.Printf("[DEBUG] FFmpeg already extracted at: %s\n", ffmpegPath)
		return nil
	}

	if err := os.WriteFile(ffmpegPath, ffmpegData, 0755); err != nil {
		return fmt.Errorf("failed to write ffmpeg binary: %w", err)
	}

	if err := os.WriteFile(ffprobePath, ffprobeData, 0755); err != nil {
		return fmt.Errorf("failed to write ffprobe binary: %w", err)
	}

	fmt.Printf("[DEBUG] Extracted ffmpeg to: %s\n", ffmpegPath)
	fmt.Printf("[DEBUG] Extracted ffprobe to: %s\n", ffprobePath)

	return nil
}

func GetFFmpegPath() (string, error) {
	if err := Initialize(); err != nil {
		return "", err
	}
	return ffmpegPath, nil
}

func GetFFprobePath() (string, error) {
	if err := Initialize(); err != nil {
		return "", err
	}
	return ffprobePath, nil
}

func NewFFmpegCommand(args ...string) (*exec.Cmd, error) {
	path, err := GetFFmpegPath()
	if err != nil {
		return nil, err
	}
	return exec.Command(path, args...), nil
}

func NewFFprobeCommand(args ...string) (*exec.Cmd, error) {
	path, err := GetFFprobePath()
	if err != nil {
		return nil, err
	}
	return exec.Command(path, args...), nil
}
