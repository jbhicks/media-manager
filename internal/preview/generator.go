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
	"sort"
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
	taskType  string // "thumbnail" or "animated"
	done      chan error
	callback  func(result interface{}, err error) // Callback when complete (string for thumbnail, []string for frames)
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
		var result interface{}

		switch t.taskType {
		case "animated":
			// Generate animation frames (8 JPG files)
			config := DefaultFrameExtractionConfig()
			frames, frameErr := GenerateAnimationFrames(t.srcPath, t.destPath, config)
			err = frameErr
			result = frames // []string of frame paths
		case "thumbnail":
			if t.mediaType == "video" {
				// Generate static video thumbnail (fast)
				err = GenerateVideoThumbnail(t.srcPath, t.destPath, 320)
			} else {
				err = GenerateThumbnail(t.srcPath, t.destPath)
			}
			result = t.destPath
		default:
			// Legacy behavior
			if t.mediaType == "video" {
				err = GenerateVideoThumbnail(t.srcPath, t.destPath, 320)
			} else {
				err = GenerateThumbnail(t.srcPath, t.destPath)
			}
			result = t.destPath
		}
		if t.done != nil {
			t.done <- err
		}
		if t.callback != nil {
			t.callback(result, err)
		}
	}
}

// GenerateUniqueFilename creates a unique hex filename based on the source path
func GenerateUniqueFilename(filePath, extension string) string {
	hash := sha256.Sum256([]byte(filePath))
	return hex.EncodeToString(hash[:]) + extension
}

// GetPreview returns the path to the preview, generating it in the background if it doesn't exist.
// For videos, this returns a static JPG thumbnail (fast). Use GetAnimatedPreview for WebM.
func GetPreview(srcPath string, mediaType string, thumbDir string) (string, error) {
	return GetPreviewWithForce(srcPath, mediaType, thumbDir, false)
}

// GetPreviewWithForce returns the path to the preview, optionally forcing regeneration.
// For videos, this returns a static JPG thumbnail (fast).
func GetPreviewWithForce(srcPath string, mediaType string, thumbDir string, forceRegenerate bool) (string, error) {
	initPool()

	// Always use JPG for static thumbnails (both images and videos)
	ext := ".jpg"
	destFilename := GenerateUniqueFilename(srcPath, ext)
	destPath := filepath.Join(thumbDir, destFilename)

	// Check if preview exists and skip if not forcing regeneration
	if !forceRegenerate {
		if _, err := os.Stat(destPath); err == nil {
			return destPath, nil
		}
	} else {
		// Remove existing preview to force regeneration
		os.Remove(destPath)
		fmt.Printf("[DEBUG] Force regenerating preview: %s\n", srcPath)
	}

	// Queue for generation
	select {
	case taskChan <- task{srcPath: srcPath, destPath: destPath, mediaType: mediaType, taskType: "thumbnail"}:
		fmt.Printf("[DEBUG] Queued thumbnail generation for: %s\n", filepath.Base(srcPath))
	default:
		// Queue full, maybe return error or log
		fmt.Printf("[WARN] Preview task queue full for: %s\n", srcPath)
	}

	return destPath, nil // Return the path even if it doesn't exist yet
}

// GetPreviewWithCallback queues preview generation and calls the callback when complete.
// For videos, this generates a static JPG thumbnail (fast).
func GetPreviewWithCallback(srcPath string, mediaType string, thumbDir string, forceRegenerate bool, callback func(path string, err error)) (string, error) {
	initPool()

	// Always use JPG for static thumbnails
	ext := ".jpg"
	destFilename := GenerateUniqueFilename(srcPath, ext)
	destPath := filepath.Join(thumbDir, destFilename)

	// Check if preview exists and skip if not forcing regeneration
	if !forceRegenerate {
		if _, err := os.Stat(destPath); err == nil {
			// File already exists, invoke callback immediately
			if callback != nil {
				go callback(destPath, nil)
			}
			return destPath, nil
		}
	} else {
		// Remove existing preview to force regeneration
		os.Remove(destPath)
		fmt.Printf("[DEBUG] Force regenerating preview: %s\n", srcPath)
	}

	// Wrap callback to convert interface{} to string
	wrappedCallback := func(result interface{}, err error) {
		if callback != nil {
			if path, ok := result.(string); ok {
				callback(path, err)
			} else {
				callback("", err)
			}
		}
	}

	// Queue for generation with callback
	select {
	case taskChan <- task{srcPath: srcPath, destPath: destPath, mediaType: mediaType, taskType: "thumbnail", callback: wrappedCallback}:
		fmt.Printf("[DEBUG] Queued thumbnail generation for: %s\n", filepath.Base(srcPath))
	default:
		// Queue full
		fmt.Printf("[WARN] Preview task queue full for: %s\n", srcPath)
		if callback != nil {
			go callback("", fmt.Errorf("preview task queue full"))
		}
	}

	return destPath, nil
}

