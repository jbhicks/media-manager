package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/user/media-manager/internal/ffmpeg"
)

var (
	poolOnce sync.Once
	taskChan chan task
)

type task struct {
	srcPath   string
	destPath  string
	mediaType string
	done      chan error
}

func initPool() {
	poolOnce.Do(func() {
		numWorkers := runtime.NumCPU()
		if numWorkers > 4 {
			numWorkers = 4 // Cap at 4 workers to avoid overwhelming I/O
		}
		taskChan = make(chan task, 100)
		for i := 0; i < numWorkers; i++ {
			go worker()
		}
	})
}

func worker() {
	for t := range taskChan {
		var err error
		if t.mediaType == "video" {
			err = GenerateAnimatedPreview(t.srcPath, t.destPath)
		} else {
			err = GenerateThumbnail(t.srcPath, t.destPath)
		}
		if t.done != nil {
			t.done <- err
		}
	}
}

// GenerateUniqueFilename creates a unique hex filename based on the source path
func GenerateUniqueFilename(filePath, extension string) string {
	hash := sha256.Sum256([]byte(filePath))
	return hex.EncodeToString(hash[:]) + extension
}

// GetPreview returns the path to the preview, generating it in the background if it doesn't exist.
func GetPreview(srcPath string, mediaType string, thumbDir string) (string, error) {
	initPool()

	ext := ".jpg"
	if mediaType == "video" {
		ext = ".gif"
	}

	destFilename := GenerateUniqueFilename(srcPath, ext)
	destPath := filepath.Join(thumbDir, destFilename)

	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	// Queue for generation
	select {
	case taskChan <- task{srcPath: srcPath, destPath: destPath, mediaType: mediaType}:
		// Task queued
	default:
		// Queue full, maybe return error or log
		fmt.Printf("[WARN] Preview task queue full for: %s\n", srcPath)
	}

	return destPath, nil // Return the path even if it doesn't exist yet
}

func getUserConfig(key string, defaultValue int) int {
	// Placeholder implementation for user configuration
	return defaultValue
}

func fileExists(path string) bool {
	fmt.Printf("[DEBUG] Checking existence of: %s\n", path)
	_, err := os.Stat(path)
	fmt.Printf("[DEBUG] os.Stat error: %v\n", err)
	return err == nil
}

func pathWritable(path string) bool {
	fmt.Printf("[DEBUG] Checking writability of: %s\n", path)
	file, err := os.Create(path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// GenerateThumbnail creates a thumbnail for the given file path.
func GenerateThumbnail(filePath, thumbPath string) error {
	filePath = filepath.Clean(filePath)
	thumbPath = filepath.Clean(thumbPath)
	fmt.Printf("[DEBUG] Generating thumbnail for: %s\n", filePath)
	fmt.Printf("[DEBUG] Output path: %s\n", thumbPath)

	// Only generate static thumbnails for images
	fileExt := strings.ToLower(filepath.Ext(filePath))
	if IsImageFile(fileExt) {
		// Ensure the output directory exists
		thumbDir := filepath.Dir(thumbPath)
		if err := os.MkdirAll(thumbDir, 0755); err != nil {
			return fmt.Errorf("failed to create thumbnail directory: %w", err)
		}
		// Check if the source file exists before attempting to generate a thumbnail
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", filePath)
		}
		return generateImageThumbnail(filePath, thumbPath)
	}
	return fmt.Errorf("unsupported file type for static thumbnail: %s", fileExt)
}

// GetMetadata extracts width, height and duration using ffprobe
func GetMetadata(filePath string) (width, height, duration int, err error) {
	cmd, err := ffmpeg.NewFFprobeCommand(
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-of", "csv=s=x:p=0",
		filePath,
	)
	if err != nil {
		return 0, 0, 0, err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Format: 1920x1080x123.456 (or just 1920x1080 if duration missing)
	parts := strings.Split(strings.TrimSpace(string(output)), "x")
	if len(parts) >= 2 {
		width, _ = strconv.Atoi(parts[0])
		height, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		durFloat, _ := strconv.ParseFloat(parts[2], 64)
		duration = int(durFloat)
	}

	return width, height, duration, nil
}

func IsImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp", ".ico", ".svg"}
	return slices.Contains(imageExts, ext)
}

func IsVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v", ".3gp", ".ogv", ".flv", ".asf", ".dv", ".mp2", ".mpg", ".mpeg", ".rm", ".wmv", ".ts", ".vob", ".divx"}
	return slices.Contains(videoExts, ext)
}

