package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"os"
)

// VideoPreviewCard represents a video thumbnail with animated GIF hover preview
type VideoPreviewCard struct {
	widget.BaseWidget
	staticImage     *canvas.Image
	animatedGif     *xwidget.AnimatedGif
	label           *widget.Label
	container       *fyne.Container
	isHovered       bool
	hasAnimation    bool
	animatedGifPath string
}

// NewVideoPreviewCard creates a new video preview card using the Fyne-X AnimatedGif widget
func NewVideoPreviewCard(staticImagePath, animatedGifPath, labelText string) *VideoPreviewCard {
	fmt.Printf("[DEBUG] NewVideoPreviewCard called - static: %s, animated: %s\n", staticImagePath, animatedGifPath)

	// Load static thumbnail
	staticImg := canvas.NewImageFromFile(staticImagePath)
	staticImg.FillMode = canvas.ImageFillContain
	staticImg.SetMinSize(fyne.NewSize(100, 80))

	// Try to load animated GIF using Fyne-X widget
	var animatedGif *xwidget.AnimatedGif
	hasAnimation := false

	if _, err := os.Stat(animatedGifPath); err == nil {
		// Create animated GIF widget using Fyne-X
		gifURI := storage.NewFileURI(animatedGifPath)
		gif, err := xwidget.NewAnimatedGif(gifURI)
		if err == nil {
			gif.SetMinSize(fyne.NewSize(100, 80))
			animatedGif = gif
			hasAnimation = true
			fmt.Printf("[DEBUG] Successfully loaded animated GIF: %s\n", animatedGifPath)
		} else {
			fmt.Printf("[DEBUG] Failed to load animated GIF: %v\n", err)
		}
	} else {
		fmt.Printf("[DEBUG] No animated GIF found: %s\n", animatedGifPath)
	}

	label := widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})

	// Create container with static image initially
	cont := container.NewVBox(staticImg, label)

	card := &VideoPreviewCard{
		staticImage:     staticImg,
		animatedGif:     animatedGif,
		label:           label,
		container:       cont,
		isHovered:       false,
		hasAnimation:    hasAnimation,
		animatedGifPath: animatedGifPath,
	}
	card.ExtendBaseWidget(card)
	return card
}

// Ensure VideoPreviewCard implements desktop.Hoverable
var _ desktop.Hoverable = (*VideoPreviewCard)(nil)

// MouseIn starts the animated GIF on hover
func (vpc *VideoPreviewCard) MouseIn(*desktop.MouseEvent) {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}

	fmt.Println("[DEBUG] VideoPreviewCard MouseIn - starting animated GIF")
	vpc.isHovered = true

	// Replace the static image with animated GIF
	vpc.container.Objects[0] = vpc.animatedGif
	vpc.container.Refresh()
}

// MouseOut stops the animated GIF when hover ends
func (vpc *VideoPreviewCard) MouseOut() {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}

	fmt.Println("[DEBUG] VideoPreviewCard MouseOut - stopping animated GIF")
	vpc.isHovered = false

	// Stop animation and replace with static image
	vpc.animatedGif.Stop()
	vpc.container.Objects[0] = vpc.staticImage
	vpc.container.Refresh()
}

// MouseMoved handles mouse movement (required by desktop.Hoverable)
func (vpc *VideoPreviewCard) MouseMoved(*desktop.MouseEvent) {
	// No action needed
}

// Tapped handles tap events (for mobile/touch support)
func (vpc *VideoPreviewCard) Tapped(*fyne.PointEvent) {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}

	fmt.Println("[DEBUG] VideoPreviewCard tapped - toggling animation")

	if vpc.isHovered {
		vpc.MouseOut()
	} else {
		vpc.MouseIn(nil)
	}
}

// MinSize returns the minimum size for the card
func (vpc *VideoPreviewCard) MinSize() fyne.Size {
	return fyne.NewSize(120, 120)
}

// CreateRenderer creates the renderer for the video preview card
func (vpc *VideoPreviewCard) CreateRenderer() fyne.WidgetRenderer {
	return &videoPreviewCardRenderer{
		card: vpc,
	}
}

// videoPreviewCardRenderer renders the video preview card
type videoPreviewCardRenderer struct {
	card *VideoPreviewCard
}

func (r *videoPreviewCardRenderer) Layout(size fyne.Size) {
	r.card.container.Resize(size)
}

func (r *videoPreviewCardRenderer) MinSize() fyne.Size {
	return r.card.container.MinSize()
}

func (r *videoPreviewCardRenderer) Refresh() {
	r.card.container.Refresh()
}

func (r *videoPreviewCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.card.container}
}

func (r *videoPreviewCardRenderer) Destroy() {
	// Stop any running animation
	if r.card.animatedGif != nil {
		r.card.animatedGif.Stop()
	}
}
