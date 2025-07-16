package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

// HoverableCard wraps a HoverableImage in a container that can detect hover events
type HoverableCard struct {
	widget.BaseWidget
	animatedGif *xwidget.AnimatedGif
	label       *widget.Label
	isHovered   bool
}

// NewHoverableCard creates a card that contains a hoverable image and label
func NewHoverableCard(gifPath, labelText string) *HoverableCard {
	// [DEBUG] NewHoverableCard called

	uri, err := storage.ParseURI(fmt.Sprintf("file://%s", gifPath))
	if err != nil {
		fmt.Printf("Error parsing GIF URI for card: %v\n", err)
		return nil
	}

	animatedGif, err := xwidget.NewAnimatedGif(uri)
	if err != nil {
		fmt.Printf("Error creating animated GIF for card: %v\n", err)
		// Handle error, perhaps return a card with a placeholder image
		return nil
	}

	label := widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})

	card := &HoverableCard{
		animatedGif: animatedGif,
		label:       label,
	}
	card.ExtendBaseWidget(card)
	return card
}

// Ensure that HoverableCard implements desktop.Hoverable
var _ desktop.Hoverable = (*HoverableCard)(nil)

// MouseIn is called when the mouse enters the card area.
func (hc *HoverableCard) MouseIn(*desktop.MouseEvent) {
	// [DEBUG] !!!! HoverableCard MouseIn event triggered !!!!
	hc.isHovered = true
	if hc.animatedGif != nil {
		hc.animatedGif.Start()
	}
	hc.Refresh()
}

// MouseOut is called when the mouse leaves the card area.
func (hc *HoverableCard) MouseOut() {
	// [DEBUG] !!!! HoverableCard MouseOut event triggered !!!!
	hc.isHovered = false
	if hc.animatedGif != nil {
		hc.animatedGif.Stop()
	}
	hc.Refresh()
}

// MouseMoved is called when the mouse moves within the card area.
func (hc *HoverableCard) MouseMoved(*desktop.MouseEvent) {
	// Optional: handle mouse movement if needed
}

// MinSize returns the minimum size for the card
func (hc *HoverableCard) MinSize() fyne.Size {
	if hc.animatedGif != nil {
		return hc.animatedGif.MinSize()
	}
	return fyne.NewSize(120, 120)
}

// CreateRenderer creates the renderer for the hoverable card
func (hc *HoverableCard) CreateRenderer() fyne.WidgetRenderer {
	return &hoverableCardRenderer{
		card:  hc,
		gif:   hc.animatedGif,
		label: hc.label,
	}
}

// hoverableCardRenderer renders the hoverable card
type hoverableCardRenderer struct {
	card  *HoverableCard
	gif   *xwidget.AnimatedGif
	label *widget.Label
}

func (r *hoverableCardRenderer) Layout(size fyne.Size) {
	// Layout GIF at the top, label at the bottom
	gifHeight := size.Height * 0.8   // 80% of the card height for GIF
	labelHeight := size.Height * 0.2 // 20% for label

	if r.gif != nil {
		r.gif.Resize(fyne.NewSize(size.Width, gifHeight))
		r.gif.Move(fyne.NewPos(0, 0))
	}

	r.label.Resize(fyne.NewSize(size.Width, labelHeight))
	r.label.Move(fyne.NewPos(0, gifHeight))
}

func (r *hoverableCardRenderer) MinSize() fyne.Size {
	if r.gif != nil {
		return r.gif.MinSize()
	}
	return fyne.NewSize(120, 120)
}

func (r *hoverableCardRenderer) Refresh() {
	if r.gif != nil {
		r.gif.Refresh()
	}
	r.label.Refresh()
}

func (r *hoverableCardRenderer) Objects() []fyne.CanvasObject {
	if r.gif != nil {
		return []fyne.CanvasObject{r.gif, r.label}
	}
	return []fyne.CanvasObject{r.label}
}

func (r *hoverableCardRenderer) Destroy() {
	if r.gif != nil {
		r.gif.Stop()
	}
}