// GetAnimatedPreviewWithCallback queues animation frame extraction
// This is called on hover to load the animated preview frames
// The callback receives []string (frame paths) on success
func GetAnimatedPreviewWithCallback(srcPath string, thumbDir string, forceRegenerate bool, callback func(framePaths []string, err error)) (string, error) {
	initPool()

	// Use a subdirectory for animation frames
	framesDir := GenerateUniqueFilename(srcPath, "")
	destPath := filepath.Join(thumbDir, "frames", framesDir)

	// Check if frames already exist and skip if not forcing regeneration
	if !forceRegenerate {
		config := DefaultFrameExtractionConfig()
		existingFrames := getExistingFrames(destPath, config.NumFrames)
		if len(existingFrames) == config.NumFrames {
			// Frames already exist, invoke callback immediately
			if callback != nil {
				go callback(existingFrames, nil)
			}
			return destPath, nil
		}
	} else {
		// Remove existing frames directory to force regeneration
		os.RemoveAll(destPath)
		fmt.Printf("[DEBUG] Force regenerating animation frames: %s\n", srcPath)
	}

	// Wrap callback to convert interface{} to []string
	wrappedCallback := func(result interface{}, err error) {
		if callback != nil {
			if frames, ok := result.([]string); ok {
				callback(frames, err)
			} else {
				callback(nil, err)
			}
		}
	}

	// Queue for generation with callback
	select {
	case taskChan <- task{srcPath: srcPath, destPath: destPath, mediaType: "video", taskType: "animated", callback: wrappedCallback}:
		fmt.Printf("[DEBUG] Queued animation frame extraction for: %s\n", filepath.Base(srcPath))
	default:
		fmt.Printf("[WARN] Preview task queue full for animated: %s\n", srcPath)
		if callback != nil {
			go callback(nil, fmt.Errorf("preview task queue full"))
		}
	}

	return destPath, nil
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
		"-filter_complex", "scale=216:216:force_original_aspect_ratio=increase,boxblur=20,setsar=1[bg];[0:v]scale=216:216:force_original_aspect_ratio=decrease,setsar=1[fg];[bg][fg]overlay=(W-w)/2:(H-h)/2",
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
	fmt.Printf("[DEBUG] Generating 3x2 animated grid preview for: %s\n", srcPath)

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

	// Capture 6 segments at 12.5%, 25%, 37.5%, 50%, 62.5%, 75%
	// Using longer segments (1.0s) for smoother animation
	segLen := 1.0
	p1, p2, p3, p4, p5, p6 := durSec*0.125, durSec*0.25, durSec*0.375, durSec*0.5, durSec*0.625, durSec*0.75

	// Filtergraph for 3x2 grid
	// Each quadrant is 72x72 (total 216x144)
	filtergraph := fmt.Sprintf(
		"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v1];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v2];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v3];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v4];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v5];"+
			"[0:v]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS,scale=72:72:force_original_aspect_ratio=increase,crop=72:72[v6];"+
			"[v1][v2][v3][v4][v5][v6]xstack=inputs=6:layout=0_0|72_0|144_0|0_72|72_72|144_72,fps=10[outv]",
		p1, p1+segLen, p2, p2+segLen, p3, p3+segLen, p4, p4+segLen, p5, p5+segLen, p6, p6+segLen,
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

	fmt.Printf("[DEBUG] Running ffmpeg command: %v\n", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed to generate 3x2 animated preview: %v, output: %s\n", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated 3x2 animated preview: %s\n", gifPath)
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

// PreviewOptions configures preview generation behavior
type PreviewOptions struct {
	SceneThreshold float64 // 0.0-1.0, scene detection sensitivity (lower = more sensitive)
	MaxScenes      int     // Maximum number of scenes to include
	FPS            int     // Output frame rate
	Width          int     // Output width in pixels
	Height         int     // Output height in pixels
	UseGPU         bool    // Enable GPU acceleration
	GPUType        string  // "cuda", "vaapi", etc.
	UseMosaic      bool    // Generate static mosaic instead of GIF
}

// DefaultPreviewOptions returns sensible defaults
func DefaultPreviewOptions() PreviewOptions {
	return PreviewOptions{
		SceneThreshold: 0.3, // Lower threshold for more sensitive scene detection
		MaxScenes:      8,   // Increased to 8 scenes for better coverage
		FPS:            6,   // Reduced to 6 FPS for smoother animation
		Width:          256, // Increased width for better detail
		Height:         144, // 16:9 aspect ratio
		UseGPU:         false,
		GPUType:        "",
		UseMosaic:      false,
	}
}

// SceneTimestamp represents a scene change detection result
type SceneTimestamp struct {
	Time  float64 // Time in seconds
	Score float64 // Scene change score
}

// detectScenes analyzes a video and returns timestamps where scene changes occur
func detectScenes(srcPath string, threshold float64) ([]SceneTimestamp, error) {
	fmt.Printf("[DEBUG] Detecting scenes in: %s (threshold: %.2f)\n", srcPath, threshold)

	// Use ffmpeg with select filter for scene detection and showinfo to get timestamps
	cmd, err := ffmpeg.NewFFmpegCommand(
		"-i", srcPath,
		"-vf", fmt.Sprintf("select='gt(scene\\,%.2f)',showinfo", threshold),
		"-vsync", "0",
		"-f", "null",
		"-",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ffmpeg command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Scene detection can succeed even if command returns non-zero
		// Check if we got output
		if len(output) == 0 {
			return nil, fmt.Errorf("scene detection failed: %w", err)
		}
	}

	// Parse scene detection output
	// Format: [Parsed_showinfo_X @ ADDR] n:X pts:Y pts_time:Z.ZZZ ...
	var scenes []SceneTimestamp
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if !strings.Contains(line, "pts_time:") {
			continue
		}

		// Extract pts_time value
		// Example: pts_time:2.04
		parts := strings.Split(line, "pts_time:")
		if len(parts) < 2 {
			continue
		}

		// Get the time value (next token after pts_time:)
		timeStr := strings.Fields(parts[1])[0]
		timeVal, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			continue
		}

		// Score is implicitly above threshold since we used select filter
		scenes = append(scenes, SceneTimestamp{
			Time:  timeVal,
			Score: threshold + 0.1, // Approximate score
		})
	}

	fmt.Printf("[DEBUG] Detected %d scenes\n", len(scenes))
	return scenes, nil
}

