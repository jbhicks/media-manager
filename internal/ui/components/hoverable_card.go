package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverableCard wraps a HoverableImage in a container that can detect hover events
type HoverableCard struct {
	widget.BaseWidget
	hoverableImage *HoverableImage
	label          *widget.Label
	isHovered      bool
}

// NewHoverableCard creates a card that contains a hoverable image and label
func NewHoverableCard(staticImage *canvas.Image, animatedImage *canvas.Image, labelText string) *HoverableCard {
	fmt.Println("[DEBUG] NewHoverableCard called")

	hoverableImg := &HoverableImage{
		StaticImage:   staticImage,
		AnimatedImage: animatedImage,
	}
	hoverableImg.ExtendBaseWidget(hoverableImg)

	label := widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})

	card := &HoverableCard{
		hoverableImage: hoverableImg,
		label:          label,
	}
	card.ExtendBaseWidget(card)
	return card
}

// Ensure that HoverableCard implements desktop.Hoverable
var _ desktop.Hoverable = (*HoverableCard)(nil)

// MouseIn is called when the mouse enters the card area.
func (hc *HoverableCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] !!!! HoverableCard MouseIn event triggered !!!!")
	hc.isHovered = true
	hc.hoverableImage.IsHovered = true
	fmt.Printf("[DEBUG] HoverableCard set IsHovered to: %v\n", hc.hoverableImage.IsHovered)
	hc.Refresh()
}

// MouseOut is called when the mouse leaves the card area.
func (hc *HoverableCard) MouseOut() {
	fmt.Println("[DEBUG] !!!! HoverableCard MouseOut event triggered !!!!")
	hc.isHovered = false
	hc.hoverableImage.IsHovered = false
	fmt.Printf("[DEBUG] HoverableCard set IsHovered to: %v\n", hc.hoverableImage.IsHovered)
	hc.Refresh()
}

// MouseMoved is called when the mouse moves within the card area.
func (hc *HoverableCard) MouseMoved(*desktop.MouseEvent) {
	// Optional: handle mouse movement if needed
}

// MinSize returns the minimum size for the card
func (hc *HoverableCard) MinSize() fyne.Size {
	return fyne.NewSize(120, 120)
}

// CreateRenderer creates the renderer for the hoverable card
func (hc *HoverableCard) CreateRenderer() fyne.WidgetRenderer {
	return &hoverableCardRenderer{
		card:  hc,
		image: hc.hoverableImage,
		label: hc.label,
	}
}

// hoverableCardRenderer renders the hoverable card
type hoverableCardRenderer struct {
	card  *HoverableCard
	image *HoverableImage
	label *widget.Label
}

func (r *hoverableCardRenderer) Layout(size fyne.Size) {
	// Layout image at the top, label at the bottom
	imageHeight := size.Height * 0.8 // 80% of the card height for image
	labelHeight := size.Height * 0.2 // 20% for label

	r.image.Resize(fyne.NewSize(size.Width, imageHeight))
	r.image.Move(fyne.NewPos(0, 0))

	r.label.Resize(fyne.NewSize(size.Width, labelHeight))
	r.label.Move(fyne.NewPos(0, imageHeight))
}

func (r *hoverableCardRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, 120)
}

func (r *hoverableCardRenderer) Refresh() {
	r.image.Refresh()
	r.label.Refresh()
}

func (r *hoverableCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.image, r.label}
}

func (r *hoverableCardRenderer) Destroy() {}