func generateImageThumbnail(srcPath, thumbPath string) error {
	fmt.Printf("[DEBUG] Generating blurred-background thumbnail for: %s\n", srcPath)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", "scale=216:216:force_original_aspect_ratio=increase,boxblur=20,setsar=1[bg];[0:v]scale=216:216:force_original_aspect_ratio=decrease,setsar=1[fg];[bg][fg]overlay=(W-w)/2:(H-h)/2",
		"-vframes", "1",
		"-y",
		thumbPath,
	)
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed to generate image thumbnail: %w\n[ffmpeg output]: %s", err, string(output))
	}

	return nil
}

func generateVideoThumbnail(srcPath, thumbPath string) error {
	// Use FFmpeg to extract a frame from the video with uniform dimensions
	fmt.Printf("[DEBUG] Running ffmpeg command: ffmpeg -i %s -ss 00:00:01 -vframes 1 -vf scale=180:180:force_original_aspect_ratio=increase,crop=180:180 -y %s\n", srcPath, thumbPath)
	fmt.Printf("[DEBUG] Source file exists: %v\n", fileExists(srcPath))
	fmt.Printf("[DEBUG] Thumbnail path writable: %v\n", pathWritable(thumbPath))
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1", // Extract only 1 frame
		"-vf", "scale=180:101:force_original_aspect_ratio=increase,crop=180:101", // Scale and crop to 180x101
		"-f", "image2", // Explicitly set output format to image2 to avoid sequence pattern errors
		"-y", // Overwrite output file
		thumbPath,
	)
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg: %w", err)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not contain an image sequence pattern") {
			fmt.Printf("[WARN] ffmpeg: output filename may need a sequence pattern or -update option.\n")
		}
		fmt.Printf("[ERROR] ffmpeg error: %v\n[ffmpeg output]: %s\n", err, string(output))
		return fmt.Errorf("failed to generate video thumbnail: %w\n[ffmpeg output]: %s", err, string(output))
	}
	if len(output) > 0 {
		fmt.Printf("[ffmpeg warning]: %s\n", string(output))
	}
	if _, err := os.Stat(thumbPath); err != nil {
		return fmt.Errorf("Thumbnail file missing: %s", thumbPath)
	}
	return nil
}

func getVideoDuration(filePath string) (time.Duration, error) {
	cmd, err := ffmpeg.NewFFprobeCommand(
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get ffprobe: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration for %s: %w", filePath, err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration %s: %w", durationStr, err)
	}

	return time.Duration(durationFloat * float64(time.Second)), nil
}

// GenerateAnimatedPreview creates a single animated GIF for video preview