// selectRepresentativeScenes picks the most important scenes from the list
func selectRepresentativeScenes(scenes []SceneTimestamp, duration time.Duration, maxScenes int) []float64 {
	if len(scenes) == 0 {
		// Fallback to evenly distributed timestamps
		fmt.Printf("[DEBUG] No scenes detected, using evenly distributed timestamps\n")
		return evenlyDistributedTimestamps(duration, maxScenes)
	}

	if len(scenes) <= maxScenes {
		// Use all detected scenes
		timestamps := make([]float64, len(scenes))
		for i, s := range scenes {
			timestamps[i] = s.Time
		}
		return timestamps
	}

	// Too many scenes, select most significant ones
	// Sort by score (descending)
	sortedScenes := make([]SceneTimestamp, len(scenes))
	copy(sortedScenes, scenes)
	sort.Slice(sortedScenes, func(i, j int) bool {
		return sortedScenes[i].Score > sortedScenes[j].Score
	})

	// Take top N scenes
	topScenes := sortedScenes[:maxScenes]

	// Sort by time for chronological order
	sort.Slice(topScenes, func(i, j int) bool {
		return topScenes[i].Time < topScenes[j].Time
	})

	timestamps := make([]float64, len(topScenes))
	for i, s := range topScenes {
		timestamps[i] = s.Time
	}

	fmt.Printf("[DEBUG] Selected %d representative scenes from %d total\n", len(timestamps), len(scenes))
	return timestamps
}

