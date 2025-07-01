package preview

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	filePath = filepath.Join("/home/josh/media-manager", filePath)
	thumbPath = filepath.Clean(thumbPath)
	fmt.Printf("[DEBUG] Generating thumbnail for: %s\n", filePath)
	fmt.Printf("[DEBUG] Output path: %s\n", thumbPath)
	// Determine file type
	ext := strings.ToLower(filepath.Ext(filePath))

	if isImageFile(ext) {
		return generateImageThumbnail(filePath, thumbPath)
	} else if isVideoFile(ext) {
		return generateVideoThumbnail(filePath, thumbPath)
	} else {
		return fmt.Errorf("unsupported file type: %s", ext)
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

	// Save thumbnail
	outFile, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return nil
}

func generateVideoThumbnail(srcPath, thumbPath string) error {
	// Use FFmpeg to extract a frame from the video
	fmt.Printf("[DEBUG] Running ffmpeg command: ffmpeg -i %s -ss 00:00:01 -vframes 1 -update 1 -y %s\n", srcPath, thumbPath)
	fmt.Printf("[DEBUG] Source file exists: %v\n", fileExists(srcPath))
	fmt.Printf("[DEBUG] Thumbnail path writable: %v\n", pathWritable(thumbPath))
	cmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1", "-update", "1", // Extract only 1 frame
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

// GenerateAnimatedPreview creates a single animated GIF for video preview
func GenerateAnimatedPreview(srcPath, gifPath string) error {
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

	// Use simpler FFmpeg command for better compatibility
	fmt.Printf("[DEBUG] Running simplified ffmpeg command for animated preview: %s\n", gifPath)
	cmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-ss", "2", // Start at 2 seconds to skip intro
		"-t", "3", // 3 second duration
		"-vf", "fps=8,scale=200:-1:flags=lanczos", // 8 FPS, 200px width
		"-f", "gif", // Force GIF format
		"-y", // Overwrite existing
		gifPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[DEBUG] Failed to generate animated preview: %v, output: %s\n", err, string(output))
		return fmt.Errorf("failed to generate animated preview: %w", err)
	}

	// Verify the GIF was created
	if _, err := os.Stat(gifPath); err != nil {
		return fmt.Errorf("animated preview file missing after generation: %s", gifPath)
	}

	fmt.Printf("[DEBUG] Successfully generated animated preview: %s\n", gifPath)
	return nil
}
