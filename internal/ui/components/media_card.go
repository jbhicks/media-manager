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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/pkg/models"
)

// Card dimension constants - 16:9 aspect ratio for widescreen media
// Base dimensions at 1.0x zoom level
const (
	BaseCardWidth    = float32(288) // 16:9 aspect ratio width at 1x zoom
	BaseCardHeight   = float32(162) // 16:9 aspect ratio height at 1x zoom
	CardCornerRadius = float32(8)   // Rounded corners
	GradientHeight   = float32(64)  // Increased height for label to take more space
	BadgePadding     = float32(6)   // Padding for badges
)

// CardWidth returns the current card width scaled by theme
func CardWidth() float32 {
	// Use theme padding as a proxy for zoom level
	// Default padding is 4, so scale = currentPadding / 4
	scale := theme.Padding() / 4
	return BaseCardWidth * scale
}

// CardHeight returns the current card height scaled by theme
func CardHeight() float32 {
	scale := theme.Padding() / 4
	return BaseCardHeight * scale
}

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
// Uses a modern full-bleed design with image filling the card and overlay labels
type MediaCard struct {
	widget.BaseWidget
	mediaType          MediaType
	filePath           string
	fileName           string
	thumbnailPath      string          // Static JPG thumbnail path
	animatedFrames     *AnimatedFrames // Frame-based animation widget
	staticImage        *canvas.Image   // Static thumbnail image
	forceRegenerate    bool
	icon               *widget.Icon
	label              *canvas.Text           // Filename label (canvas.Text for styling)
	labelBackground    *canvas.LinearGradient // Bottom gradient overlay
	background         *canvas.Rectangle      // Card background with rounded corners
	hoverOverlay       *canvas.Rectangle      // Semi-transparent hover effect
	highlightOverlay   *canvas.Rectangle      // Temporary highlight for launch focus
	content            fyne.CanvasObject
	isHovered          bool
	hasAnimation       bool
	animatedRequested  bool // Whether animated preview has been requested
	onDelete           func()
	onPreviewGenerated func(filePath, previewPath string) // Callback to update database
	onDragStart        func()
	onDragged          func(abs fyne.Position)
	onDragEnd          func(dropPos fyne.Position)
	onOpen             func()
	isDragging         bool
	didDrag            bool
	lastDragAbs        fyne.Position
	dragStartAbs       fyne.Position
	dragAccumX         float32
	dragAccumY         float32
	thumbDir           string
	duration           int
	extension          string
	durationBadge      *canvas.Rectangle // Background for duration badge
	durationLabel      *canvas.Text      // Duration text
	extensionBadge     *canvas.Rectangle // Background for extension badge
	extensionLabel     *canvas.Text      // Extension text
}

func NewMediaCard(file models.MediaFile, thumbDir string, onPreviewGenerated func(filePath, previewPath string)) *MediaCard {
	return NewMediaCardWithForce(file, thumbDir, false, onPreviewGenerated)
}