// evenlyDistributedTimestamps returns N evenly spaced timestamps across the duration
func evenlyDistributedTimestamps(duration time.Duration, count int) []float64 {
	durSec := duration.Seconds()
	timestamps := make([]float64, count)

	// Distribute evenly: for count=6, use 12.5%, 25%, 37.5%, 50%, 62.5%, 75%
	// Skip first 10% and last 15% to avoid intros/outros
	startPercent := 0.1
	endPercent := 0.85
	availableRange := endPercent - startPercent

	for i := 0; i < count; i++ {
		progress := startPercent + (availableRange * float64(i) / float64(count-1))
		timestamps[i] = durSec * progress
	}

	return timestamps
}

func GenerateAnimatedPreview(srcPath, gifPath string) error {
	// Use smart preview generation with scene detection for better scene representation
	opts := DefaultPreviewOptions()
	return GenerateSmartPreview(srcPath, gifPath, opts)
}

// GenerateFastPreview creates a preview WITHOUT scene detection (much faster)
// Uses evenly distributed timestamps instead of analyzing the whole video
func GenerateFastPreview(srcPath, gifPath string, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating fast preview for: %s\n", srcPath)

	// Check if preview already exists
	if _, err := os.Stat(gifPath); err == nil {
		fmt.Printf("[DEBUG] Preview already exists: %s\n", gifPath)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(gifPath), 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	// Use evenly distributed timestamps (FAST - no scene detection)
	timestamps := evenlyDistributedTimestamps(duration, opts.MaxScenes)
	fmt.Printf("[DEBUG] Using evenly distributed timestamps: %v\n", timestamps)

	// Generate preview based on type
	if opts.UseMosaic {
		return generateSceneMosaic(srcPath, gifPath, timestamps, opts)
	}

	return generateSceneBasedGIFWithCPU(srcPath, gifPath, timestamps, opts)
}

// GenerateSmartPreview creates an optimized preview using scene detection (SLOW)
func GenerateSmartPreview(srcPath, gifPath string, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating smart preview for: %s\n", srcPath)

	// Check if preview already exists
	if _, err := os.Stat(gifPath); err == nil {
		fmt.Printf("[DEBUG] Preview already exists: %s\n", gifPath)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(gifPath), 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	// Detect scenes
	scenes, err := detectScenes(srcPath, opts.SceneThreshold)
	if err != nil {
		fmt.Printf("[WARN] Scene detection failed: %v, falling back to even distribution\n", err)
		scenes = nil // Will trigger fallback
	}

	// Select representative timestamps
	timestamps := selectRepresentativeScenes(scenes, duration, opts.MaxScenes)

	// Generate preview based on type
	if opts.UseMosaic {
		return generateSceneMosaic(srcPath, gifPath, timestamps, opts)
	}

	// Use GPU or CPU based on options
	if opts.UseGPU && opts.GPUType != "" {
		err := generateSceneBasedGIFWithGPU(srcPath, gifPath, timestamps, opts)
		if err != nil {
			fmt.Printf("[WARN] GPU generation failed: %v, falling back to CPU\n", err)
			return generateSceneBasedGIFWithCPU(srcPath, gifPath, timestamps, opts)
		}
		return err
	}

	return generateSceneBasedGIFWithCPU(srcPath, gifPath, timestamps, opts)
}

// generateSceneBasedGIFWithCPU creates a GIF using two-pass palette encoding
func generateSceneBasedGIFWithCPU(srcPath, gifPath string, timestamps []float64, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating scene-based GIF with CPU\n")

	// Step 1: Generate optimal palette
	paletteFile := filepath.Join(filepath.Dir(gifPath), ".palette_"+filepath.Base(gifPath)+".png")
	defer os.Remove(paletteFile) // Clean up palette file

	// Build filter for scene extraction
	var sceneFilters []string
	cols := 4 // 4x2 horizontal layout for 8 scenes
	rows := 2
	cellSize := opts.Width / cols

	for i, ts := range timestamps {
		if i >= cols*rows {
			break // Limit to grid size
		}
		// Increased duration to 1.5 seconds for longer clips
		sceneFilters = append(sceneFilters, fmt.Sprintf(
			"[0:v]trim=start=%.2f:duration=1.5,setpts=PTS-STARTPTS,fps=%d,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black,format=rgb24[v%d]",
			ts, opts.FPS, cellSize, cellSize, cellSize, cellSize, i,
		))
	}

	// Create 4x2 horizontal layout
	gridFilter := fmt.Sprintf("[v0][v1][v2][v3][v4][v5][v6][v7]xstack=inputs=8:layout=0_0|%d_0|%d_0|%d_0|0_%d|%d_%d|%d_%d|%d_%d[stacked]",
		cellSize, cellSize*2, cellSize*3, cellSize, cellSize, cellSize*2, cellSize*3, cellSize)

	paletteFilter := strings.Join(sceneFilters, ";") + ";" + gridFilter + ";[stacked]palettegen=max_colors=256"

	cmd1, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-filter_complex", paletteFilter,
		"-y", paletteFile,
	)
	if err != nil {
		return fmt.Errorf("failed to create palette command: %w", err)
	}

	fmt.Printf("[DEBUG] Generating palette...\n")
	output, err := cmd1.CombinedOutput()
	if err != nil {
		return fmt.Errorf("palette generation failed: %w\nOutput: %s", err, string(output))
	}

	// Step 2: Generate GIF using palette
	gifFilter := strings.Join(sceneFilters, ";") + ";" + gridFilter + ";[stacked][1:v]paletteuse=dither=bayer:bayer_scale=5"

	cmd2, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-i", paletteFile,
		"-filter_complex", gifFilter,
		"-y", gifPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create GIF command: %w", err)
	}

	fmt.Printf("[DEBUG] Generating GIF with palette...\n")
	output, err = cmd2.CombinedOutput()
	if err != nil {
		return fmt.Errorf("GIF generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated scene-based GIF: %s\n", gifPath)
	return nil
}

// generateSceneMosaic creates a static image mosaic of scene thumbnails
func generateSceneMosaic(srcPath, mosaicPath string, timestamps []float64, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating scene mosaic for: %s\n", srcPath)

	// Use scene detection with thumbnail filter for best frame selection
	// Create a 4x2 grid for 8 scenes
	cols := 4
	rows := 2
	cellSize := opts.Width / cols

	var sceneFilters []string
	for i, ts := range timestamps {
		if i >= cols*rows {
			break // Limit to grid size
		}
		// Increased duration to 1.5 seconds
		sceneFilters = append(sceneFilters, fmt.Sprintf(
			"[0:v]trim=start=%.2f:duration=1.5,thumbnail=25,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d[v%d]",
			ts, cellSize, cellSize, cellSize, cellSize, i,
		))
	}

	// Create 4x2 grid
	gridFilter := fmt.Sprintf("[v0][v1][v2][v3][v4][v5][v6][v7]xstack=inputs=8:layout=0_0|%d_0|%d_0|%d_0|0_%d|%d_%d|%d_%d|%d_%d",
		cellSize, cellSize*2, cellSize*3, cellSize, cellSize, cellSize*2, cellSize*3, cellSize)

	filterComplex := strings.Join(sceneFilters, ";") + ";" + gridFilter

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-filter_complex", filterComplex,
		"-frames:v", "1",
		"-y", mosaicPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create mosaic command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mosaic generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated mosaic: %s\n", mosaicPath)
	return nil
}

// GenerateSceneBasedPreview creates a single preview image using FFmpeg's scene detection
// This is "Method 1" - auto-detects interesting scene changes and tiles them into a preview image
// The output is a static image, not an animated GIF
func GenerateSceneBasedPreview(srcPath, outputPath string, opts *ScenePreviewOptions) error {
	if opts == nil {
		opts = DefaultScenePreviewOptions()
	}

	fmt.Printf("[DEBUG] Generating scene-based preview (Method 1) for: %s\n", srcPath)

	// Check if preview already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("[DEBUG] Scene preview already exists: %s\n", outputPath)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Build the FFmpeg command for scene-based preview
	// select='gt(scene,threshold)' - Select frames where scene change score > threshold
	// scale=W:H - Scale to thumbnail size
	// tile=COLSxROWS - Arrange selected frames in a grid
	filter := fmt.Sprintf(
		"select='gt(scene\\,%.2f)',scale=%d:%d,tile=%dx%d",
		opts.SceneThreshold,
		opts.TileWidth,
		opts.TileHeight,
		opts.Cols,
		opts.Rows,
	)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", filter,
		"-frames:v", "1",
		"-y", outputPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create scene preview command: %w", err)
	}

	fmt.Printf("[DEBUG] Running FFmpeg scene detection: %v\n", cmd.Args)
	output, err := cmd.CombinedOutput()

	// Check if output was created, regardless of FFmpeg exit code
	if _, statErr := os.Stat(outputPath); statErr == nil {
		fmt.Printf("[DEBUG] Successfully generated scene-based preview: %s\n", outputPath)
		return nil
	}

	// Scene detection failed or didn't find enough scenes - try fallback
	return generateScenePreviewFallback(srcPath, outputPath, opts, output, err)
}

