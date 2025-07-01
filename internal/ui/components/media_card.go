package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/user/media-manager/internal/preview"
	"os"
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
	staticImage     *canvas.Image
	animatedGif     *xwidget.AnimatedGif
	icon            *widget.Icon
	label           *widget.Label
	labelBackground *canvas.Rectangle
	background      *canvas.Rectangle
	content         fyne.CanvasObject
	isHovered       bool
	hasAnimation    bool
}

// NewMediaCard creates a new uniform media card
func NewMediaCard(filePath, fileName string, mediaType MediaType) *MediaCard {
	// Truncate filename for display - make it shorter
	displayName := fileName
	if len(displayName) > 8 { // Much shorter to fit better
		displayName = displayName[:5] + "..."
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
	card.label.Hide() // Hidden by default, shown on hover

	// Create background rectangle
	card.background = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))

	// Create semi-transparent background for label text
	labelBgColor := theme.Color(theme.ColorNameShadow)
	card.labelBackground = canvas.NewRectangle(labelBgColor)
	card.labelBackground.Hide() // Hidden by default, shown on hover	card.ExtendBaseWidget(card)
	return card
}

func (mc *MediaCard) setupContent() {
	switch mc.mediaType {
	case MediaTypeImage:
		// Generate/use thumbnail for images too
		homeDir, _ := os.UserHomeDir()
		thumbDir := filepath.Join(homeDir, ".media-manager", "thumbnails")
		staticThumbPath := filepath.Join(thumbDir, mc.fileName+".jpg")

		if _, err := os.Stat(staticThumbPath); err == nil {
			// Use existing thumbnail
			mc.staticImage = canvas.NewImageFromFile(staticThumbPath)
		} else {
			// Generate thumbnail for image in background
			_ = os.MkdirAll(thumbDir, 0755)
			go func() {
				err := preview.GenerateThumbnail(mc.filePath, staticThumbPath)
				if err == nil {
					// Refresh the card after thumbnail generation
					mc.staticImage = canvas.NewImageFromFile(staticThumbPath)
					mc.staticImage.FillMode = canvas.ImageFillContain
					mc.content = mc.staticImage
					mc.Refresh()
				}
			}()
			// Use original image as placeholder while generating thumbnail
			mc.staticImage = canvas.NewImageFromFile(mc.filePath)
		}
		mc.staticImage.FillMode = canvas.ImageFillContain
		mc.content = mc.staticImage

	case MediaTypeVideo:
		// Setup video thumbnail
		homeDir, _ := os.UserHomeDir()
		thumbDir := filepath.Join(homeDir, ".media-manager", "thumbnails")
		staticThumbPath := filepath.Join(thumbDir, mc.fileName+".jpg")
		animatedGifPath := filepath.Join(thumbDir, mc.fileName+".gif")

		if _, err := os.Stat(staticThumbPath); err == nil {
			mc.staticImage = canvas.NewImageFromFile(staticThumbPath)
			mc.staticImage.FillMode = canvas.ImageFillContain
			mc.content = mc.staticImage

			// Check for animated GIF
			if _, err := os.Stat(animatedGifPath); err == nil {
				gifURI := storage.NewFileURI(animatedGifPath)
				if gif, err := xwidget.NewAnimatedGif(gifURI); err == nil {
					mc.animatedGif = gif
					mc.hasAnimation = true
				}
			}
		} else {
			// Generate video thumbnail if it doesn't exist
			mc.icon = widget.NewIcon(theme.FileVideoIcon())
			mc.content = mc.icon

			go func() {
				mc.ensureVideoThumbnail(mc.filePath, staticThumbPath)
				// After generation, update the card
				if _, err := os.Stat(staticThumbPath); err == nil {
					mc.staticImage = canvas.NewImageFromFile(staticThumbPath)
					mc.staticImage.FillMode = canvas.ImageFillContain
					mc.content = mc.staticImage
					mc.Refresh()
				}
			}()
		}

	default: // MediaTypeFile
		mc.icon = widget.NewIcon(theme.FileIcon())
		mc.content = mc.icon
	}
}

