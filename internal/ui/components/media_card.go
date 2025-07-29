package components

import (
	"bytes"
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

// logToDebugFile appends a message to app_debug.log for error tracking
func logToDebugFile(msg string) {
	f, err := os.OpenFile("app_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Could not open app_debug.log: %v\n", err)
		return
	}
	_, _ = f.WriteString(fmt.Sprintf("%s\n", msg))
	f.Close()
}

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
	if _, err := os.Stat(mc.filePath); os.IsNotExist(err) {
		fmt.Printf("[WARN] File does not exist, skipping image thumbnail: %s\n", mc.filePath)
		return
	}
	fmt.Printf("[DEBUG] Generating image thumbnail for: %s\n", mc.filePath)
	fmt.Printf("[DEBUG] generateImageThumbnail called for %s\n", mc.fileName)
	ext := strings.ToLower(filepath.Ext(mc.filePath))
	if ext == ".gif" {
		// For GIFs, just use the original file as a still image
		if file, err := os.Open(mc.filePath); err == nil {
			defer file.Close()
			if _, _, err := image.DecodeConfig(file); err == nil {
				img := canvas.NewImageFromFile(mc.filePath)
				img.FillMode = canvas.ImageFillContain
				mc.content = img
				mc.Refresh()
				return
			} else {
				logToDebugFile(fmt.Sprintf("[ERROR] Invalid GIF image: %s, error: %v", mc.filePath, err))
				mc.content = widget.NewIcon(theme.FileImageIcon())
				mc.Refresh()
				return
			}
		} else {
			logToDebugFile(fmt.Sprintf("[ERROR] Could not open GIF file: %s, error: %v", mc.filePath, err))
			mc.content = widget.NewIcon(theme.FileImageIcon())
			mc.Refresh()
			return
		}
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
		var stderr bytes.Buffer
		cmd := exec.Command("ffmpeg", "-i", mc.filePath, "-vf", "scale=180:180:force_original_aspect_ratio=increase,crop=180:180", "-frames:v", "1", thumbPath)
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logToDebugFile(fmt.Sprintf("[ERROR] ffmpeg failed to generate thumbnail for %s: %v\n[ffmpeg stderr]: %s\n", mc.filePath, err, stderr.String()))
			mc.content = widget.NewIcon(theme.FileImageIcon())
			mc.Refresh()
			return
		}
	}
	if file, err := os.Open(thumbPath); err == nil {
		defer file.Close()
		if _, _, err := image.DecodeConfig(file); err == nil {
			img := canvas.NewImageFromFile(thumbPath)
			img.FillMode = canvas.ImageFillContain
			mc.content = img
			mc.Refresh()
			return
		} else {
			logToDebugFile(fmt.Sprintf("[ERROR] Invalid thumbnail image: %s, error: %v", thumbPath, err))
			mc.content = widget.NewIcon(theme.FileImageIcon())
			mc.Refresh()
			return
		}
	} else {
		logToDebugFile(fmt.Sprintf("[ERROR] Could not open thumbnail file: %s, error: %v", thumbPath, err))
		mc.content = widget.NewIcon(theme.FileImageIcon())
		mc.Refresh()
		return
	}

}

func (mc *MediaCard) generateGifPreview() {
	if _, err := os.Stat(mc.filePath); os.IsNotExist(err) {
		fmt.Printf("[WARN] File does not exist, skipping GIF preview: %s\n", mc.filePath)
		return
	}
	fmt.Printf("[DEBUG] Generating animated GIF preview for: %s\n", mc.filePath)
	fmt.Printf("[DEBUG] generateGifPreview called for %s\n", mc.fileName)
	if mc.mediaType != MediaTypeVideo {
		return
	}

	homeDir, _ := os.UserHomeDir()
	gifDir := filepath.Join(homeDir, ".media-manager", "previews")
	os.MkdirAll(gifDir, 0755)
	gifPath := filepath.Join(gifDir, strings.TrimSuffix(filepath.Base(mc.filePath), filepath.Ext(mc.filePath))+".gif")

	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		var stderr bytes.Buffer
		cmd := exec.Command("ffmpeg",
			"-i", mc.filePath,
			"-vf", "fps=12,scale=180:180:force_original_aspect_ratio=increase,crop=180:180",
			"-frames:v", "24",
			gifPath)
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logToDebugFile(fmt.Sprintf("[ERROR] ffmpeg failed to generate GIF for %s: %v\n[ffmpeg stderr]: %s\n", mc.filePath, err, stderr.String()))
			mc.content = widget.NewIcon(theme.FileImageIcon())
			mc.Refresh()
			return
		}
	}

	// Validate GIF before loading
	if file, err := os.Open(gifPath); err == nil {
		defer file.Close()
		if _, err := gif.DecodeConfig(file); err == nil {
			uri := storage.NewFileURI(gifPath)
			animatedGif, err := xwidget.NewAnimatedGif(uri)
			if err == nil {
				animatedGif.Stop() // Show first frame only
				mc.animatedGif = animatedGif
				mc.content = animatedGif
				mc.hasAnimation = true
				mc.Refresh()
				return
			} else {
				// fallback: use static image
				img := canvas.NewImageFromFile(gifPath)
				img.FillMode = canvas.ImageFillContain
				mc.content = img
				mc.Refresh()
				return
			}
		} else {
			logToDebugFile(fmt.Sprintf("[ERROR] Invalid GIF preview: %s, error: %v", gifPath, err))
			mc.content = widget.NewIcon(theme.FileImageIcon())
			mc.Refresh()
			return
		}
	} else {
		logToDebugFile(fmt.Sprintf("[ERROR] Could not open GIF preview file: %s, error: %v", gifPath, err))
		mc.content = widget.NewIcon(theme.FileImageIcon())
		mc.Refresh()
		return
	}
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
	padding := float32(8)
	// Use the grid cell size for everything
	w, h := size.Width, size.Height

	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	contentH := h - 40 // leave space for label
	contentW := w - 2*padding
	if r.content != nil {
		r.content.Resize(fyne.NewSize(contentW, contentH))
		r.content.Move(fyne.NewPos(padding, padding))
	}

	labelHeight := float32(32)
	labelY := h - labelHeight - padding
	labelWidth := w - 2*padding

	r.labelBackground.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.labelBackground.Move(fyne.NewPos(padding, labelY))

	r.label.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.label.Move(fyne.NewPos(padding, labelY))
}

func (r *mediaCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(180, 180)
}

func (r *mediaCardRenderer) Refresh() {
	fyne.Do(func() {
		r.content = r.card.content
		canvas.Refresh(r.background)
		if r.content != nil {
			canvas.Refresh(r.content)
		}
		canvas.Refresh(r.labelBackground)
		canvas.Refresh(r.label)
		r.Layout(r.background.Size())
	})
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
