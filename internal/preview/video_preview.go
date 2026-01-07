package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user/media-manager/internal/ffmpeg"
)

// VideoPreviewConfig configures video preview generation
type VideoPreviewConfig struct {
	// Output format: "webm", "gif", or "jpg" (static thumbnail)
	Format string
	// Width of output (height auto-calculated to maintain aspect ratio)
	Width int
	// FPS for animated previews
	FPS int
	// Duration of each clip segment in seconds
	ClipDuration float64
	// Number of clips to extract
	NumClips int
	// Use GPU acceleration if available
	UseGPU bool
	// CRF quality (lower = better quality, larger file). 0-63 for VP9
	CRF int
}

// DefaultVideoPreviewConfig returns optimized defaults
func DefaultVideoPreviewConfig() VideoPreviewConfig {
	return VideoPreviewConfig{
		Format:       "webm", // WebM is 11x smaller than GIF
		Width:        320,
		FPS:          10,
		ClipDuration: 0.5,
		NumClips:     4,
		UseGPU:       false, // TEMPORARILY DISABLED: GPU processing causes mouse lag
		CRF:          35,   // Good balance of quality and size
	}
}

// GPU detection cache
var (
	gpuOnce       sync.Once
	detectedGPU   string
	gpuAvailable  bool
)

// DetectGPU checks for available GPU acceleration and caches the result
func DetectGPU() (gpuType string, available bool) {
	gpuOnce.Do(func() {
		fmt.Println("[DEBUG] Detecting GPU acceleration support...")
		
		hwaccels, err := GetFFmpegHardwareAccelerations()
		if err != nil {
			fmt.Printf("[WARN] Failed to detect hardware accelerations: %v\n", err)
			return
		}

		fmt.Printf("[DEBUG] Available hardware accelerations: %v\n", hwaccels)

		// Priority order: CUDA (NVIDIA) > VAAPI (Intel/AMD Linux) > D3D11VA (Windows) > VideoToolbox (macOS)
		priorities := []string{"cuda", "vaapi", "d3d11va", "videotoolbox", "qsv"}
		
		for _, priority := range priorities {
			for _, hw := range hwaccels {
				if strings.ToLower(hw) == priority {
					detectedGPU = priority
					gpuAvailable = true
					fmt.Printf("[INFO] GPU acceleration available: %s\n", detectedGPU)
					return
				}
			}
		}

		fmt.Println("[INFO] No GPU acceleration available, using CPU")
	})

	return detectedGPU, gpuAvailable
}

// GenerateVideoThumbnail creates a static JPG thumbnail mosaic from a video
// Shows 4 frames (2x2 grid) from different parts of the video
func GenerateVideoThumbnail(srcPath, destPath string, width int) error {
	fmt.Printf("[DEBUG] Generating static video thumbnail mosaic: %s\n", srcPath)

	// Check if already exists
	if _, err := os.Stat(destPath); err == nil {
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		// Fallback to 10 seconds if we can't get duration
		duration = 10 * time.Second
	}
	
	durSec := duration.Seconds()
	
	// Calculate 4 timestamps at 10%, 35%, 65%, 90% of the video
	t1 := durSec * 0.10
	t2 := durSec * 0.35
	t3 := durSec * 0.65
	t4 := durSec * 0.90
	
	// Minimum timestamps
	if t1 < 0.5 { t1 = 0.5 }
	if t2 < 1.0 { t2 = 1.0 }
	if t3 < 1.5 { t3 = 1.5 }
	if t4 < 2.0 { t4 = 2.0 }
	
	// Each cell is half the total width, maintaining 16:9 aspect ratio
	cellW := width / 2
	cellH := cellW * 9 / 16
	
	// Use select filter with timestamps to get frames at specific times
	// This requires a more complex approach - use multiple seeks
	err = generateMosaicThumbnail(srcPath, destPath, []float64{t1, t2, t3, t4}, cellW, cellH)
	if err != nil {
		// Fallback to single frame if mosaic fails
		fmt.Printf("[WARN] Mosaic thumbnail failed, falling back to single frame: %v\n", err)
		return generateSingleFrameThumbnail(srcPath, destPath, width, durSec*0.25)
	}
	
	return nil
}