func NewMediaCardWithForce(file models.MediaFile, thumbDir string, forceRegenerate bool, onPreviewGenerated func(filePath, previewPath string)) *MediaCard {
	mediaType := GetMediaType(file.Filename)

	// Use friendly title if available, otherwise use filename
	displayName := file.FriendlyTitle
	if displayName == "" {
		displayName = file.Filename
	}

	// Allow more characters for display since we want text to take most space
	// Use a longer limit based on card width - approximately 50-60 characters for readable text
	if len(displayName) > 60 {
		displayName = displayName[:57] + "..."
	}

	card := &MediaCard{
		mediaType:          mediaType,
		filePath:           file.Path,
		fileName:           file.Filename,
		thumbnailPath:      file.PreviewPath,
		thumbDir:           thumbDir,
		duration:           file.Duration,
		extension:          strings.ToUpper(strings.TrimPrefix(filepath.Ext(file.Filename), ".")),
		isHovered:          false,
		hasAnimation:       false,
		forceRegenerate:    forceRegenerate,
		onPreviewGenerated: onPreviewGenerated,
	}

	card.setupContent()

	// Setup card background with rounded corners
	card.background = canvas.NewRectangle(color.NRGBA{30, 30, 30, 255})
	card.background.CornerRadius = CardCornerRadius
	card.background.StrokeColor = color.NRGBA{60, 60, 60, 255}
	card.background.StrokeWidth = 1

	// Setup hover overlay (initially transparent)
	card.hoverOverlay = canvas.NewRectangle(color.Transparent)
	card.hoverOverlay.CornerRadius = CardCornerRadius

	// Setup highlight overlay (initially transparent)
	card.highlightOverlay = canvas.NewRectangle(color.Transparent)
	card.highlightOverlay.CornerRadius = CardCornerRadius
	card.highlightOverlay.StrokeWidth = 0

	// Setup bottom gradient for label legibility (transparent to semi-opaque black)
	card.labelBackground = canvas.NewLinearGradient(
		color.NRGBA{0, 0, 0, 0},   // Top: fully transparent
		color.NRGBA{0, 0, 0, 200}, // Bottom: 78% opaque black
		90,                        // Vertical gradient
	)

	// Setup filename label - white text with shadow effect
	// Smaller text size to fit more characters
	card.label = canvas.NewText(displayName, color.White)
	card.label.TextSize = 10
	card.label.TextStyle = fyne.TextStyle{}

	// Setup extension badge (top-right pill)
	card.extensionBadge = canvas.NewRectangle(color.NRGBA{0, 0, 0, 150})
	card.extensionBadge.CornerRadius = 4
	card.extensionLabel = canvas.NewText(card.extension, color.White)
	card.extensionLabel.TextSize = 10
	card.extensionLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Setup duration badge (bottom-right pill) - only for videos with duration
	if card.duration > 0 {
		mins := card.duration / 60
		secs := card.duration % 60
		card.durationBadge = canvas.NewRectangle(color.NRGBA{0, 0, 0, 180})
		card.durationBadge.CornerRadius = 4
		card.durationLabel = canvas.NewText(fmt.Sprintf("%d:%02d", mins, secs), color.White)
		card.durationLabel.TextSize = 11
		card.durationLabel.TextStyle = fyne.TextStyle{Bold: true}
	}

	card.ExtendBaseWidget(card)
	return card
}

func (mc *MediaCard) setupContent() {
	switch mc.mediaType {
	case MediaTypeImage:
		go mc.loadPreview("image")
	case MediaTypeVideo:
		// Always request the static thumbnail in background so cards update as soon as
		// preview generation completes. Animated previews remain lazy on hover.
		mc.content = widget.NewIcon(theme.MediaPlayIcon())
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
		// Update database with the preview path
		if mc.onPreviewGenerated != nil {
			mc.onPreviewGenerated(mc.filePath, path)
		}
	})
}

// updateStaticThumbnail updates the card with a static JPG thumbnail
func (mc *MediaCard) updateStaticThumbnail(path string) {
	// Load the image resource in the background to avoid blocking the UI thread.
	go func() {
		res, err := fyne.LoadResourceFromPath(path)
		if err != nil {
			fmt.Printf("[ERROR] Failed to load thumbnail resource %s: %v\n", path, err)
			return
		}
		fyne.Do(func() {
			img := canvas.NewImageFromResource(res)
			img.FillMode = canvas.ImageFillStretch
			img.ScaleMode = canvas.ImageScaleFastest
			mc.staticImage = img
			mc.content = img
			mc.Refresh()
		})
	}()
}

// loadAnimatedPreview loads animation frames on hover (for videos)
func (mc *MediaCard) loadAnimatedPreview() {
	if mc.animatedRequested {
		return // Already requested
	}
	mc.animatedRequested = true

	// Request frame extraction for animated preview
	preview.GetAnimatedPreviewWithCallback(mc.filePath, mc.thumbDir, mc.forceRegenerate, func(framePaths []string, err error) {
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate animation frames for %s: %v\n", mc.filePath, err)
			return
		}
		if len(framePaths) == 0 {
			return
		}
		mc.updateAnimatedContent(framePaths)
	})
}

