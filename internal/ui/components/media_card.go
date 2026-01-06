package components

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"

	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/pkg/models"
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
	forceRegenerate bool
	icon            *widget.Icon
	label           *widget.Label
	labelBackground fyne.CanvasObject
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	isHovered       bool
	hasAnimation    bool
	onDelete        func()
	thumbDir        string
	duration        int
	extension       string
	durationLabel   *widget.Label
	extensionLabel  *widget.Label
}

func NewMediaCard(file models.MediaFile, thumbDir string) *MediaCard {
	return NewMediaCardWithForce(file, thumbDir, false)
}

func NewMediaCardWithForce(file models.MediaFile, thumbDir string, forceRegenerate bool) *MediaCard {
	mediaType := GetMediaType(file.Filename)
	fmt.Printf("[DEBUG] NewMediaCard: Creating card for %s (Type: %v, Force: %v)\n", file.Filename, mediaType, forceRegenerate)
	displayName := file.Filename
	if len(displayName) > 22 {
		displayName = displayName[:19] + "..."
	}

	card := &MediaCard{
		mediaType:       mediaType,
		filePath:        file.Path,
		fileName:        file.Filename,
		thumbnailPath:   file.PreviewPath,
		thumbDir:        thumbDir,
		duration:        file.Duration,
		extension:       strings.ToUpper(strings.TrimPrefix(filepath.Ext(file.Filename), ".")),
		isHovered:       false,
		hasAnimation:    false,
		forceRegenerate: forceRegenerate,
	}

	card.setupContent()

	// Setup overlays
	if card.duration > 0 {
		mins := card.duration / 60
		secs := card.duration % 60
		card.durationLabel = widget.NewLabelWithStyle(fmt.Sprintf("%d:%02d", mins, secs), fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	}
	card.extensionLabel = widget.NewLabelWithStyle(card.extension, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})

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
		go mc.loadPreview("image")
	case MediaTypeVideo:
		go mc.loadPreview("video")
	case MediaTypeFile:
		mc.content = widget.NewIcon(theme.FileIcon())
	}
}

func (mc *MediaCard) loadPreview(mediaTypeStr string) {
	// If we already have a path from DB and not forcing regenerate, use it initially
	if mc.thumbnailPath != "" && !mc.forceRegenerate {
		if _, err := os.Stat(mc.thumbnailPath); err == nil {
			mc.updateContent(mc.thumbnailPath)
			return
		}
	}

	// Request from centralized manager
	path, err := preview.GetPreviewWithForce(mc.filePath, mediaTypeStr, mc.thumbDir, mc.forceRegenerate)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get preview for %s: %v\n", mc.filePath, err)
		return
	}

	// Wait for the file to exist (the worker pool will create it)
	// For now, we'll just check existence periodically or wait until Refresh is called.
	// A better way would be a callback, but let's start simple.
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		if _, err := os.Stat(path); err == nil {
			mc.thumbnailPath = path
			mc.updateContent(path)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (mc *MediaCard) updateContent(path string) {
	fyne.Do(func() {
		if mc.mediaType == MediaTypeVideo {
			uri := storage.NewFileURI(path)
			animatedGif, err := xwidget.NewAnimatedGif(uri)
			if err == nil {
				animatedGif.Stop()
				mc.animatedGif = animatedGif
				mc.content = animatedGif
				mc.hasAnimation = true
			} else {
				img := canvas.NewImageFromFile(path)
				img.FillMode = canvas.ImageFillContain
				mc.content = img
			}
		} else {
			img := canvas.NewImageFromFile(path)
			img.FillMode = canvas.ImageFillContain
			mc.content = img
		}
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
		cmd = exec.Command("cmd", "/C", "start", "", mc.filePath)
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
	return fyne.NewSize(216, 192)
}

func (mc *MediaCard) CreateRenderer() fyne.WidgetRenderer {
	return &mediaCardRenderer{
		card:            mc,
		background:      mc.background,
		content:         mc.content,
		labelBackground: mc.labelBackground,
		label:           mc.label,
		durationLabel:   mc.durationLabel,
		extensionLabel:  mc.extensionLabel,
	}
}

type mediaCardRenderer struct {
	card            *MediaCard
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	labelBackground fyne.CanvasObject
	label           *widget.Label
	durationLabel   *widget.Label
	extensionLabel  *widget.Label
}

func (r *mediaCardRenderer) Layout(size fyne.Size) {
	padding := float32(8)
	// Use the grid cell size for everything
	w, h := size.Width, size.Height

	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	contentH := h - 48 // leave space for label (20% bigger)
	contentW := w - 2*padding
	if r.content != nil {
		r.content.Resize(fyne.NewSize(contentW, contentH))
		r.content.Move(fyne.NewPos(padding, padding))
	}
	labelHeight := float32(38) // 20% bigger
	labelY := h - labelHeight - padding
	labelWidth := w - 2*padding

	r.labelBackground.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.labelBackground.Move(fyne.NewPos(padding, labelY))

	r.label.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.label.Move(fyne.NewPos(padding, labelY))

	// Duration bottom right of content
	if r.durationLabel != nil {
		ds := r.durationLabel.MinSize()
		r.durationLabel.Resize(ds)
		r.durationLabel.Move(fyne.NewPos(w-padding-ds.Width-4, contentH-ds.Height-4))
	}
	// Extension top right
	if r.extensionLabel != nil {
		es := r.extensionLabel.MinSize()
		r.extensionLabel.Resize(es)
		r.extensionLabel.Move(fyne.NewPos(w-padding-es.Width-4, padding+4))
	}
}

func (r *mediaCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(216, 192)
}

func (r *mediaCardRenderer) Refresh() {
	// Update content reference from card (may have changed asynchronously)
	r.content = r.card.content

	// Refresh all objects - no fyne.Do needed as renderer methods are called on main thread
	canvas.Refresh(r.background)
	if r.content != nil {
		canvas.Refresh(r.content)
	}
	canvas.Refresh(r.labelBackground)
	canvas.Refresh(r.label)
	if r.durationLabel != nil {
		canvas.Refresh(r.durationLabel)
	}
	if r.extensionLabel != nil {
		canvas.Refresh(r.extensionLabel)
	}
	// Note: Layout() should NOT be called from Refresh() per Fyne docs
}
func (r *mediaCardRenderer) Objects() []fyne.CanvasObject {
	// Build objects list, ensuring no nil values (content may be nil during async load)
	objs := []fyne.CanvasObject{r.background}
	if r.content != nil {
		objs = append(objs, r.content)
	}
	objs = append(objs, r.labelBackground, r.label)
	if r.durationLabel != nil {
		objs = append(objs, r.durationLabel)
	}
	if r.extensionLabel != nil {
		objs = append(objs, r.extensionLabel)
	}
	return objs
}

func (r *mediaCardRenderer) Destroy() {
	if r.card.animatedGif != nil {
		r.card.animatedGif.Stop()
	}
}

func GetMediaType(filename string) MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".tif", ".ico", ".svg":
		return MediaTypeImage
	case ".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv", ".ts", ".3gp", ".mpeg", ".mpg", ".wmv", ".m4v", ".vob", ".divx":
		return MediaTypeVideo
	default:
		return MediaTypeFile
	}
}