// generateMosaicThumbnail creates a 2x2 grid from 4 frames at specified timestamps
func generateMosaicThumbnail(srcPath, destPath string, timestamps []float64, cellW, cellH int) error {
	fmt.Printf("[DEBUG] Generating 2x2 mosaic thumbnail at times: %v\n", timestamps)
	
	// Create temp directory for individual frames
	tempDir, err := os.MkdirTemp("", "mosaic_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Extract 4 frames with fast seeking
	framePaths := make([]string, 4)
	for i, ts := range timestamps {
		framePath := filepath.Join(tempDir, fmt.Sprintf("frame_%d.jpg", i))
		framePaths[i] = framePath
		
		cmd, err := ffmpeg.NewFFmpegCommand(
			"-loglevel", "warning",
			"-ss", fmt.Sprintf("%.2f", ts),
			"-i", srcPath,
			"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", cellW, cellH, cellW, cellH),
			"-frames:v", "1",
			"-q:v", "2",
			"-y", framePath,
		)
		if err != nil {
			return fmt.Errorf("failed to create ffmpeg command for frame %d: %w", i, err)
		}
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("frame extraction %d failed: %w\nOutput: %s", i, err, string(output))
		}
	}
	
	// Now combine the 4 frames into a 2x2 grid using xstack
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", framePaths[0],
		"-i", framePaths[1],
		"-i", framePaths[2],
		"-i", framePaths[3],
		"-filter_complex", "[0:v][1:v][2:v][3:v]xstack=inputs=4:layout=0_0|w0_0|0_h0|w0_h0",
		"-frames:v", "1",
		"-q:v", "2",
		"-y", destPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg command for mosaic: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mosaic thumbnail generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Generated mosaic thumbnail: %s\n", destPath)
	return nil
}

// generateSingleFrameThumbnail creates a single frame thumbnail (fallback)
func generateSingleFrameThumbnail(srcPath, destPath string, width int, seekTime float64) error {
	if seekTime < 1 {
		seekTime = 1
	}

	filter := fmt.Sprintf("scale=%d:-1", width)
	
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-ss", fmt.Sprintf("%.2f", seekTime),
		"-i", srcPath,
		"-vf", filter,
		"-frames:v", "1",
		"-q:v", "2",
		"-y", destPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("single frame thumbnail failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Generated single frame thumbnail: %s\n", destPath)
	return nil
}

