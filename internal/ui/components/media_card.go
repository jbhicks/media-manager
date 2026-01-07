package components

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
	mediaType          MediaType
	filePath           string
	fileName           string
	thumbnailPath      string              // Static JPG thumbnail path
	animatedFrames     *AnimatedFrames     // Frame-based animation widget
	staticImage        *canvas.Image       // Static thumbnail image
	forceRegenerate    bool
	icon               *widget.Icon
	label              *widget.Label
	labelBackground    fyne.CanvasObject
	background         *canvas.Rectangle
	content            fyne.CanvasObject
	isHovered          bool
	hasAnimation       bool
	animatedRequested  bool                // Whether animated preview has been requested
	onDelete           func()
	thumbDir           string
	duration           int
	extension          string
	durationLabel      *widget.Label
	extensionLabel     *widget.Label
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
			mc.updateStaticThumbnail(mc.thumbnailPath)
			return
		}
	}

	// Request static JPG thumbnail from centralized manager with callback
	preview.GetPreviewWithCallback(mc.filePath, mediaTypeStr, mc.thumbDir, mc.forceRegenerate, func(path string, err error) {
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate preview for %s: %v\n", mc.filePath, err)
			return
		}
		mc.thumbnailPath = path
		mc.updateStaticThumbnail(path)
	})
}

// updateStaticThumbnail updates the card with a static JPG thumbnail
func (mc *MediaCard) updateStaticThumbnail(path string) {
	fyne.Do(func() {
		img := canvas.NewImageFromFile(path)
		img.FillMode = canvas.ImageFillContain
		mc.staticImage = img
		mc.content = img
		mc.Refresh()
	})
}

// loadAnimatedPreview loads animation frames on hover (for videos)
func (mc *MediaCard) loadAnimatedPreview() {
	if mc.animatedRequested {
		fmt.Printf("[DEBUG] loadAnimatedPreview: Already requested for %s\n", mc.fileName)
		return // Already requested
	}
	mc.animatedRequested = true
	fmt.Printf("[DEBUG] loadAnimatedPreview: Requesting frames for %s\n", mc.fileName)

	// Request frame extraction for animated preview
	preview.GetAnimatedPreviewWithCallback(mc.filePath, mc.thumbDir, mc.forceRegenerate, func(framePaths []string, err error) {
		fmt.Printf("[DEBUG] loadAnimatedPreview callback: got %d frames, err=%v\n", len(framePaths), err)
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate animation frames for %s: %v\n", mc.filePath, err)
			return
		}
		if len(framePaths) == 0 {
			fmt.Printf("[WARN] No animation frames generated for %s\n", mc.filePath)
			return
		}
		mc.updateAnimatedContent(framePaths)
	})
}

// updateAnimatedContent updates the card to show animated preview using frames
func (mc *MediaCard) updateAnimatedContent(framePaths []string) {
	fmt.Printf("[DEBUG] updateAnimatedContent: Creating AnimatedFrames with %d frames, isHovered=%v\n", len(framePaths), mc.isHovered)
	fyne.Do(func() {
		// Create AnimatedFrames widget
		animFrames := NewAnimatedFrames(framePaths)
		if animFrames == nil {
			fmt.Printf("[ERROR] NewAnimatedFrames returned nil\n")
			return
		}
		mc.animatedFrames = animFrames
		mc.hasAnimation = true

		// If still hovering, start animation - use the display image directly
		if mc.isHovered {
			fmt.Printf("[DEBUG] updateAnimatedContent: Starting animation\n")
			mc.content = animFrames.CurrentImage()
			animFrames.Start()
			mc.Refresh()
		} else {
			fmt.Printf("[DEBUG] updateAnimatedContent: Not hovering, animation ready but not started\n")
		}
	})
}

var _ desktop.Hoverable = (*MediaCard)(nil)