func GenerateAnimatedPreviewCPU(srcPath, gifPath string) error {
	fmt.Printf("[DEBUG] Generating 2x2 animated grid preview for: %s\n", srcPath)

	// Check if animated preview already exists
	if _, err := os.Stat(gifPath); err == nil {
		fmt.Printf("[DEBUG] Animated preview already exists: %s\n", gifPath)
		return nil
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(gifPath), 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}
	durSec := duration.Seconds()

	// Capture 4 segments at 10%, 40%, 70%, 90%
	// Using a shorter segment (e.g. 0.8s) to keep GIF size small
	segLen := 0.8
	p1, p4, p7, p9 := durSec*0.1, durSec*0.4, durSec*0.7, durSec*0.9

	// Filtergraph for 2x2 grid
	// Each quadrant is 108x108 (total 216x216)
	filtergraph := fmt.Sprintf(
		"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=108:108:force_original_aspect_ratio=increase,crop=108:108[v1];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=108:108:force_original_aspect_ratio=increase,crop=108:108[v2];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=108:108:force_original_aspect_ratio=increase,crop=108:108[v3];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=108:108:force_original_aspect_ratio=increase,crop=108:108[v4];"+
			"[v1][v2][v3][v4]xstack=inputs=4:layout=0_0|108_0|0_108|108_108,fps=12[outv]",
		p1, p1+segLen, p4, p4+segLen, p7, p7+segLen, p9, p9+segLen,
	)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-filter_complex", filtergraph,
		"-map", "[outv]",
		"-y",
		gifPath,
	)
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed to generate 2x2 animated preview: %v, output: %s\n", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated 2x2 animated preview: %s\n", gifPath)
	return nil
}

// GetFFmpegHardwareAccelerations returns a list of supported hardware accelerations by ffmpeg.
func GetFFmpegHardwareAccelerations() ([]string, error) {
	cmd, err := ffmpeg.NewFFmpegCommand("-hwaccels")
	if err != nil {
		return nil, fmt.Errorf("failed to get ffmpeg: %w", err)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run ffmpeg -hwaccels: %w\n%s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	var hwaccels []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue // Skip header line
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			hwaccels = append(hwaccels, fields[0])
		}
	}
	return hwaccels, nil
}

func GenerateAnimatedPreview(srcPath, gifPath string) error {
	// Temporarily disable GPU acceleration for debugging
	return GenerateAnimatedPreviewCPU(srcPath, gifPath)
}

// ExtractGifFrames extracts all frames from a GIF into a sequence of PNG images.
// remove1x1Frames removes any images in framePaths that are 1x1 pixels and returns the filtered list.
func remove1x1Frames(framePaths []string) ([]string, error) {
	var filtered []string
	for _, path := range framePaths {
		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("[DEBUG] Could not open frame %s: %v\n", path, err)
			continue
		}
		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			fmt.Printf("[DEBUG] Could not decode frame %s: %v\n", path, err)
			continue
		}
		bounds := img.Bounds()
		if bounds.Dx() == 1 && bounds.Dy() == 1 {
			fmt.Printf("[DEBUG] Removing 1x1 frame: %s\n", path)
			os.Remove(path)
			continue
		}
		filtered = append(filtered, path)
	}
	return filtered, nil
}

func ExtractGifFrames(gifPath, outputDir string) ([]string, error) {
	fmt.Printf("[DEBUG] Extracting frames from GIF: %s to %s\n", gifPath, outputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory for GIF frames: %w", err)
	}

	// FFmpeg command to extract frames
	outputPattern := filepath.Join(outputDir, "frame_%d.jpg")
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", gifPath,
		"-vsync", "0", // Ensure all frames are extracted
		"-vf", "fps=8", // Force 8 frames per second
		"-frame_pts", "1", // Add presentation timestamp to frame filename
		"-f", "image2", // Force image2 format
		"-qscale:v", "2", // High quality jpeg output
		outputPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get ffmpeg: %w", err)
	}

	fmt.Printf("[DEBUG] Running ffmpeg for frame extraction: %v\n", cmd.Args)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[ERROR] ffmpeg error: %v\n[ffmpeg output]: %s\n", err, string(output))
		return nil, fmt.Errorf("failed to extract GIF frames: %v, output: %s", err, string(output))
	}
	if len(output) > 0 {
		fmt.Printf("[ffmpeg warning]: %s\n", string(output))
	}
	// Collect paths of extracted frames
	var framePaths []string
	// This is a bit hacky, but we need to find the number of frames extracted.
	// A more robust solution would be to parse ffmpeg output or use ffprobe.
	// For now, we'll assume a reasonable number of frames and check existence.
	// Look for frames up to a reasonable number
	for i := range [24]int{} {
		framePath := filepath.Join(outputDir, fmt.Sprintf("frame_%d.jpg", i))
		if _, err := os.Stat(framePath); err == nil {
			framePaths = append(framePaths, framePath)
			fmt.Printf("[DEBUG] Found frame: %s\n", framePath)
		} else if os.IsNotExist(err) {
			fmt.Printf("[DEBUG] Frame does not exist: %s\n", framePath)
			// Don't break - try to find all available frames
		}
	}

	if len(framePaths) == 0 {
		return nil, fmt.Errorf("no frames extracted from GIF: %s", gifPath)
	}

	fmt.Printf("[DEBUG] Successfully extracted %d frames.\n", len(framePaths))
	// Remove 1x1 frames as post-processing
	filtered, err := remove1x1Frames(framePaths)
	if err != nil {
		fmt.Printf("[DEBUG] Error during 1x1 frame removal: %v\n", err)
	}
	return filtered, nil
}