// updateAnimatedContent updates the card to show animated preview using frames
func (mc *MediaCard) updateAnimatedContent(framePaths []string) {
	// Load frames asynchronously to avoid blocking UI thread during disk I/O
	NewAnimatedFramesAsync(framePaths, func(animFrames *AnimatedFrames) {
		if animFrames == nil {
			return
		}
		mc.animatedFrames = animFrames
		mc.hasAnimation = true

		// If still hovering, start animation - use the display image directly
		if mc.isHovered {
			mc.content = animFrames.CurrentImage()
			animFrames.Start()
			mc.Refresh()
		}
	})
}

const dragMoveThreshold = float32(4)

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

var _ desktop.Hoverable = (*MediaCard)(nil)
var _ fyne.Draggable = (*MediaCard)(nil)

// Dragged records the pointer position, treats movement past 4px as a real
// drag, and notifies listeners. Tiny jitter stays a click.
func (mc *MediaCard) Dragged(ev *fyne.DragEvent) {
	if ev != nil {
		if !mc.isDragging {
			mc.dragStartAbs = ev.AbsolutePosition
		}
		mc.lastDragAbs = ev.AbsolutePosition
		mc.dragAccumX += ev.Dragged.DX
		mc.dragAccumY += ev.Dragged.DY
	}

	moved := abs32(mc.dragAccumX) >= dragMoveThreshold || abs32(mc.dragAccumY) >= dragMoveThreshold
	if !moved && ev != nil {
		dx := ev.AbsolutePosition.X - mc.dragStartAbs.X
		dy := ev.AbsolutePosition.Y - mc.dragStartAbs.Y
		moved = abs32(dx) >= dragMoveThreshold || abs32(dy) >= dragMoveThreshold
	}

	if moved {
		if !mc.didDrag {
			mc.didDrag = true
			if mc.onDragStart != nil {
				mc.onDragStart()
			}
		}
		if mc.onDragged != nil && ev != nil {
			mc.onDragged(ev.AbsolutePosition)
		}
	}

	mc.isDragging = true
}

// DragEnd is called when the drag operation ends.
func (mc *MediaCard) DragEnd() {
	wasDrag := mc.didDrag
	mc.isDragging = false
	mc.dragAccumX = 0
	mc.dragAccumY = 0
	if wasDrag && mc.onDragEnd != nil {
		mc.onDragEnd(mc.lastDragAbs)
	}
}

// SetOnDragStart sets the callback invoked when the card starts being dragged.
func (mc *MediaCard) SetOnDragStart(callback func()) {
	mc.onDragStart = callback
}

// SetOnDragged sets the callback invoked on each move after the drag threshold.
func (mc *MediaCard) SetOnDragged(callback func(abs fyne.Position)) {
	mc.onDragged = callback
}

// SetOnDragEnd sets the callback invoked when the card stops being dragged.
// dropPos is the last absolute pointer position from Dragged.
func (mc *MediaCard) SetOnDragEnd(callback func(dropPos fyne.Position)) {
	mc.onDragEnd = callback
}