func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] MediaCard MouseIn - hover started")
	mc.isHovered = true
	mc.background.FillColor = theme.Color(theme.ColorNameHover)
	mc.background.Refresh()

	fmt.Printf("[DEBUG] MouseIn: mediaType=%v, hasAnimation=%v, mc.content=%v\n", mc.mediaType, mc.hasAnimation, mc.content)
	
	// For videos, start loading animated preview on hover (if not already loaded)
	if mc.mediaType == MediaTypeVideo {
		if mc.hasAnimation && mc.animatedFrames != nil {
			// Animation frames ready - use the animated frames' display image directly
			fmt.Printf("[DEBUG] MouseIn: Switching content to animatedFrames.displayImage and starting animation\n")
			mc.content = mc.animatedFrames.CurrentImage()
			mc.animatedFrames.Start()
			// Force full widget refresh to update renderer's content reference
			mc.BaseWidget.Refresh()
		} else if !mc.animatedRequested {
			// Request animated frame extraction
			go mc.loadAnimatedPreview()
		}
	}
}

func (mc *MediaCard) MouseOut() {
	fmt.Println("[DEBUG] MediaCard MouseOut - hover ended")
	mc.isHovered = false
	mc.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	mc.background.Refresh()

	// Stop animation and return to static thumbnail
	if mc.animatedFrames != nil {
		mc.animatedFrames.Stop()
	}
	
	// Restore static thumbnail if we have it
	if mc.staticImage != nil {
		mc.content = mc.staticImage
		mc.Refresh()
	}

	fmt.Printf("[DEBUG] MouseOut: mediaType=%v, hasAnimation=%v, animatedFrames=%v\n", mc.mediaType, mc.hasAnimation, mc.animatedFrames != nil)
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
	r := &mediaCardRenderer{
		card:            mc,
		background:      mc.background,
		content:         mc.content,
		labelBackground: mc.labelBackground,
		label:           mc.label,
		durationLabel:   mc.durationLabel,
		extensionLabel:  mc.extensionLabel,
	}
	r.updateObjectsCache()
	return r
}

type mediaCardRenderer struct {
	card            *MediaCard
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	labelBackground fyne.CanvasObject
	label           *widget.Label
	durationLabel   *widget.Label
	extensionLabel  *widget.Label
	objects         []fyne.CanvasObject // Cached objects slice
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
	oldContent := r.content
	r.content = r.card.content

	// Rebuild objects cache if content changed
	if oldContent != r.content {
		fmt.Printf("[DEBUG] mediaCardRenderer.Refresh: Content changed from %T to %T\n", oldContent, r.content)
		r.updateObjectsCache()
		// Force layout update when content changes
		if r.content != nil {
			size := r.card.Size()
			padding := float32(8)
			contentH := size.Height - 48
			contentW := size.Width - 2*padding
			r.content.Resize(fyne.NewSize(contentW, contentH))
			r.content.Move(fyne.NewPos(padding, padding))
		}
	}

	// Refresh all canvas objects - some may be called from background threads via callbacks
	r.background.Refresh()
	if r.content != nil {
		r.content.Refresh()
	}
	r.labelBackground.Refresh()
	r.label.Refresh()
	if r.durationLabel != nil {
		r.durationLabel.Refresh()
	}
	if r.extensionLabel != nil {
		r.extensionLabel.Refresh()
	}
}

// updateObjectsCache rebuilds the cached objects slice when content changes
func (r *mediaCardRenderer) updateObjectsCache() {
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
	r.objects = objs
}

func (r *mediaCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *mediaCardRenderer) Destroy() {
	if r.card.animatedFrames != nil {
		r.card.animatedFrames.Stop()
	}
}

func GetMediaType(filename string) MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".tif", ".ico", ".svg":
		return MediaTypeImage
	case ".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv", ".ts", ".3gp", ".mpeg", ".mpg", ".wmv", ".m4v", ".vob", ".divx", ".asf", ".dv", ".mp2", ".rm":
		return MediaTypeVideo
	default:
		return MediaTypeFile
	}
}