// Ensure MediaCard implements desktop.Hoverable
var _ desktop.Hoverable = (*MediaCard)(nil)

// MouseIn handles hover start
func (mc *MediaCard) MouseIn(*desktop.MouseEvent) {
	mc.isHovered = true
	mc.background.FillColor = theme.Color(theme.ColorNameHover)
	mc.background.Refresh()

	// Show label on hover
	mc.label.Show()
	mc.labelBackground.Show()

	// Start animation for videos
	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.content = mc.animatedGif
		mc.animatedGif.Start()
		mc.Refresh()
	}
}

// MouseOut handles hover end
func (mc *MediaCard) MouseOut() {
	mc.isHovered = false
	mc.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	mc.background.Refresh()

	// Hide label when not hovering
	mc.label.Hide()
	mc.labelBackground.Hide()

	// Stop animation for videos
	if mc.mediaType == MediaTypeVideo && mc.hasAnimation && mc.animatedGif != nil {
		mc.animatedGif.Stop()
		mc.content = mc.staticImage
		mc.Refresh()
	}
}

// MouseMoved handles mouse movement
func (mc *MediaCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

// MinSize returns the fixed size for all cards
func (mc *MediaCard) MinSize() fyne.Size {
	return fyne.NewSize(180, 180) // Twice as big - more room for content
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
	labelBackground *canvas.Rectangle
	label           *widget.Label
}

func (r *mediaCardRenderer) Layout(size fyne.Size) {
	// Background fills the entire card
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	// Content fills almost the entire card - this is crucial for image display
	padding := float32(2)
	contentSize := fyne.NewSize(size.Width-padding*2, size.Height-padding*2)
	r.content.Resize(contentSize)
	r.content.Move(fyne.NewPos(padding, padding))

	// Ensure the content is refreshed to display properly
	canvas.Refresh(r.content)

	// Label positioned well within the bottom area - only visible on hover
	labelHeight := float32(12)
	labelY := size.Height - labelHeight - float32(4) // Well within bounds
	labelX := padding + float32(2)
	labelWidth := size.Width - padding*2 - float32(4)

	// Label background positioned behind the text
	r.labelBackground.Resize(fyne.NewSize(labelWidth, labelHeight))
	r.labelBackground.Move(fyne.NewPos(labelX, labelY))

	// Label positioned over the background
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
	canvas.Refresh(r.content)
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
	case ".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv":
		return MediaTypeVideo
	default:
		return MediaTypeFile
	}
}

// ensureVideoThumbnail generates a video thumbnail if it doesn't exist
func (mc *MediaCard) ensureVideoThumbnail(videoPath, thumbPath string) {
	fmt.Printf("[DEBUG] Checking for video thumbnail: %s\n", thumbPath)
	if _, err := os.Stat(thumbPath); err == nil {
		fmt.Printf("[DEBUG] Thumbnail already exists: %s\n", thumbPath)
		return // thumbnail exists
	}
	fmt.Printf("[DEBUG] Thumbnail not found, generating for: %s\n", videoPath)
	_ = os.MkdirAll(filepath.Dir(thumbPath), 0755)

	// Use uniform 200x200 thumbnail generation
	cmd := []string{"ffmpeg", "-y", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1",
		"-vf", "scale=200:200:force_original_aspect_ratio=decrease,pad=200:200:(ow-iw)/2:(oh-ih)/2",
		thumbPath}
	fmt.Printf("[DEBUG] Generating uniform video thumbnail: %v\n", cmd)
	err := mc.runCommand(cmd)
	if err != nil {
		fmt.Printf("[DEBUG] ffmpeg error: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] ffmpeg thumbnail generated: %s\n", thumbPath)
	}
}

func (mc *MediaCard) runCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}
	return exec.Command(args[0], args[1:]...).Run()
}