// generateThumbnailWithGPU uses GPU acceleration for thumbnail
func generateThumbnailWithGPU(srcPath, destPath string, width int, seekTime float64, gpuType string) error {
	var cmdArgs []string
	
	switch gpuType {
	case "cuda":
		// NVIDIA CUDA acceleration
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", "cuda",
			"-hwaccel_output_format", "cuda",
			"-ss", fmt.Sprintf("%.2f", seekTime),
			"-i", srcPath,
			"-vf", fmt.Sprintf("scale_cuda=%d:-1,hwdownload,format=nv12", width),
			"-frames:v", "1",
			"-q:v", "2",
			"-y", destPath,
		}
	case "vaapi":
		// Intel/AMD VAAPI acceleration
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", "vaapi",
			"-ss", fmt.Sprintf("%.2f", seekTime),
			"-i", srcPath,
			"-vf", fmt.Sprintf("format=nv12,hwupload,scale_vaapi=w=%d:h=-1,hwdownload", width),
			"-frames:v", "1",
			"-q:v", "2",
			"-y", destPath,
		}
	case "d3d11va":
		// Windows D3D11 - limited filter support, just use for decode
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", "d3d11va",
			"-ss", fmt.Sprintf("%.2f", seekTime),
			"-i", srcPath,
			"-vf", fmt.Sprintf("scale=%d:-1", width),
			"-frames:v", "1",
			"-q:v", "2",
			"-y", destPath,
		}
	default:
		return fmt.Errorf("unsupported GPU type: %s", gpuType)
	}

	cmd, err := ffmpeg.NewFFmpegCommand(cmdArgs...)
	if err != nil {
		return err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("GPU thumbnail failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GenerateWebMPreview creates an animated WebM preview (much smaller than GIF)
func GenerateWebMPreview(srcPath, destPath string, config VideoPreviewConfig) error {
	fmt.Printf("[DEBUG] Generating WebM preview: %s\n", srcPath)

	// Check if already exists
	if _, err := os.Stat(destPath); err == nil {
		fmt.Printf("[DEBUG] WebM preview already exists: %s\n", destPath)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	durSec := duration.Seconds()
	if durSec < 2 {
		// Very short video, just convert the whole thing
		return generateSimpleWebM(srcPath, destPath, config)
	}

	// Try GPU first if enabled
	if config.UseGPU {
		gpuType, gpuAvailable := DetectGPU()
		if gpuAvailable {
			err := generateWebMWithGPU(srcPath, destPath, durSec, config, gpuType)
			if err == nil {
				return nil
			}
			fmt.Printf("[WARN] GPU WebM generation failed, falling back to CPU: %v\n", err)
		}
	}

	// CPU fallback
	return generateWebMWithCPU(srcPath, destPath, durSec, config)
}

// generateSimpleWebM converts a very short video to WebM
func generateSimpleWebM(srcPath, destPath string, config VideoPreviewConfig) error {
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", fmt.Sprintf("scale=%d:-1,fps=%d", config.Width, config.FPS),
		"-c:v", "libvpx-vp9",
		"-crf", fmt.Sprintf("%d", config.CRF),
		"-b:v", "0",
		"-an", // No audio
		"-y", destPath,
	)
	if err != nil {
		return err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("simple WebM failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// generateWebMWithCPU creates WebM preview using CPU
func generateWebMWithCPU(srcPath, destPath string, durSec float64, config VideoPreviewConfig) error {
	fmt.Printf("[DEBUG] Generating WebM with CPU\n")

	// Build select filter for evenly distributed clips
	// This extracts short clips at regular intervals throughout the video
	interval := durSec / float64(config.NumClips+1)
	
	var selectParts []string
	for i := 1; i <= config.NumClips; i++ {
		startTime := interval * float64(i)
		endTime := startTime + config.ClipDuration
		selectParts = append(selectParts, fmt.Sprintf("between(t\\,%.2f\\,%.2f)", startTime, endTime))
	}

	selectFilter := strings.Join(selectParts, "+")
	filter := fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,scale=%d:-1,fps=%d", 
		selectFilter, config.Width, config.FPS)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", filter,
		"-c:v", "libvpx-vp9",
		"-crf", fmt.Sprintf("%d", config.CRF),
		"-b:v", "0",
		"-an",
		"-y", destPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("WebM generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Generated WebM preview: %s\n", destPath)
	return nil
}

// generateWebMWithGPU creates WebM preview using GPU acceleration
func generateWebMWithGPU(srcPath, destPath string, durSec float64, config VideoPreviewConfig, gpuType string) error {
	fmt.Printf("[DEBUG] Generating WebM with GPU (%s)\n", gpuType)

	// Build select filter
	interval := durSec / float64(config.NumClips+1)
	
	var selectParts []string
	for i := 1; i <= config.NumClips; i++ {
		startTime := interval * float64(i)
		endTime := startTime + config.ClipDuration
		selectParts = append(selectParts, fmt.Sprintf("between(t\\,%.2f\\,%.2f)", startTime, endTime))
	}
	selectFilter := strings.Join(selectParts, "+")

	var cmdArgs []string

	switch gpuType {
	case "cuda":
		// NVIDIA CUDA - use hardware decode, software encode (VP9 doesn't have NVENC)
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", "cuda",
			"-i", srcPath,
			"-vf", fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,scale=%d:-1,fps=%d",
				selectFilter, config.Width, config.FPS),
			"-c:v", "libvpx-vp9",
			"-crf", fmt.Sprintf("%d", config.CRF),
			"-b:v", "0",
			"-an",
			"-y", destPath,
		}
	case "vaapi":
		// VAAPI with VP9 encode support (Intel)
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", "vaapi",
			"-hwaccel_output_format", "vaapi",
			"-i", srcPath,
			"-vf", fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,scale_vaapi=w=%d:h=-1,fps=%d",
				selectFilter, config.Width, config.FPS),
			"-c:v", "vp9_vaapi",
			"-global_quality", fmt.Sprintf("%d", config.CRF),
			"-an",
			"-y", destPath,
		}
	default:
		// Fallback to CPU with hardware decode only
		cmdArgs = []string{
			"-loglevel", "warning",
			"-hwaccel", gpuType,
			"-i", srcPath,
			"-vf", fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,scale=%d:-1,fps=%d",
				selectFilter, config.Width, config.FPS),
			"-c:v", "libvpx-vp9",
			"-crf", fmt.Sprintf("%d", config.CRF),
			"-b:v", "0",
			"-an",
			"-y", destPath,
		}
	}

	cmd, err := ffmpeg.NewFFmpegCommand(cmdArgs...)
	if err != nil {
		return err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("GPU WebM generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Generated WebM preview with GPU: %s\n", destPath)
	return nil
}

