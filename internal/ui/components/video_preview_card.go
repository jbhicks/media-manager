package components

import (
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

type VideoPreviewCard struct {
	widget.BaseWidget
	staticImage     *canvas.Image
	animatedGif     *xwidget.AnimatedGif
	label           *widget.Label
	container       *fyne.Container
	background      *canvas.Rectangle
	labelBackground *canvas.Rectangle
	isHovered       bool
	hasAnimation    bool
	animatedGifPath string
}

// MouseIn starts the animated GIF on hover
func (vpc *VideoPreviewCard) MouseIn(*desktop.MouseEvent) {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}
	vpc.isHovered = true
	gifContainer := container.NewGridWrap(fyne.NewSize(100, 80), vpc.animatedGif)
	vpc.container.Objects[0] = gifContainer
	vpc.container.Refresh()
}

// MouseOut stops the animated GIF when hover ends
func (vpc *VideoPreviewCard) MouseOut() {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}
	vpc.isHovered = false
	vpc.animatedGif.Stop()
	imgContainer := container.NewGridWrap(fyne.NewSize(100, 80), vpc.staticImage)
	vpc.container.Objects[0] = imgContainer
	vpc.container.Refresh()
}

// MouseMoved handles mouse movement (required by desktop.Hoverable)
func (vpc *VideoPreviewCard) MouseMoved(*desktop.MouseEvent) {}

// Tapped handles tap events (for mobile/touch support)
func (vpc *VideoPreviewCard) Tapped(*fyne.PointEvent) {
	if !vpc.hasAnimation || vpc.animatedGif == nil {
		return
	}
	if vpc.isHovered {
		vpc.MouseOut()
	} else {
		vpc.MouseIn(nil)
	}
}

// CreateRenderer creates the renderer for the video preview card
func (vpc *VideoPreviewCard) CreateRenderer() fyne.WidgetRenderer {
	return &videoPreviewCardRenderer{
		card: vpc,
	}
}

type videoPreviewCardRenderer struct {
	card *VideoPreviewCard
}

func (r *videoPreviewCardRenderer) Layout(size fyne.Size) {
	padding := float32(4)
	w, h := 0, 0
	if r.card.staticImage != nil && r.card.staticImage.File != "" {
		if file, err := os.Open(r.card.staticImage.File); err == nil {
			defer file.Close()
			if cfg, _, err := image.DecodeConfig(file); err == nil {
				w = cfg.Width
				h = cfg.Height
			}
		}
	}
	maxW, maxH := float32(180), float32(120)
	contentW, contentH := maxW, maxH
	if w > 0 && h > 0 {
		aspect := float32(w) / float32(h)
		if aspect > 1 {
			contentW = min(maxW, float32(w))
			contentH = contentW / aspect
			if contentH > maxH {
				contentH = maxH
				contentW = maxH * aspect
			}
		} else {
			contentH = min(maxH, float32(h))
			contentW = contentH * aspect
			if contentW > maxW {
				contentW = maxW
				contentH = maxW / aspect
			}
		}
	}
	labelSize := r.card.label.MinSize()

	r.card.background.Resize(fyne.NewSize(contentW+2*padding, contentH+labelSize.Height+3*padding))
	r.card.background.Move(fyne.NewPos(0, 0))

	if r.card.container != nil {
		r.card.container.Resize(fyne.NewSize(contentW, contentH))
		r.card.container.Move(fyne.NewPos(padding, padding))
	}

	labelX := padding
	labelY := padding + contentH + padding
	labelWidth := contentW
	labelHeight := labelSize.Height

	if r.card.labelBackground != nil {
		r.card.labelBackground.Resize(fyne.NewSize(labelWidth, labelHeight))
		r.card.labelBackground.Move(fyne.NewPos(labelX, labelY))
	}
	if r.card.label != nil {
		r.card.label.Resize(fyne.NewSize(labelWidth, labelHeight))
		r.card.label.Move(fyne.NewPos(labelX, labelY))
	}
}

func (r *videoPreviewCardRenderer) MinSize() fyne.Size {
	w, h := 0, 0
	if r.card.staticImage != nil && r.card.staticImage.File != "" {
		if file, err := os.Open(r.card.staticImage.File); err == nil {
			defer file.Close()
			if cfg, _, err := image.DecodeConfig(file); err == nil {
				w = cfg.Width
				h = cfg.Height
			}
		}
	}
	maxW, maxH := float32(180), float32(120)
	contentW, contentH := maxW, maxH
	if w > 0 && h > 0 {
		aspect := float32(w) / float32(h)
		if aspect > 1 {
			contentW = min(maxW, float32(w))
			contentH = contentW / aspect
			if contentH > maxH {
				contentH = maxH
				contentW = maxH * aspect
			}
		} else {
			contentH = min(maxH, float32(h))
			contentW = contentH * aspect
			if contentW > maxW {
				contentW = maxW
				contentH = maxW / aspect
			}
		}
	}
	labelSize := r.card.label.MinSize()
	padding := float32(4)
	return fyne.NewSize(contentW+2*padding, contentH+labelSize.Height+3*padding)
}

func (r *videoPreviewCardRenderer) Refresh() {
	if r.card.container != nil {
		r.card.container.Refresh()
	}
}

func (r *videoPreviewCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.card.background, r.card.container, r.card.labelBackground}
}

func (r *videoPreviewCardRenderer) Destroy() {
	if r.card.animatedGif != nil {
		r.card.animatedGif.Stop()
	}
}
