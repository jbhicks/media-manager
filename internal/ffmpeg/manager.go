package ffmpeg

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	releaseVersion = "v0.1.0"
	githubRepo     = "jbhicks/media-manager"
)

var (
	ffmpegPath  string
	ffprobePath string
	once        sync.Once
	initErr     error
)

type binaryInfo struct {
	name     string
	url      string
	checksum string
}

func getBinaryInfo() (ffmpegInfo, ffprobeInfo binaryInfo, err error) {
	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", githubRepo, releaseVersion)

	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			ffmpegInfo = binaryInfo{
				name: "ffmpeg",
				url:  baseURL + "/ffmpeg-darwin-arm64",
			}
			ffprobeInfo = binaryInfo{
				name: "ffprobe",
				url:  baseURL + "/ffprobe-darwin-arm64",
			}
		case "amd64":
			err = fmt.Errorf("macOS Intel binaries not available yet")
		default:
			err = fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
	case "linux":
		err = fmt.Errorf("Linux binaries not available yet")
	case "windows":
		err = fmt.Errorf("Windows binaries not available yet")
	default:
		err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return
}

func Initialize() error {
	once.Do(func() {
		initErr = ensureBinaries()
	})
	return initErr
}

func ensureBinaries() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	binDir := filepath.Join(homeDir, ".media-manager", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	ffmpegInfo, ffprobeInfo, err := getBinaryInfo()
	if err != nil {
		return trySystemPath(err)
	}

	ffmpegPath = filepath.Join(binDir, ffmpegInfo.name)
	ffprobePath = filepath.Join(binDir, ffprobeInfo.name)

	if _, err := os.Stat(ffmpegPath); err == nil {
		if _, err := os.Stat(ffprobePath); err == nil {
			fmt.Printf("[INFO] Using cached ffmpeg binaries from: %s\n", binDir)
			return nil
		}
	}

	fmt.Printf("[INFO] Downloading ffmpeg binaries (one-time setup)...\n")

	if err := downloadBinary(ffmpegInfo.url, ffmpegPath); err != nil {
		return trySystemPath(fmt.Errorf("failed to download ffmpeg: %w", err))
	}

	if err := downloadBinary(ffprobeInfo.url, ffprobePath); err != nil {
		return trySystemPath(fmt.Errorf("failed to download ffprobe: %w", err))
	}

	fmt.Printf("[INFO] FFmpeg binaries downloaded successfully\n")
	return nil
}

func downloadBinary(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}

	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to make executable: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	return nil
}

func trySystemPath(downloadErr error) error {
	fmt.Printf("[WARN] %v\n", downloadErr)
	fmt.Printf("[INFO] Attempting to use system ffmpeg from PATH...\n")

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err == nil {
		ffprobePath, err = exec.LookPath("ffprobe")
		if err == nil {
			fmt.Printf("[INFO] Using system ffmpeg: %s\n", ffmpegPath)
			return nil
		}
	}

	return fmt.Errorf("ffmpeg not available: download failed and not found in PATH: %w", downloadErr)
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

func CalculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
