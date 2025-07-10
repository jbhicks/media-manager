package components

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

type MediaType int

const (
	MediaTypeImage MediaType = iota
	MediaTypeVideo
	MediaTypeFile
)

// MediaCard represents a uniform card for all media types
type MediaCard struct {
	widget.BaseWidget
	mediaType       MediaType
	filePath        string
	fileName        string
	thumbnailPath   string
	animatedGif     *xwidget.AnimatedGif // fyne-x GIF widget for animated previews
	icon            *widget.Icon
	label           *widget.Label
	labelBackground fyne.CanvasObject
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	isHovered       bool
	hasAnimation    bool
	onDelete        func()
	previewWidth    int
	previewHeight   int
}

func NewMediaCard(filePath, fileName string, mediaType MediaType, thumbPath string) *MediaCard {
	fmt.Printf("[DEBUG] NewMediaCard: Creating card for %s (Type: %v)\n", fileName, mediaType)
	displayName := fileName
	if len(displayName) > 22 {
		displayName = displayName[:19] + "..."
	}

	card := &MediaCard{
		mediaType:     mediaType,
		filePath:      filePath,
		fileName:      fileName,
		thumbnailPath: thumbPath,
		isHovered:     false,
		hasAnimation:  false,
	}

	card.setupContent()
	card.label = widget.NewLabelWithStyle(displayName, fyne.TextAlignCenter, fyne.TextStyle{})
	card.label.Wrapping = fyne.TextWrapWord
	card.background = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	card.background.StrokeColor = color.NRGBA{100, 100, 100, 255}
	card.background.StrokeWidth = 1

	card.labelBackground = canvas.NewLinearGradient(
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 180},
		90,
	)
	card.ExtendBaseWidget(card)

	// [DEBUG] NewMediaCard: Card created for %s. hasAnimation: %v, animatedGif: %v\n", fileName, card.hasAnimation, card.animatedGif != nil)
	return card
}

func (mc *MediaCard) setupContent() {
	fmt.Printf("[DEBUG] setupContent: Setting up content for %s (Type: %v)\n", mc.fileName, mc.mediaType)

	switch mc.mediaType {
	case MediaTypeImage:
		go mc.generateImageThumbnail()
	case MediaTypeVideo:
		go mc.generateGifPreview()
	case MediaTypeFile:
		mc.content = widget.NewIcon(theme.FileIcon())
	}
}

// generateImageThumbnail generates a still thumbnail for images (not GIFs), or uses the original for GIFs
func (mc *MediaCard) generateImageThumbnail() {
	fmt.Printf("[DEBUG] generateImageThumbnail called for %s\n", mc.fileName)
	ext := strings.ToLower(filepath.Ext(mc.filePath))
	if ext == ".gif" {
		// For GIFs, just use the original file as a still image
		img := canvas.NewImageFromFile(mc.filePath)
		img.FillMode = canvas.ImageFillContain
		mc.content = img
		mc.Refresh()
		return
	}
	// For other images, generate a thumbnail (jpg)
	homeDir, _ := os.UserHomeDir()
	thumbDir := filepath.Join(homeDir, ".media-manager", "thumbnails")
	os.MkdirAll(thumbDir, 0755)
	thumbFileName := strings.ReplaceAll(strings.TrimSuffix(filepath.Base(mc.filePath), filepath.Ext(mc.filePath)), " ", "_") + "_thumb.jpg"
	thumbPath := filepath.Join(thumbDir, thumbFileName)
	// Only generate if not exists
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		// Use ffmpeg to generate a thumbnail for any image type
		cmd := exec.Command("ffmpeg", "-i", mc.filePath, "-vf", "scale=180:-1:force_original_aspect_ratio=decrease", "-frames:v", "1", thumbPath)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate image thumbnail for %s: %v\n", mc.filePath, err)
			return
		}
	}
	img := canvas.NewImageFromFile(thumbPath)
	img.FillMode = canvas.ImageFillContain
	// Try to get image dimensions
	if file, err := os.Open(thumbPath); err == nil {
		defer file.Close()
		if srcImg, _, err := image.DecodeConfig(file); err == nil {
			mc.previewWidth = srcImg.Width
			mc.previewHeight = srcImg.Height
		}
	}
	mc.content = img
	mc.Refresh()
}

func (mc *MediaCard) generateGifPreview() {
	fmt.Printf("[DEBUG] generateGifPreview called for %s\n", mc.fileName)
	if mc.mediaType != MediaTypeVideo {
		return
	}

	// Only generate and use GIF for videos
	homeDir, _ := os.UserHomeDir()
	gifDir := filepath.Join(homeDir, ".media-manager", "previews")
	os.MkdirAll(gifDir, 0755)

	gifPath := filepath.Join(gifDir, strings.TrimSuffix(filepath.Base(mc.filePath), filepath.Ext(mc.filePath))+".gif")

	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		cmd := exec.Command("ffmpeg",
			"-i", mc.filePath,
			"-vf", "fps=12,scale=180:-1:force_original_aspect_ratio=decrease",
			"-frames:v", "24",
			gifPath)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate GIF for %s: %v\n", mc.filePath, err)
			return
		}
	}

	uri := storage.NewFileURI(gifPath)
	animatedGif, err := xwidget.NewAnimatedGif(uri)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create AnimatedGif widget for %s: %v\n", gifPath, err)
		return
	}
	// Try to get GIF dimensions
	if file, err := os.Open(gifPath); err == nil {
		defer file.Close()
		if cfg, err := gif.DecodeConfig(file); err == nil {
			mc.previewWidth = cfg.Width
			mc.previewHeight = cfg.Height
		}
	}
	animatedGif.Stop() // Show first frame only
	fyne.Do(func() {
		mc.animatedGif = animatedGif
		mc.content = animatedGif
		mc.hasAnimation = true
		mc.Refresh()
	})
}

