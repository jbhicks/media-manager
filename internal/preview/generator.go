package preview

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

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
	// Check if the source file exists
	fmt.Printf("[DEBUG] Generating thumbnail for: %s\n", filePath)
	fmt.Printf("[DEBUG] Output path: %s\n", thumbPath)

	// Ensure the thumbnail directory exists
	if err := os.MkdirAll(filepath.Dir(thumbPath), 0755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Check if the source file exists before attempting to generate a thumbnail
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", filePath)
	}

	fileExt := strings.ToLower(filepath.Ext(filePath))
	switch {
	case isImageFile(fileExt):
		return generateImageThumbnail(filePath, thumbPath)
	case isVideoFile(fileExt):
		return generateVideoThumbnail(filePath, thumbPath)
	default:
		return fmt.Errorf("unsupported file type: %s", fileExt)
	}
}

func isImageFile(ext string) bool {

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return true
		}
	}
	return false
}

func generateImageThumbnail(srcPath, thumbPath string) error {
	// Open source image
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image to a fixed width (200px) while maintaining aspect ratio
	newWidth := uint(200)
	resizedImg := resize.Resize(newWidth, 0, img, resize.Lanczos3)

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

	return nil
}

func generateVideoThumbnail(srcPath, thumbPath string) error {
	// Use FFmpeg to extract a frame from the video with uniform dimensions
	fmt.Printf("[DEBUG] Running ffmpeg command with fixed dimensions: ffmpeg -i %s -ss 00:00:01 -vframes 1 -vf scale=200:200:force_original_aspect_ratio=decrease,pad=200:200:(ow-iw)/2:(oh-ih)/2 -y %s\n", srcPath, thumbPath)
	fmt.Printf("[DEBUG] Source file exists: %v\n", fileExists(srcPath))
	fmt.Printf("[DEBUG] Thumbnail path writable: %v\n", pathWritable(thumbPath))
	cmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1", // Extract only 1 frame
		"-vf", "scale=180:-1", // Scale to 180px wide, preserve aspect ratio
		"-y", // Overwrite output file
		thumbPath,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate video thumbnail: %w", err)
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
	fmt.Printf("[DEBUG] Generating animated GIF preview for: %s\n", srcPath)

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

	// Calculate interval for 10 evenly distributed frames
	numFrames := 10
	frameInterval := int(duration.Seconds() * 25 / float64(numFrames)) // Assuming 25fps
	filterComplex := fmt.Sprintf("select='not(mod(n,%d))',setpts=N/FRAME_RATE/TB,fps=8,scale=200:-1", frameInterval)

	cmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-vf", filterComplex,
		"-f", "gif", // Force GIF format
		"-y", // Overwrite existing
		gifPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate animated preview: %v, output: %s\n", err, string(output))
	}

	// Verify the GIF was created
	if _, err := os.Stat(gifPath); err != nil {
		return fmt.Errorf("animated preview file missing after generation: %s", gifPath)
	}

	fmt.Printf("[DEBUG] Successfully generated animated preview: %s\n", gifPath)
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
	filterComplex := fmt.Sprintf("select='%s',setpts=N/FRAME_RATE/TB,fps=8,scale=200:200:force_original_aspect_ratio=increase,crop=200:200", strings.Join(selectFilters, "+"))

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
