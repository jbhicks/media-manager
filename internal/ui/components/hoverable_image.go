package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverableImage is a custom widget that shows a static image by default and an animated GIF on hover.
type HoverableImage struct {
	widget.BaseWidget
	StaticImage   *canvas.Image
	AnimatedImage *canvas.Image
	IsHovered     bool
}

func (hi *HoverableImage) MinSize() fyne.Size {
	return fyne.NewSize(100, 80)
}

// NewHoverableImage creates a new HoverableImage widget.
func NewHoverableImage(staticImage *canvas.Image, animatedImage *canvas.Image) *HoverableImage {
	fmt.Println("[DEBUG] NewHoverableImage called")
	hi := &HoverableImage{
		StaticImage:   staticImage,
		AnimatedImage: animatedImage,
	}
	hi.ExtendBaseWidget(hi)
	return hi
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (hi *HoverableImage) CreateRenderer() fyne.WidgetRenderer {
	return &hoverableImageRenderer{
		hoverableImage: hi,
	}
}

// Ensure that HoverableImage implements desktop.Hoverable
var _ desktop.Hoverable = (*HoverableImage)(nil)

// MouseIn is called when the mouse enters the widget's area.
func (hi *HoverableImage) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] !!!! MouseIn event triggered !!!!")
	hi.IsHovered = true
	hi.Refresh()
}

// MouseOut is called when the mouse leaves the widget's area.
func (hi *HoverableImage) MouseOut() {
	fmt.Println("[DEBUG] !!!! MouseOut event triggered !!!!")
	hi.IsHovered = false
	hi.Refresh()
}

// MouseMoved is called when the mouse moves within the widget's area.
func (hi *HoverableImage) MouseMoved(*desktop.MouseEvent) {
	// Optional: handle mouse movement if needed
}

// Tapped is called when a regular tap is detected.
func (hi *HoverableImage) Tapped(*fyne.PointEvent) {
	// Handle tap events if needed
}

// TappedSecondary is called when a secondary tap is detected.
func (hi *HoverableImage) TappedSecondary(*fyne.PointEvent) {
	// Handle secondary tap events if needed
}

// hoverableImageRenderer is the renderer for the HoverableImage widget.
type hoverableImageRenderer struct {
	hoverableImage *HoverableImage
}

func (r *hoverableImageRenderer) Layout(size fyne.Size) {
	if r.hoverableImage.IsHovered && r.hoverableImage.AnimatedImage != nil {
		r.hoverableImage.AnimatedImage.Resize(size)
	} else if r.hoverableImage.StaticImage != nil {
		r.hoverableImage.StaticImage.Resize(size)
	}
}

func (r *hoverableImageRenderer) MinSize() fyne.Size {
	if r.hoverableImage.StaticImage != nil {
		return r.hoverableImage.StaticImage.MinSize()
	}
	return fyne.NewSize(100, 80)
}

func (r *hoverableImageRenderer) Refresh() {
	fmt.Printf("[DEBUG] HoverableImage Refresh called, IsHovered: %v\n", r.hoverableImage.IsHovered)
	if r.hoverableImage.IsHovered && r.hoverableImage.AnimatedImage != nil {
		fmt.Println("[DEBUG] Switching to animated image")
		r.hoverableImage.AnimatedImage.Refresh()
	} else if r.hoverableImage.StaticImage != nil {
		fmt.Println("[DEBUG] Switching to static image")
		r.hoverableImage.StaticImage.Refresh()
	}
}

func (r *hoverableImageRenderer) Objects() []fyne.CanvasObject {
	if r.hoverableImage.IsHovered {
		fmt.Println("[DEBUG] Objects() returning animated image")
		return []fyne.CanvasObject{r.hoverableImage.AnimatedImage}
	}
	return []fyne.CanvasObject{r.hoverableImage.StaticImage}
}

func (r *hoverableImageRenderer) Destroy() {}
