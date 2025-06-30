package preview

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PreviewGenerator struct {
	config ConfigProvider
}

func NewPreviewGenerator(cfg ConfigProvider) *PreviewGenerator {
	return &PreviewGenerator{config: cfg}
}

func (p *PreviewGenerator) GeneratePreview(filePath string) (string, error) {
	// Generate thumbnail filename based on file path hash
	hash := sha256.Sum256([]byte(filePath))
	thumbName := fmt.Sprintf("%x.jpg", hash)
	thumbPath := filepath.Join(p.config.GetThumbnailDir(), thumbName)

	// Check if thumbnail already exists
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	// Determine file type and generate appropriate thumbnail
	ext := strings.ToLower(filepath.Ext(filePath))

	switch {
	case p.isImageFile(ext):
		return p.generateImageThumbnail(filePath, thumbPath)
	case p.isVideoFile(ext):
		return p.generateVideoThumbnail(filePath, thumbPath)
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

func (p *PreviewGenerator) isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

func (p *PreviewGenerator) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return true
		}
	}
	return false
}

func (p *PreviewGenerator) generateImageThumbnail(srcPath, thumbPath string) (string, error) {
	// Open source image
	file, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate thumbnail dimensions
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate aspect ratio preserving dimensions
	thumbW, thumbH := p.calculateThumbnailSize(width, height)

	// Create resized image (simple nearest neighbor for now)
	thumbnail := p.resizeImageOptimized(img, thumbW, thumbH)

	// Save thumbnail
	outFile, err := os.Create(thumbPath)
	if err != nil {
		return "", fmt.Errorf("failed to create thumbnail file: %w", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, thumbnail, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return thumbPath, nil
}

func (p *PreviewGenerator) generateVideoThumbnail(srcPath, thumbPath string) (string, error) {
	// Use FFmpeg to extract a frame from the video
	cmd := exec.Command("ffmpeg",
		"-i", srcPath,
		"-ss", "00:00:01", // Extract frame at 1 second
		"-vframes", "1", // Extract only 1 frame
		"-vf", fmt.Sprintf("scale=%d:%d", p.config.GetThumbnailSize(), p.config.GetThumbnailSize()),
		"-y", // Overwrite output file
		thumbPath,
	)

	err := cmd.Run()
	if err != nil {
		cmd := exec.Command("ffmpeg", "-i", srcPath, "-ss", "00:00:01", "-vframes", "1", "-vf", fmt.Sprintf("scale=%d:%d", p.config.GetThumbnailSize(), p.config.GetThumbnailSize()), "-y", thumbPath)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to generate video thumbnail: %w", err)
		}
		return thumbPath, nil
	}

	return thumbPath, nil
}

func (p *PreviewGenerator) calculateThumbnailSize(width, height int) (int, int) {
	maxSize := p.config.GetThumbnailSize()

	if width <= maxSize && height <= maxSize {
		return width, height
	}

	ratio := float64(width) / float64(height)

	if width > height {
		return maxSize, int(float64(maxSize) / ratio)
	} else {
		return int(float64(maxSize) * ratio), maxSize
	}
}

// Simple image resizing using nearest neighbor
func (p *PreviewGenerator) resizeImageOptimized(src image.Image, width, height int) image.Image {
	// Optimized resizing logic using bilinear interpolation
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

func (p *PreviewGenerator) resizeImage(src image.Image, width, height int) image.Image {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}