// generateScenePreviewFallback tries to generate a preview with a lower threshold
// or falls back to evenly distributed frames if scene detection fails
func generateScenePreviewFallback(srcPath, outputPath string, opts *ScenePreviewOptions, prevOutput []byte, prevErr error) error {
	fmt.Printf("[DEBUG] Scene detection didn't create output, trying fallback. Previous error: %v\nOutput: %s\n", prevErr, string(prevOutput))

	// Try with a much lower threshold to capture any scene changes
	lowerThreshold := opts.SceneThreshold * 0.5
	if lowerThreshold < 0.1 {
		lowerThreshold = 0.1
	}

	filter := fmt.Sprintf(
		"select='gt(scene\\,%.2f)',scale=%d:%d,tile=%dx%d",
		lowerThreshold,
		opts.TileWidth,
		opts.TileHeight,
		opts.Cols,
		opts.Rows,
	)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", filter,
		"-frames:v", "1",
		"-y", outputPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create fallback command: %w", err)
	}

	fmt.Printf("[DEBUG] Trying lower threshold (%.2f)\n", lowerThreshold)
	output, _ := cmd.CombinedOutput()

	// Check if output was created
	if _, statErr := os.Stat(outputPath); statErr == nil {
		fmt.Printf("[DEBUG] Fallback scene preview succeeded with lower threshold\n")
		return nil
	}

	// Final fallback: use evenly distributed frames
	fmt.Printf("[DEBUG] Lower threshold also didn't create output: %s\n", string(output))
	return generateEvenlyDistributedPreview(srcPath, outputPath, opts)
}