func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	if !mc.isDragging && mc.didDrag {
		mc.didDrag = false
	}
	mc.isHovered = true

	// Show hover overlay effect
	mc.hoverOverlay.FillColor = color.NRGBA{255, 255, 255, 30} // Slight white overlay
	mc.hoverOverlay.Refresh()

	// For videos, start loading animated preview on hover (if not already loaded)
	if mc.mediaType == MediaTypeVideo {
		if mc.hasAnimation && mc.animatedFrames != nil {
			// Animation frames ready - use the animated frames' display image directly
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
	mc.isHovered = false

	// Hide hover overlay
	mc.hoverOverlay.FillColor = color.Transparent
	mc.hoverOverlay.Refresh()

	// Stop animation and return to static thumbnail
	if mc.animatedFrames != nil {
		mc.animatedFrames.Stop()
	}

	// Restore static thumbnail if we have it
	if mc.staticImage != nil {
		mc.content = mc.staticImage
		mc.Refresh()
	}
}

func (mc *MediaCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

func (mc *MediaCard) Tapped(*fyne.PointEvent) {
	if mc.didDrag {
		mc.didDrag = false
		return
	}
	if mc.onOpen != nil {
		mc.onOpen()
		return
	}
	fmt.Printf("[DEBUG] MediaCard Tapped: %s\n", mc.filePath)
	err := mc.openFile()
	if err != nil {
		fmt.Printf("[DEBUG] Error opening file: %v\n", err)
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

// FilePath returns the path of the media file represented by this card.
func (mc *MediaCard) FilePath() string {
	return mc.filePath
}

func (mc *MediaCard) SetHighlighted(highlighted bool) {
	fyne.Do(func() {
		if highlighted {
			mc.highlightOverlay.FillColor = color.NRGBA{255, 210, 0, 40}
			mc.highlightOverlay.StrokeColor = color.NRGBA{255, 210, 0, 220}
			mc.highlightOverlay.StrokeWidth = 2
			mc.highlightOverlay.Refresh()

			go func() {
				time.Sleep(3 * time.Second)
				fyne.Do(func() {
					mc.highlightOverlay.FillColor = color.Transparent
					mc.highlightOverlay.StrokeColor = color.Transparent
					mc.highlightOverlay.StrokeWidth = 0
					mc.highlightOverlay.Refresh()
				})
			}()
			return
		}

		mc.highlightOverlay.FillColor = color.Transparent
		mc.highlightOverlay.StrokeColor = color.Transparent
		mc.highlightOverlay.StrokeWidth = 0
		mc.highlightOverlay.Refresh()
	})
}

func (mc *MediaCard) openFile() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		fmt.Printf("[DEBUG] Opening file on Windows: %s\n", mc.filePath)
		// Check if file exists first
		if _, err := os.Stat(mc.filePath); os.IsNotExist(err) {
			fmt.Printf("[DEBUG] File does not exist: %s\n", mc.filePath)
			return fmt.Errorf("file does not exist: %s", mc.filePath)
		}
		fmt.Printf("[DEBUG] File exists, attempting to open\n")
		// Use explorer.exe to open the file with default program
		cmd = exec.Command("explorer.exe", mc.filePath)
	case "darwin": // macOS
		cmd = exec.Command("open", mc.filePath)
	default: // Linux and others
		cmd = exec.Command("xdg-open", mc.filePath)
	}

	fmt.Printf("[DEBUG] Executing command: %v\n", cmd.Args)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("[DEBUG] cmd.Start() failed: %v\n", err)
		return fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("[DEBUG] Command finished with error: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] Command finished successfully\n")
		}
	}()
	return nil
}

func (mc *MediaCard) MinSize() fyne.Size {
	return fyne.NewSize(CardWidth(), CardHeight())
}

func (mc *MediaCard) CreateRenderer() fyne.WidgetRenderer {
	r := &mediaCardRenderer{
		card:             mc,
		background:       mc.background,
		content:          mc.content,
		hoverOverlay:     mc.hoverOverlay,
		highlightOverlay: mc.highlightOverlay,
		labelBackground:  mc.labelBackground,
		label:            mc.label,
		durationBadge:    mc.durationBadge,
		durationLabel:    mc.durationLabel,
		extensionBadge:   mc.extensionBadge,
		extensionLabel:   mc.extensionLabel,
	}
	r.updateObjectsCache()
	return r
}

type mediaCardRenderer struct {
	card             *MediaCard
	background       *canvas.Rectangle
	content          fyne.CanvasObject
	hoverOverlay     *canvas.Rectangle
	highlightOverlay *canvas.Rectangle
	labelBackground  *canvas.LinearGradient
	label            *canvas.Text
	durationBadge    *canvas.Rectangle
	durationLabel    *canvas.Text
	extensionBadge   *canvas.Rectangle
	extensionLabel   *canvas.Text
	objects          []fyne.CanvasObject // Cached objects slice
}

