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
	xwidget "fyne.io/x/fyne/widget"
	"image/color"
	"os/exec"
	"path/filepath"
	"strings"
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
}

// NewMediaCard creates a new uniform media card
func NewMediaCard(filePath, fileName string, mediaType MediaType, thumbPath string) *MediaCard {
	fmt.Printf("[DEBUG] NewMediaCard: Creating card for %s (Type: %v)\n", fileName, mediaType)
	// Truncate filename for display - make it shorter
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
		hasAnimation:  false, // Will be set to true if animation is available/generated
	}

	// Setup content based on media type
	card.setupContent()

	// Create label
	card.label = widget.NewLabelWithStyle(displayName, fyne.TextAlignCenter, fyne.TextStyle{})
	// Always visible label for overlay effect

	// Create background rectangle with a visible color
	card.background = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))

	// Set a darker stroke/border for visibility
	card.background.StrokeColor = color.NRGBA{100, 100, 100, 255}
	card.background.StrokeWidth = 1

	// Create semi-transparent background for label text
	card.labelBackground = canvas.NewLinearGradient(
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 180},
		90,
	)
	card.ExtendBaseWidget(card)

	fmt.Printf("[DEBUG] NewMediaCard: Card created for %s. hasAnimation: %v, animatedGif: %v\n", fileName, card.hasAnimation, card.animatedGif != nil)
	return card
}

func (mc *MediaCard) setupContent() {
	var homeDir string
	var err error
	var uri fyne.URI
	var gifWidget *xwidget.AnimatedGif
	var fileInfo os.FileInfo

	// Video card setup
	fmt.Printf("[DEBUG] setupContent: Setting up content for %s (Type: %v)", mc.fileName, mc.mediaType)

	switch mc.mediaType {
	case MediaTypeImage:
		mc.content = widget.NewIcon(theme.FileImageIcon())
	case MediaTypeVideo:
		mc.setDefaultVideoIcon()
	case MediaTypeFile:
		mc.content = widget.NewIcon(theme.FileIcon())
	}

	// Generate thumbnail in background
	go mc.generateThumbnail()
}

func (mc *MediaCard) generateThumbnail() {
	// Generate thumbnail in background
}

func (mc *MediaCard) setDefaultVideoIcon() {
	mc.icon = widget.NewIcon(theme.FileVideoIcon())
	mc.content = mc.icon
}

// Ensure MediaCard implements desktop.Hoverable
var _ desktop.Hoverable = (*MediaCard)(nil)

// MouseIn handles hover start (GIF preview only)
func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] MediaCard MouseIn - hover started")
	mc.isHovered = true
	mc.background.FillColor = theme.Color(theme.ColorNameHover)
	mc.background.Refresh()

	fmt.Printf("[DEBUG] MouseIn: mediaType=%v, hasAnimation=%v, mc.content=%v\n", mc.mediaType, mc.hasAnimation, mc.content)
	// Start animation for videos
	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.content = mc.animatedGif
		mc.animatedGif.Start()
		mc.Refresh()
	} else if mc.mediaType == MediaTypeImage {
		// For images, make sure the static image is visible
		mc.Refresh()
	}
}

// MouseOut handles hover end (GIF preview only)
func (mc *MediaCard) MouseOut() {
	fmt.Println("[DEBUG] MediaCard MouseOut - hover ended")
	mc.isHovered = false
	mc.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	mc.background.Refresh()

	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.animatedGif.Stop()
	}

	fmt.Printf("[DEBUG] MouseOut: mediaType=%v, hasAnimation=%v, animatedGif=%v\n", mc.mediaType, mc.hasAnimation, mc.animatedGif != nil)
}

// MouseMoved handles mouse movement
func (mc *MediaCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

// Tapped handles tap (click) events
func (mc *MediaCard) Tapped(*fyne.PointEvent) {
	fmt.Printf("[DEBUG] MediaCard Tapped: %s\n", mc.filePath)
	err := mc.openFile()
	if err != nil {
		fmt.Printf("[DEBUG] Error opening file: %v\n", err)
	}
}

// TappedSecondary handles secondary tap (right-click) events
func (mc *MediaCard) TappedSecondary(e *fyne.PointEvent) {
	fmt.Printf("[DEBUG] MediaCard TappedSecondary: %s\n", mc.filePath)
	deleteMenuItem := fyne.NewMenuItem("Delete", func() {
		err := os.Remove(mc.filePath)
		if err != nil {
			fmt.Printf("[ERROR] Failed to delete file: %v\n", err)
			return
		}
		if mc.onDelete != nil {
			mc.onDelete()
		}
	})
	canvas := fyne.CurrentApp().Driver().CanvasForObject(mc)
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("", deleteMenuItem), canvas, e.AbsolutePosition)
}

// SetOnDelete sets the callback function for when the card is deleted.
func (mc *MediaCard) SetOnDelete(callback func()) {
	mc.onDelete = callback
}

// openFile opens the media file using the default system application
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

	// Detach the process from the current one
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("[DEBUG] Command finished with error: %v\n", err)
		}
	}()
	return nil
}

// MinSize returns the fixed size for all cards
func (mc *MediaCard) MinSize() fyne.Size {
	return fyne.NewSize(180, 180) // Fixed size for all cards
}

// CreateRenderer creates the renderer for the media card
func (mc *MediaCard) CreateRenderer() fyne.WidgetRenderer {
	return &mediaCardRenderer{
		card:            mc,
		background:      mc.background,
		content:         mc.content,
		labelBackground: mc.labelBackground,
		label:           mc.label,
	}
}

// mediaCardRenderer renders the media card with precise control over layout
type mediaCardRenderer struct {
	card            *MediaCard
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	labelBackground fyne.CanvasObject
	label           *widget.Label
}

func (r *mediaCardRenderer) Layout(size fyne.Size) {
	fmt.Printf("[DEBUG] mediaCardRenderer.Layout - received size: %v\n", size)
	// Background fills the entire card

	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	labelAreaHeight := r.label.MinSize().Height + float32(8)
	contentAvailableHeight := size.Height - labelAreaHeight
	contentAvailableWidth := size.Width

	// Make the content fill most of the card space
	contentSize := fyne.NewSize(contentAvailableWidth, contentAvailableHeight)

	// Position the content centered in the available space
	contentX := float32(0)
	contentY := float32(0)

	// Resize and position the content
	if r.content != nil {
		r.content.Resize(contentSize)
		r.content.Move(fyne.NewPos(contentX, contentY))
	}

	fmt.Printf("[DEBUG] mediaCardRenderer.Layout - contentSize: %v\n", contentSize)

	// Force a refresh of the content
	canvas.Refresh(r.content)

	// Label and overlay always visible, anchored to bottom
	labelMinHeight := float32(22)
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
	return fyne.NewSize(180, 180)
}

func (r *mediaCardRenderer) Refresh() {
	// Update content reference in case it changed (video hover)
	r.content = r.card.content

	// Refresh all objects
	canvas.Refresh(r.background)
	if r.content != nil {
		canvas.Refresh(r.content)
	}
	canvas.Refresh(r.labelBackground)
	canvas.Refresh(r.label)

	// Force a layout update to ensure content is positioned correctly
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

// Helper functions to determine media type
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