// generateEvenlyDistributedPreview creates a preview with frames from evenly spaced intervals
// Used as a fallback when scene detection doesn't find enough scene changes
func generateEvenlyDistributedPreview(srcPath, outputPath string, opts *ScenePreviewOptions) error {
	fmt.Printf("[DEBUG] Falling back to evenly distributed frame preview\n")

	// Get video duration
	duration, err := getVideoDuration(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}

	durSec := duration.Seconds()
	totalFrames := opts.Cols * opts.Rows

	// Build select filter for evenly distributed frames
	// e.g., select='eq(n,100)+eq(n,200)+eq(n,300)+...'
	// We use time-based selection instead: between(t,X,X+0.1)
	var selectParts []string
	for i := 0; i < totalFrames; i++ {
		// Distribute frames across the video at even intervals
		// Skip first 5% and last 5% to avoid intros/outros
		progress := 0.05 + (0.9 * float64(i) / float64(totalFrames-1))
		timestamp := durSec * progress
		selectParts = append(selectParts, fmt.Sprintf("between(t\\,%.2f\\,%.2f)", timestamp, timestamp+0.1))
	}

	filter := fmt.Sprintf(
		"select='%s',scale=%d:%d,tile=%dx%d",
		strings.Join(selectParts, "+"),
		opts.TileWidth,
		opts.TileHeight,
		opts.Cols,
		opts.Rows,
	)

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-loglevel", "warning",
		"-i", srcPath,
		"-vf", filter,
		"-frames:v", "1",
		"-y", outputPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create evenly distributed preview command: %w", err)
	}

	fmt.Printf("[DEBUG] Running evenly distributed frame extraction\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("evenly distributed preview failed: %w\nOutput: %s", err, string(output))
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("preview not created after fallback: %s", outputPath)
	}

	fmt.Printf("[DEBUG] Successfully generated evenly distributed preview: %s\n", outputPath)
	return nil
}

// ScenePreviewOptions configures scene-based preview generation (Method 1)
type ScenePreviewOptions struct {
	SceneThreshold float64 // 0.0-1.0, scene detection sensitivity (higher = fewer scenes detected)
	TileWidth      int     // Width of each tile in pixels
	TileHeight     int     // Height of each tile in pixels
	Cols           int     // Number of columns in the tile grid
	Rows           int     // Number of rows in the tile grid
}