func GenerateAnimatedPreviewGPU(srcPath, gifPath, hwaccel string) error {

	fmt.Printf("[DEBUG] Generating animated GIF preview with GPU (%s) for: %s\n", hwaccel, srcPath)

	// Check if animated preview already exists
	if _, err := os.Stat(gifPath); err == nil {
		fmt.Printf("[DEBUG] Animated preview already exists: %s\n", gifPath)
		return nil
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(gifPath), 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	// Calculate timestamps for 10 evenly distributed frames
	numFrames := 10
	interval := duration / time.Duration(numFrames+1)
	var selectFilters []string
	for i := 1; i <= numFrames; i++ {
		seekTime := interval * time.Duration(i)
		selectFilters = append(selectFilters, fmt.Sprintf("eq(n,\"%d\")", int(seekTime.Seconds()*25))) // Assuming 25fps for frame selection
	}

	var cmdArgs []string
	filterComplex := fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,fps=12,scale=180:180:force_original_aspect_ratio=increase,crop=180:180", strings.Join(selectFilters, "+"))

	switch hwaccel {
	case "cuda":
		cmdArgs = []string{
			"-hwaccel", "cuda",
			"-c:v", "h264_cuvid", // Assuming H.264 input, adjust as needed
			"-i", srcPath,
			"-vf", "hwupload_cuda," + filterComplex,
			"-f", "gif",
			"-y",
			gifPath,
		}
	case "vaapi":
		cmdArgs = []string{
			"-hwaccel", "vaapi",
			"-i", srcPath,
			"-vf", "format=nv12,hwupload,scale_vaapi=w=200:h=200:force_original_aspect_ratio=increase,crop=200:200,hwdownload,format=bgr0", // Example VAAPI filter
			"-f", "gif",
			"-y",
			gifPath,
		}
	case "nvenc":
		// NVENC is primarily an encoder, so decoding might still be CPU-bound or require specific decoders
		// For simplicity, we'll use a generic GPU filter here, but a real implementation might need more specific handling
		cmdArgs = []string{
			"-i", srcPath,
			"-vf", "scale=200:-1", // Use the same filter as CPU for now, as NVENC is for encoding
			"-f", "gif",
			"-y",
			gifPath,
		}
	default:
		return fmt.Errorf("unsupported hardware acceleration: %s", hwaccel)
	}

	cmd, err := ffmpeg.NewFFmpegCommand(cmdArgs...)
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate animated preview with GPU (%s): %v, output: %s\n", hwaccel, err, string(output))
	}

	// Verify the GIF was created
	if _, err := os.Stat(gifPath); err != nil {
		return fmt.Errorf("animated preview file missing after GPU generation: %s", gifPath)
	}

	fmt.Printf("[DEBUG] Successfully generated animated preview with GPU (%s): %s\n", hwaccel, gifPath)
	return nil
}