// PreviewPaths holds paths to both static thumbnail and animated preview
type PreviewPaths struct {
	Thumbnail string   // Static JPG thumbnail (always generated, fast)
	Animated  string   // Animated preview directory (for frame-based animation)
	Frames    []string // Individual frame paths (populated after extraction)
}

// GetPreviewPaths returns the expected paths for thumbnail and animated preview
func GetPreviewPaths(srcPath, thumbDir string) PreviewPaths {
	thumbFilename := GenerateUniqueFilename(srcPath, ".jpg")
	// For frame-based animation, we use a subdirectory per video
	animDirName := GenerateUniqueFilename(srcPath, "")
	
	return PreviewPaths{
		Thumbnail: filepath.Join(thumbDir, thumbFilename),
		Animated:  filepath.Join(thumbDir, "frames", animDirName),
	}
}

// FrameExtractionConfig configures frame extraction for animation
type FrameExtractionConfig struct {
	NumFrames int // Number of frames to extract (default 8)
	Width     int // Width of output frames (default 320)
	Quality   int // JPEG quality 1-31, lower is better (default 2)
}

// DefaultFrameExtractionConfig returns sensible defaults
func DefaultFrameExtractionConfig() FrameExtractionConfig {
	return FrameExtractionConfig{
		NumFrames: 8,
		Width:     320,
		Quality:   2,
	}
}

// GenerateAnimationFrames extracts evenly distributed frames from a video
// Returns the paths to extracted frame files
func GenerateAnimationFrames(srcPath, destDir string, config FrameExtractionConfig) ([]string, error) {
	fmt.Printf("[DEBUG] Generating animation frames: %s -> %s\n", srcPath, destDir)

	// Check if frames already exist
	existingFrames := getExistingFrames(destDir, config.NumFrames)
	if len(existingFrames) == config.NumFrames {
		fmt.Printf("[DEBUG] Animation frames already exist: %d frames\n", len(existingFrames))
		return existingFrames, nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create frames directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video duration: %w", err)
	}

	durSec := duration.Seconds()
	if durSec < 1 {
		// Very short video, just extract a single frame
		config.NumFrames = 1
	}

	// Try GPU first
	gpuType, gpuAvailable := DetectGPU()
	if gpuAvailable {
		frames, err := extractFramesWithGPU(srcPath, destDir, durSec, config, gpuType)
		if err == nil {
			return frames, nil
		}
		fmt.Printf("[WARN] GPU frame extraction failed, falling back to CPU: %v\n", err)
	}

	// CPU fallback
	return extractFramesWithCPU(srcPath, destDir, durSec, config)
}

// getExistingFrames checks if frames already exist
func getExistingFrames(destDir string, expectedCount int) []string {
	var frames []string
	for i := 1; i <= expectedCount; i++ {
		framePath := filepath.Join(destDir, fmt.Sprintf("frame_%03d.jpg", i))
		if _, err := os.Stat(framePath); err == nil {
			frames = append(frames, framePath)
		} else {
			break // Stop at first missing frame
		}
	}
	return frames
}

// extractFramesWithCPU extracts frames using CPU
func extractFramesWithCPU(srcPath, destDir string, durSec float64, config FrameExtractionConfig) ([]string, error) {
	fmt.Printf("[DEBUG] Extracting frames with CPU\n")

	var framePaths []string

	// Calculate timestamps for evenly distributed frames
	interval := durSec / float64(config.NumFrames+1)

	for i := 1; i <= config.NumFrames; i++ {
		timestamp := interval * float64(i)
		framePath := filepath.Join(destDir, fmt.Sprintf("frame_%03d.jpg", i))

		err := extractSingleFrame(srcPath, framePath, timestamp, config.Width, config.Quality, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to extract frame %d: %w", i, err)
		}
		framePaths = append(framePaths, framePath)
	}

	fmt.Printf("[DEBUG] Extracted %d frames to %s\n", len(framePaths), destDir)
	return framePaths, nil
}

