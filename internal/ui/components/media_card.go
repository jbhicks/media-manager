package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"image/color"
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
	animatedGif     *xwidget.AnimatedGif
	icon            *widget.Icon
	label           *widget.Label
	labelBackground fyne.CanvasObject
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	isHovered       bool
	hasAnimation    bool
}

// NewMediaCard creates a new uniform media card
func NewMediaCard(filePath, fileName string, mediaType MediaType) *MediaCard {
	// Truncate filename for display - make it shorter
	displayName := fileName
	if len(displayName) > 22 {
		displayName = displayName[:19] + "..."
	}

	card := &MediaCard{
		mediaType:    mediaType,
		filePath:     filePath,
		fileName:     fileName,
		isHovered:    false,
		hasAnimation: false,
	}

	// Setup content based on media type
	card.setupContent()

	// Create label
	card.label = widget.NewLabelWithStyle(displayName, fyne.TextAlignCenter, fyne.TextStyle{})
	// Always visible label for overlay effect

	// Create background rectangle
	card.background = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))

	// Create semi-transparent background for label text
	card.labelBackground = canvas.NewLinearGradient(
		color.NRGBA{0, 0, 0, 0},
		color.NRGBA{0, 0, 0, 180},
		90,
	)
	card.ExtendBaseWidget(card)
	return card
}

func (mc *MediaCard) setupContent() {
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
	var img fyne.Resource
	var err error

	switch mc.mediaType {
	case MediaTypeImage:
		img, err = fyne.LoadResourceFromPath(mc.filePath)
		if err != nil {
			fmt.Printf("Error loading image %s: %v\n", mc.filePath, err)
			return
		}
	case MediaTypeVideo:
		// Generate video thumbnail (single frame PNG) using ffmpeg
		thumbPath := filepath.Join(os.TempDir(), "media-manager-thumbs", mc.fileName+".png")
		_ = os.MkdirAll(filepath.Dir(thumbPath), 0755)

		cmd := exec.Command("ffmpeg", "-y", "-i", mc.filePath, "-ss", "00:00:01.000", "-vframes", "1", "-vf", "scale=200:200:force_original_aspect_ratio=increase,crop=200:200", thumbPath)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error generating video thumbnail for %s: %v\n", mc.filePath, err)
			return
		}
		img, err = fyne.LoadResourceFromPath(thumbPath)
		if err != nil {
			fmt.Printf("Error loading video thumbnail %s: %v\n", thumbPath, err)
			return
		}

		// Generate animated GIF for hover preview
		gifPath := filepath.Join(os.TempDir(), "media-manager-thumbs", mc.fileName+".gif")
		_ = os.MkdirAll(filepath.Dir(gifPath), 0755)

		// Extract 2 seconds of video starting from 1 second mark, scale to 200px width, and convert to GIF
		cmd = exec.Command("ffmpeg", "-y", "-i", mc.filePath, "-ss", "00:00:01.000", "-t", "2", "-vf", "fps=10,scale=200:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse", gifPath)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error generating animated GIF for %s: %v\n", mc.filePath, err)
		} else {
			gifURI := storage.NewFileURI(gifPath)
			animGif, animErr := xwidget.NewAnimatedGif(gifURI)
			if animErr == nil {
				mc.animatedGif = animGif
				mc.hasAnimation = true
			} else {
				fmt.Printf("Error creating animated GIF widget for %s: %v\n", gifPath, animErr)
			}
		}
	default:
		return
	}

	if img != nil {
		fyne.Do(func() {
			mc.content = canvas.NewImageFromResource(img)
			mc.Refresh()
		})
	}
}

func (mc *MediaCard) setDefaultVideoIcon() {
	mc.icon = widget.NewIcon(theme.FileVideoIcon())
	mc.content = mc.icon
}

// Ensure MediaCard implements desktop.Hoverable
var _ desktop.Hoverable = (*MediaCard)(nil)

// MouseIn handles hover start
func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] MediaCard MouseIn - hover started")
	mc.isHovered = true
	mc.background.FillColor = theme.Color(theme.ColorNameHover)
	mc.background.Refresh()

	fmt.Printf("[DEBUG] MouseIn: mediaType=%v, hasAnimation=%v, animatedGif=%v\n", mc.mediaType, mc.hasAnimation, mc.animatedGif != nil)
	// Start animation for videos
	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.content = mc.animatedGif
		fmt.Println("[DEBUG] Starting animated GIF")
		mc.animatedGif.Start()
		mc.Refresh()
	}
}

// MouseOut handles hover end
func (mc *MediaCard) MouseOut() {
	fmt.Println("[DEBUG] MediaCard MouseOut - hover ended")
	mc.isHovered = false
	mc.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	mc.background.Refresh()

	fmt.Printf("[DEBUG] MouseOut: mediaType=%v, hasAnimation=%v, animatedGif=%v\n", mc.mediaType, mc.hasAnimation, mc.animatedGif != nil)
	// Stop animation for videos
	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.animatedGif.Stop()
		mc.content = mc.animatedGif // Revert to the animated GIF (which shows its first frame when stopped)
		mc.Refresh()
	}
}

// MouseMoved handles mouse movement
func (mc *MediaCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

// MinSize returns the fixed size for all cards
func (mc *MediaCard) MinSize() fyne.Size {
	return fyne.NewSize(96, 64) // Clamp to icon and label
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
	// Background fills the entire card
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	padding := float32(2)
	labelAreaHeight := r.label.MinSize().Height + float32(8)
	contentAvailableHeight := size.Height - padding*2 - labelAreaHeight
	contentAvailableWidth := size.Width - padding*2

	// Center content using its MinSize, do not stretch
	contentMin := r.content.MinSize()
	contentX := padding + (contentAvailableWidth-contentMin.Width)/2
	contentY := padding + (contentAvailableHeight-contentMin.Height)/2
	r.content.Resize(contentMin)
	r.content.Move(fyne.NewPos(contentX, contentY))

	// Label and overlay always visible, anchored to bottom
	labelMinHeight := float32(22)
	labelWidth := size.Width - float32(8)
	labelX := float32(4)
	labelTextSize := r.label.MinSize()
	labelHeight := labelTextSize.Height
	if labelHeight < labelMinHeight {
		labelHeight = labelMinHeight
	}
	labelY := size.Height - labelHeight - float32(8)

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

func (mc *MediaCard) runCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}
	return exec.Command(args[0], args[1:]...).Run()
}
