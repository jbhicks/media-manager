package preview

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
	switch {
	case isImageFile(fileExt):
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
	case isVideoFile(fileExt):
		// No longer generate static thumbnails for videos
		return nil
	default:
		return fmt.Errorf("unsupported file type: %s", fileExt)
	}
}

func isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	return slices.Contains(imageExts, ext)
}

func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v", ".3gp"}
	return slices.Contains(videoExts, ext)
}

func generateImageThumbnail(srcPath, thumbPath string) error {
	// Open source image
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Get file info to verify it's a valid file
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("image file is empty")
	}

	fmt.Printf("[DEBUG] Image file size: %d bytes\n", fileInfo.Size())

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		// Try to debug what went wrong
		file.Seek(0, 0) // Reset to beginning of file
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		fmt.Printf("[DEBUG] Image decode failed. First %d bytes: %v\n", n, buffer[:n])
		return fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}
	fmt.Printf("[DEBUG] Successfully decoded image format: %s\n", format)

	// Get the original dimensions
	bounds := img.Bounds()
	fmt.Printf("[DEBUG] Original image dimensions: %dx%d\n", bounds.Dx(), bounds.Dy())

	// Target size for the thumbnail
	targetWidth, targetHeight := uint(180), uint(180)

	// Calculate scaling factor to cover the target dimensions
	originalWidth, originalHeight := uint(bounds.Dx()), uint(bounds.Dy())
	scaleX := float64(targetWidth) / float64(originalWidth)
	scaleY := float64(targetHeight) / float64(originalHeight)

	var scaledImg image.Image
	if scaleX > scaleY { // Original is taller than target aspect ratio, scale by width
		scaledImg = resize.Resize(targetWidth, 0, img, resize.Lanczos3)
	} else { // Original is wider than target aspect ratio, scale by height
		scaledImg = resize.Resize(0, targetHeight, img, resize.Lanczos3)
	}

	// Calculate crop rectangle
	scaledBounds := scaledImg.Bounds()
	cropX := (scaledBounds.Dx() - int(targetWidth)) / 2
	cropY := (scaledBounds.Dy() - int(targetHeight)) / 2

	// Crop the scaled image using SubImage
	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	subImager, ok := scaledImg.(SubImager)
	if !ok {
		return fmt.Errorf("image does not support SubImage interface")
	}

	resizedImg := subImager.SubImage(image.Rect(cropX, cropY, cropX+int(targetWidth), cropY+int(targetHeight)))

	// Save thumbnail
	outFile, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 85})
	if err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	fmt.Printf("[DEBUG] Successfully created thumbnail at: %s\n", thumbPath)
	return nil
}

func generateVideoThumbnail(srcPath, thumbPath string) error {
	// Use FFmpeg to extract a frame from the video with uniform dimensions
	fmt.Printf("[DEBUG] Running ffmpeg command: ffmpeg -i %s -ss 00:00:01 -vframes 1 -vf scale=180:180:force_original_aspect_ratio=increase,crop=180:180 -y %s\n", srcPath, thumbPath)
	fmt.Printf("[DEBUG] Source file exists: %v\n", fileExists(srcPath))
	fmt.Printf("[DEBUG] Thumbnail path writable: %v\n", pathWritable(thumbPath))
	cmd := exec.Command("ffmpeg",
		"-loglevel", "warning",
		"-i", srcPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1", // Extract only 1 frame
		"-vf", "scale=180:101:force_original_aspect_ratio=increase,crop=180:101", // Scale and crop to 180x101
		"-f", "image2", // Explicitly set output format to image2 to avoid sequence pattern errors
		"-y", // Overwrite output file
		thumbPath,
	)
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
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

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
	fmt.Printf("[DEBUG] Generating scene-overview animated GIF preview for: %s\n", srcPath)

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

	// Decide on number of segments and segment length
	numSegments := 10
	segLen := 1.0 // seconds
	if durSec < float64(numSegments+1) {
		numSegments = int(durSec) - 1
		if numSegments < 1 {
			numSegments = 1
		}
	}
	interval := durSec / float64(numSegments+1)

	// Build filtergraph for ffmpeg
	var filterParts []string
	var concatInputs []string
	for i := 0; i < numSegments; i++ {
		start := interval * float64(i+1)
		end := start + segLen
		filterParts = append(filterParts,
			fmt.Sprintf("[0:v]trim=start=%.3f:end=%.3f,setpts=PTS-STARTPTS[v%d]", start, end, i))
		concatInputs = append(concatInputs, fmt.Sprintf("[v%d]", i))
	}
	filterParts = append(filterParts,
		fmt.Sprintf("%sconcat=n=%d:v=1:a=0,scale=180:101:force_original_aspect_ratio=increase,crop=180:101,fps=12[outv]",
			strings.Join(concatInputs, ""), numSegments))
	filtergraph := strings.Join(filterParts, ";")

	cmd := exec.Command("ffmpeg",
		"-loglevel", "warning",
		"-i", srcPath,
		"-filter_complex", filtergraph,
		"-map", "[outv]",
		"-y",
		gifPath,
	)

	fmt.Printf("[DEBUG] Running ffmpeg for scene-overview GIF: %v\n", cmd.Args)
	fmt.Printf("[DEBUG] ffmpeg filtergraph: %s\n", filtergraph)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[ERROR] ffmpeg error: %v\n[ffmpeg output]: %s\n", err, string(output))
		return fmt.Errorf("failed to generate animated preview: %v, output: %s\n", err, string(output))
	}
	if len(output) > 0 {
		fmt.Printf("[ffmpeg warning]: %s\n", string(output))
	}
	// Verify the GIF was created
	if _, err := os.Stat(gifPath); err != nil {
		return fmt.Errorf("animated preview file missing after generation: %s", gifPath)
	}

	fmt.Printf("[DEBUG] Successfully generated scene-overview animated preview: %s\n", gifPath)
	return nil
}

// GetFFmpegHardwareAccelerations returns a list of supported hardware accelerations by ffmpeg.
func GetFFmpegHardwareAccelerations() ([]string, error) {
	cmd := exec.Command("ffmpeg", "-hwaccels")
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
	cmd := exec.Command("ffmpeg",
		"-loglevel", "warning",
		"-i", gifPath,
		"-vsync", "0", // Ensure all frames are extracted
		"-vf", "fps=8", // Force 8 frames per second
		"-frame_pts", "1", // Add presentation timestamp to frame filename
		"-f", "image2", // Force image2 format
		"-qscale:v", "2", // High quality jpeg output
		outputPattern,
	)

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

	cmd := exec.Command("ffmpeg", cmdArgs...)

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