var _ desktop.Hoverable = (*MediaCard)(nil)

func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] MediaCard MouseIn - hover started")
	mc.isHovered = true
	mc.background.FillColor = theme.Color(theme.ColorNameHover)
	mc.background.Refresh()

	fmt.Printf("[DEBUG] MouseIn: mediaType=%v, hasAnimation=%v, mc.content=%v\n", mc.mediaType, mc.hasAnimation, mc.content)
	if mc.hasAnimation && mc.animatedGif != nil {
		mc.content = mc.animatedGif
		mc.animatedGif.Start()
		mc.Refresh()
	}
}

func (mc *MediaCard) MouseOut() {
	fmt.Println("[DEBUG] MediaCard MouseOut - hover ended")
	mc.isHovered = false
	mc.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	mc.background.Refresh()

	if mc.hasAnimation && mc.animatedGif != nil {
		mc.animatedGif.Stop()
	}

	fmt.Printf("[DEBUG] MouseOut: mediaType=%v, hasAnimation=%v, animatedGif=%v\n", mc.mediaType, mc.hasAnimation, mc.animatedGif != nil)
}

func (mc *MediaCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

func (mc *MediaCard) Tapped(*fyne.PointEvent) {
	// [DEBUG] MediaCard Tapped: %s\n", mc.filePath)
	err := mc.openFile()
	if err != nil {
		// [DEBUG] Error opening file: %v\n", err)
	}
}

func (mc *MediaCard) TappedSecondary(e *fyne.PointEvent) {
	// [DEBUG] MediaCard TappedSecondary: %s\n", mc.filePath)
	deleteMenuItem := fyne.NewMenuItem("Delete", func() {
		err := os.Remove(mc.filePath)
		if err != nil {
			fmt.Printf("[ERROR] Failed to delete file: %v\n", err)
			return
		}
		fmt.Printf("[INFO] Deleted file: %s\n", mc.filePath)
		if mc.onDelete != nil {
			mc.onDelete()
		}
	})
	canvas := fyne.CurrentApp().Driver().CanvasForObject(mc)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", deleteMenuItem), canvas, e.AbsolutePosition)
}

func (mc *MediaCard) SetOnDelete(callback func()) {
	mc.onDelete = callback
}

func (mc *MediaCard) openFile() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", mc.filePath)
	case "darwin": // macOS
		cmd = exec.Command("open", mc.filePath)
	default: // Linux and others
		cmd = exec.Command("xdg-open", mc.filePath)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("[DEBUG] Command finished with error: %v\n", err)
		}
	}()
	return nil
}

func (mc *MediaCard) MinSize() fyne.Size {
	return fyne.NewSize(180, 101)
}

func (mc *MediaCard) CreateRenderer() fyne.WidgetRenderer {
	return &mediaCardRenderer{
		card:            mc,
		background:      mc.background,
		content:         mc.content,
		labelBackground: mc.labelBackground,
		label:           mc.label,
	}
}

type mediaCardRenderer struct {
	card            *MediaCard
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	labelBackground fyne.CanvasObject
	label           *widget.Label
}

func (r *mediaCardRenderer) Layout(size fyne.Size) {

	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	labelAreaHeight := r.label.MinSize().Height + float32(8)
	contentAvailableHeight := size.Height - labelAreaHeight
	contentAvailableWidth := size.Width

	// Dynamically determine content height based on preview aspect ratio
	contentHeight := contentAvailableHeight
	if r.card.previewWidth > 0 && r.card.previewHeight > 0 {
		aspect := float32(r.card.previewHeight) / float32(r.card.previewWidth)
		contentHeight = contentAvailableWidth * aspect
		if contentHeight > contentAvailableHeight {
			contentHeight = contentAvailableHeight
		}
	} else {
		// Fallback to 16:9 aspect ratio
		contentHeight = contentAvailableWidth * 9.0 / 16.0
	}
	contentSize := fyne.NewSize(contentAvailableWidth, contentHeight)

	contentX := float32(0)
	contentY := float32(0)

	if r.content != nil {
		r.content.Resize(contentSize)
		r.content.Move(fyne.NewPos(contentX, contentY))
	}

	canvas.Refresh(r.content)

	labelMinHeight := float32(44) // Increased for more title space
	labelWidth := size.Width - float32(8)
	labelX := float32(4)
	labelHeight := labelMinHeight
	labelY := size.Height - labelHeight - float32(4)

	r.labelBackground.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.labelBackground.Move(fyne.NewPos(labelX, labelY))

	r.label.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.label.Move(fyne.NewPos(labelX, labelY))
}

func (r *mediaCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(180, 101)
}

func (r *mediaCardRenderer) Refresh() {
	r.content = r.card.content

	canvas.Refresh(r.background)
	if r.content != nil {
		canvas.Refresh(r.content)
	}
	canvas.Refresh(r.labelBackground)
	canvas.Refresh(r.label)

	r.Layout(r.background.Size())
}

func (r *mediaCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.content, r.labelBackground, r.label}
}

func (r *mediaCardRenderer) Destroy() {
	if r.card.animatedGif != nil {
		r.card.animatedGif.Stop()
	}
}

func GetMediaType(filename string) MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return MediaTypeImage
	case ".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv", ".ts", ".3gp":
		return MediaTypeVideo
	default:
		return MediaTypeFile
	}
}