// DefaultScenePreviewOptions returns sensible defaults for scene-based preview
func DefaultScenePreviewOptions() *ScenePreviewOptions {
	return &ScenePreviewOptions{
		SceneThreshold: 0.4, // Moderate sensitivity - captures significant scene changes
		TileWidth:      160, // Width per tile
		TileHeight:     120, // Height per tile (4:3 aspect)
		Cols:           4,   // 4 columns
		Rows:           4,   // 4 rows = 16 scenes total
	}
}

// GetOutputSize returns the total output image dimensions
func (opts *ScenePreviewOptions) GetOutputSize() (width, height int) {
	return opts.TileWidth * opts.Cols, opts.TileHeight * opts.Rows
}

// generateSceneBasedGIFWithGPU creates a GIF using GPU acceleration
func generateSceneBasedGIFWithGPU(srcPath, gifPath string, timestamps []float64, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating scene-based GIF with GPU (%s)\n", opts.GPUType)

	switch opts.GPUType {
	case "cuda":
		return generateGIFWithCUDA(srcPath, gifPath, timestamps, opts)
	case "vaapi":
		return generateGIFWithVAAPI(srcPath, gifPath, timestamps, opts)
	default:
		return fmt.Errorf("unsupported GPU type: %s", opts.GPUType)
	}
}

// generateGIFWithCUDA uses NVIDIA GPU acceleration
func generateGIFWithCUDA(srcPath, gifPath string, timestamps []float64, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating GIF with CUDA acceleration\n")

	// CUDA pipeline: decode with cuvid, process on GPU, download for GIF encoding
	cols := 4
	rows := 2
	cellSize := opts.Width / cols

	var sceneFilters []string
	for i, ts := range timestamps {
		if i >= cols*rows {
			break
		}
		// Increased duration to 1.5 seconds
		sceneFilters = append(sceneFilters, fmt.Sprintf(
			"[0:v]trim=start=%.2f:duration=1.5,setpts=PTS-STARTPTS,fps=%d,hwupload_cuda,scale_cuda=%d:%d,hwdownload,format=yuv420p[v%d]",
			ts, opts.FPS, cellSize, cellSize, i,
		))
	}

	gridFilter := fmt.Sprintf("[v0][v1][v2][v3][v4][v5][v6][v7]xstack=inputs=8:layout=0_0|%d_0|%d_0|%d_0|0_%d|%d_%d|%d_%d|%d_%d",
		cellSize, cellSize*2, cellSize*3, cellSize, cellSize, cellSize*2, cellSize*3, cellSize)

	filterComplex := strings.Join(sceneFilters, ";") + ";" + gridFilter

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-i", srcPath,
		"-filter_complex", filterComplex,
		"-y", gifPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create CUDA command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CUDA GIF generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated GIF with CUDA\n")
	return nil
}

// generateGIFWithVAAPI uses Intel/AMD GPU acceleration
func generateGIFWithVAAPI(srcPath, gifPath string, timestamps []float64, opts PreviewOptions) error {
	fmt.Printf("[DEBUG] Generating GIF with VAAPI acceleration\n")

	cols := 4
	rows := 2
	cellSize := opts.Width / cols

	var sceneFilters []string
	for i, ts := range timestamps {
		if i >= cols*rows {
			break
		}
		// Increased duration to 1.5 seconds
		sceneFilters = append(sceneFilters, fmt.Sprintf(
			"[0:v]trim=start=%.2f:duration=1.5,setpts=PTS-STARTPTS,fps=%d,format=nv12,hwupload,scale_vaapi=w=%d:h=%d,hwdownload,format=nv12[v%d]",
			ts, opts.FPS, cellSize, cellSize, i,
		))
	}

	gridFilter := fmt.Sprintf("[v0][v1][v2][v3][v4][v5][v6][v7]xstack=inputs=8:layout=0_0|%d_0|%d_0|%d_0|0_%d|%d_%d|%d_%d|%d_%d",
		cellSize, cellSize*2, cellSize*3, cellSize, cellSize, cellSize*2, cellSize*3, cellSize)

	filterComplex := strings.Join(sceneFilters, ";") + ";" + gridFilter

	cmd, err := ffmpeg.NewFFmpegCommand(
		"-init_hw_device", "vaapi",
		"-hwaccel", "vaapi",
		"-i", srcPath,
		"-filter_complex", filterComplex,
		"-y", gifPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create VAAPI command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("VAAPI GIF generation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("[DEBUG] Successfully generated GIF with VAAPI\n")
	return nil
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