// extractFramesWithGPU extracts frames using GPU acceleration
func extractFramesWithGPU(srcPath, destDir string, durSec float64, config FrameExtractionConfig, gpuType string) ([]string, error) {
	fmt.Printf("[DEBUG] Extracting frames with GPU (%s)\n", gpuType)

	var framePaths []string

	// Calculate timestamps for evenly distributed frames
	interval := durSec / float64(config.NumFrames+1)

	for i := 1; i <= config.NumFrames; i++ {
		timestamp := interval * float64(i)
		framePath := filepath.Join(destDir, fmt.Sprintf("frame_%03d.jpg", i))

		err := extractSingleFrame(srcPath, framePath, timestamp, config.Width, config.Quality, gpuType, "gpu")
		if err != nil {
			// If GPU fails, try CPU for this frame
			err = extractSingleFrame(srcPath, framePath, timestamp, config.Width, config.Quality, "", "")
			if err != nil {
				return nil, fmt.Errorf("failed to extract frame %d: %w", i, err)
			}
		}
		framePaths = append(framePaths, framePath)
	}

	fmt.Printf("[DEBUG] Extracted %d frames with GPU to %s\n", len(framePaths), destDir)
	return framePaths, nil
}

// extractSingleFrame extracts a single frame at a given timestamp
func extractSingleFrame(srcPath, destPath string, timestamp float64, width, quality int, gpuType, mode string) error {
	var cmdArgs []string

	if mode == "gpu" && gpuType != "" {
		switch gpuType {
		case "cuda":
			cmdArgs = []string{
				"-loglevel", "warning",
				"-hwaccel", "cuda",
				"-ss", fmt.Sprintf("%.2f", timestamp),
				"-i", srcPath,
				"-vf", fmt.Sprintf("scale=%d:-1", width),
				"-frames:v", "1",
				"-q:v", fmt.Sprintf("%d", quality),
				"-y", destPath,
			}
		case "d3d11va":
			cmdArgs = []string{
				"-loglevel", "warning",
				"-hwaccel", "d3d11va",
				"-ss", fmt.Sprintf("%.2f", timestamp),
				"-i", srcPath,
				"-vf", fmt.Sprintf("scale=%d:-1", width),
				"-frames:v", "1",
				"-q:v", fmt.Sprintf("%d", quality),
				"-y", destPath,
			}
		default:
			cmdArgs = []string{
				"-loglevel", "warning",
				"-hwaccel", gpuType,
				"-ss", fmt.Sprintf("%.2f", timestamp),
				"-i", srcPath,
				"-vf", fmt.Sprintf("scale=%d:-1", width),
				"-frames:v", "1",
				"-q:v", fmt.Sprintf("%d", quality),
				"-y", destPath,
			}
		}
	} else {
		// CPU mode
		cmdArgs = []string{
			"-loglevel", "warning",
			"-ss", fmt.Sprintf("%.2f", timestamp),
			"-i", srcPath,
			"-vf", fmt.Sprintf("scale=%d:-1", width),
			"-frames:v", "1",
			"-q:v", fmt.Sprintf("%d", quality),
			"-y", destPath,
		}
	}

	cmd, err := ffmpeg.NewFFmpegCommand(cmdArgs...)
	if err != nil {
		return err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("frame extraction failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetAnimationFramePaths returns the expected paths for animation frames
func GetAnimationFramePaths(destDir string, numFrames int) []string {
	var paths []string
	for i := 1; i <= numFrames; i++ {
		paths = append(paths, filepath.Join(destDir, fmt.Sprintf("frame_%03d.jpg", i)))
	}
	return paths
}

// EnsureVideoThumbnail generates a static thumbnail if it doesn't exist
func EnsureVideoThumbnail(srcPath, thumbDir string) (string, error) {
	paths := GetPreviewPaths(srcPath, thumbDir)
	
	// Check if thumbnail exists
	if _, err := os.Stat(paths.Thumbnail); err == nil {
		return paths.Thumbnail, nil
	}

	// Generate thumbnail
	err := GenerateVideoThumbnail(srcPath, paths.Thumbnail, 320)
	if err != nil {
		return "", err
	}

	return paths.Thumbnail, nil
}

// EnsureAnimatedPreview generates animation frames if they don't exist
// Returns the paths to all extracted frames
func EnsureAnimatedPreview(srcPath, thumbDir string) ([]string, error) {
	paths := GetPreviewPaths(srcPath, thumbDir)
	
	// Generate animation frames
	config := DefaultFrameExtractionConfig()
	frames, err := GenerateAnimationFrames(srcPath, paths.Animated, config)
	if err != nil {
		return nil, err
	}

	return frames, nil
}