func (r *mediaCardRenderer) Layout(size fyne.Size) {
	w, h := size.Width, size.Height

	// Background fills entire card
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	// Content (image) fills entire card - full bleed design
	if r.content != nil {
		r.content.Resize(size)
		r.content.Move(fyne.NewPos(0, 0))
	}

	// Hover overlay fills entire card
	r.hoverOverlay.Resize(size)
	r.hoverOverlay.Move(fyne.NewPos(0, 0))

	// Highlight overlay fills entire card
	r.highlightOverlay.Resize(size)
	r.highlightOverlay.Move(fyne.NewPos(0, 0))

	// Bottom gradient for label - spans full width at bottom
	r.labelBackground.Resize(fyne.NewSize(w, GradientHeight))
	r.labelBackground.Move(fyne.NewPos(0, h-GradientHeight))

	// Filename label - positioned at bottom with padding, allowing more vertical space
	labelPadding := float32(6)
	r.label.Move(fyne.NewPos(labelPadding, h-GradientHeight+labelPadding))

	// Extension badge (top-right)
	if r.extensionLabel != nil && r.extensionBadge != nil {
		textSize := fyne.MeasureText(r.extensionLabel.Text, r.extensionLabel.TextSize, r.extensionLabel.TextStyle)
		badgeW := textSize.Width + BadgePadding*2
		badgeH := textSize.Height + BadgePadding
		r.extensionBadge.Resize(fyne.NewSize(badgeW, badgeH))
		r.extensionBadge.Move(fyne.NewPos(w-badgeW-BadgePadding, BadgePadding))
		r.extensionLabel.Move(fyne.NewPos(w-badgeW-BadgePadding+BadgePadding, BadgePadding+BadgePadding/2))
	}

	// Duration badge (bottom-right, above the filename area)
	if r.durationLabel != nil && r.durationBadge != nil {
		textSize := fyne.MeasureText(r.durationLabel.Text, r.durationLabel.TextSize, r.durationLabel.TextStyle)
		badgeW := textSize.Width + BadgePadding*2
		badgeH := textSize.Height + BadgePadding
		// Position above the gradient area, accounting for increased gradient height
		badgeY := h - GradientHeight - badgeH - BadgePadding + 10 // Adjust for taller gradient
		r.durationBadge.Resize(fyne.NewSize(badgeW, badgeH))
		r.durationBadge.Move(fyne.NewPos(w-badgeW-BadgePadding, badgeY))
		r.durationLabel.Move(fyne.NewPos(w-badgeW-BadgePadding+BadgePadding, badgeY+BadgePadding/2))
	}
}

func (r *mediaCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(CardWidth(), CardHeight())
}

func (r *mediaCardRenderer) Refresh() {
	// Update content reference from card (may have changed asynchronously)
	oldContent := r.content
	r.content = r.card.content

	// Rebuild objects cache if content changed
	if oldContent != r.content {
		r.updateObjectsCache()
		// Force layout update when content changes
		if r.content != nil {
			size := r.card.Size()
			r.content.Resize(size)
			r.content.Move(fyne.NewPos(0, 0))
		}
	}

	// Refresh all canvas objects
	r.background.Refresh()
	if r.content != nil {
		r.content.Refresh()
	}
	r.hoverOverlay.Refresh()
	r.highlightOverlay.Refresh()
	r.labelBackground.Refresh()
	r.label.Refresh()
	if r.durationBadge != nil {
		r.durationBadge.Refresh()
	}
	if r.durationLabel != nil {
		r.durationLabel.Refresh()
	}
	if r.extensionBadge != nil {
		r.extensionBadge.Refresh()
	}
	if r.extensionLabel != nil {
		r.extensionLabel.Refresh()
	}
}

// updateObjectsCache rebuilds the cached objects slice when content changes
func (r *mediaCardRenderer) updateObjectsCache() {
	// Layer order (back to front):
	// 1. background
	// 2. content (image)
	// 3. hover overlay
	// 4. label gradient background
	// 5. extension badge + label
	// 6. duration badge + label
	// 7. filename label
	objs := []fyne.CanvasObject{r.background}
	if r.content != nil {
		objs = append(objs, r.content)
	}
	objs = append(objs, r.hoverOverlay, r.highlightOverlay, r.labelBackground)
	if r.extensionBadge != nil {
		objs = append(objs, r.extensionBadge)
	}
	if r.extensionLabel != nil {
		objs = append(objs, r.extensionLabel)
	}
	if r.durationBadge != nil {
		objs = append(objs, r.durationBadge)
	}
	if r.durationLabel != nil {
		objs = append(objs, r.durationLabel)
	}
	objs = append(objs, r.label)
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
