package views

import (
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/ui/components"
)

const (
	dragGhostOffsetX = float32(16)
	dragGhostOffsetY = float32(8)
	dragGhostPadX    = float32(10)
	dragGhostPadY    = float32(6)
	dragGhostMaxName = 36
)

func (n *folderNode) setDropHighlight(on bool) {
	if n == nil {
		return
	}
	if on {
		n.Importance = widget.HighImportance
		n.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		n.Importance = widget.MediumImportance
		n.TextStyle = fyne.TextStyle{}
	}
	n.Refresh()
}

type dragLayer struct {
	widget.BaseWidget
	view *MainView
}

var _ fyne.Scrollable = (*dragLayer)(nil)

func newDragLayer(v *MainView) *dragLayer {
	d := &dragLayer{view: v}
	d.ExtendBaseWidget(d)
	return d
}

func (d *dragLayer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (d *dragLayer) Scrolled(ev *fyne.ScrollEvent) {
	if d.view != nil {
		d.view.forwardScroll(ev)
	}
}

type dragLayerRenderer struct {
	layer *dragLayer
	objs  []fyne.CanvasObject
}

func (r *dragLayerRenderer) Layout(fyne.Size)             {}
func (r *dragLayerRenderer) MinSize() fyne.Size           { return fyne.NewSize(0, 0) }
func (r *dragLayerRenderer) Refresh()                     {}
func (r *dragLayerRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *dragLayerRenderer) Destroy()                     {}

func (d *dragLayer) CreateRenderer() fyne.WidgetRenderer {
	d.view.ensureDragGhost()
	return &dragLayerRenderer{layer: d, objs: []fyne.CanvasObject{d.view.dragGhostBg, d.view.dragGhostLabel}}
}

func (v *MainView) wrapWithDragGhost(content fyne.CanvasObject) fyne.CanvasObject {
	v.ensureDragGhost()
	return container.NewStack(content, newDragLayer(v))
}

func (v *MainView) forwardScroll(ev *fyne.ScrollEvent) {
	if v == nil || ev == nil {
		return
	}
	targets := []*container.Scroll{v.mediaGridWrapper, v.treeScroll, v.tagsScroll}
	for _, scroll := range targets {
		if scrollContainsAbs(scroll, ev.AbsolutePosition) {
			scroll.Scrolled(ev)
			v.snapshotMediaScroll()
			return
		}
	}
	if v.mediaGridWrapper != nil {
		v.mediaGridWrapper.Scrolled(ev)
	}
	v.snapshotMediaScroll()
}

func scrollContainsAbs(scroll *container.Scroll, abs fyne.Position) bool {
	if scroll == nil {
		return false
	}
	app := fyne.CurrentApp()
	if app == nil {
		return false
	}
	drv := app.Driver()
	if drv == nil || drv.CanvasForObject(scroll) == nil {
		return false
	}
	size := scroll.Size()
	if size.Width <= 0 || size.Height <= 0 {
		return false
	}
	pos := drv.AbsolutePositionForObject(scroll)
	return abs.X >= pos.X && abs.X < pos.X+size.Width &&
		abs.Y >= pos.Y && abs.Y < pos.Y+size.Height
}

func (v *MainView) snapshotMediaScroll() {
	if v == nil || v.mediaGridWrapper == nil {
		return
	}
	// Only overwrite the remembered offset when the widget still has a real one.
	// After a clamp-to-zero layout we must keep the previous snapshot.
	if v.lastScrollDir == v.mediaDir || v.lastScrollDir == "" {
		if v.mediaGridWrapper.Offset.Y > 0 || v.mediaGridWrapper.Offset.X > 0 {
			v.mediaScrollOffset = v.mediaGridWrapper.Offset
		}
	}
}

func (v *MainView) forgetMediaScroll() {
	if v == nil {
		return
	}
	v.mediaScrollOffset = fyne.NewPos(0, 0)
}

func applyScrollOffset(scroll *container.Scroll, offset fyne.Position) {
	if scroll == nil {
		return
	}
	scroll.Offset = offset
	scroll.Refresh()
}

func (v *MainView) restoreMediaScroll() {
	if v == nil || v.mediaGridWrapper == nil {
		return
	}
	if v.mediaDir == "" || (v.lastScrollDir != "" && v.lastScrollDir != v.mediaDir) {
		return
	}
	applyScrollOffset(v.mediaGridWrapper, v.mediaScrollOffset)
	off := v.mediaScrollOffset
	fyne.Do(func() {
		if v.mediaGridWrapper == nil {
			return
		}
		if v.mediaDir == "" || (v.lastScrollDir != "" && v.lastScrollDir != v.mediaDir) {
			return
		}
		applyScrollOffset(v.mediaGridWrapper, off)
	})
}

func (v *MainView) ensureDragGhost() {
	if v.dragGhostBg != nil && v.dragGhostLabel != nil {
		return
	}
	bg := canvas.NewRectangle(color.NRGBA{30, 30, 36, 235})
	bg.CornerRadius = 14
	bg.StrokeColor = color.NRGBA{120, 180, 255, 220}
	bg.StrokeWidth = 1
	bg.Hide()
	label := canvas.NewText("", color.White)
	label.TextSize = 12
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Hide()
	v.dragGhostBg = bg
	v.dragGhostLabel = label
}

func (v *MainView) wireCardDrag(card *components.MediaCard, filePath string) {
	if card == nil {
		return
	}
	name := filepath.Base(filePath)
	card.SetOnDragStart(func() {
		v.beginCardDrag(name)
	})
	card.SetOnDragged(func(abs fyne.Position) {
		v.updateCardDrag(abs)
	})
	card.SetOnDragEnd(func(dropPos fyne.Position) {
		v.endCardDrag()
		v.handleCardDrop(filePath, dropPos)
	})
}

func (v *MainView) beginCardDrag(name string) {
	v.ensureDragGhost()
	if v.dragGhostLabel != nil {
		v.dragGhostLabel.Text = truncateDragName(name)
		v.dragGhostLabel.Refresh()
	}
}

func (v *MainView) updateCardDrag(abs fyne.Position) {
	v.moveDragGhost(abs)
	v.highlightFolderAt(abs)
}

func (v *MainView) endCardDrag() {
	v.hideDragGhost()
	v.clearFolderHighlight()
}

func (v *MainView) moveDragGhost(abs fyne.Position) {
	v.ensureDragGhost()
	if v.dragGhostBg == nil || v.dragGhostLabel == nil {
		return
	}
	text := v.dragGhostLabel.Text
	if text == "" {
		text = " "
	}
	textSize := fyne.MeasureText(text, v.dragGhostLabel.TextSize, v.dragGhostLabel.TextStyle)
	w := textSize.Width + dragGhostPadX*2
	h := textSize.Height + dragGhostPadY*2
	pos := fyne.NewPos(abs.X+dragGhostOffsetX, abs.Y+dragGhostOffsetY)
	v.dragGhostBg.Resize(fyne.NewSize(w, h))
	v.dragGhostBg.Move(pos)
	v.dragGhostLabel.Move(fyne.NewPos(pos.X+dragGhostPadX, pos.Y+dragGhostPadY))
	v.dragGhostBg.Show()
	v.dragGhostLabel.Show()
	v.dragGhostBg.Refresh()
	v.dragGhostLabel.Refresh()
}

func (v *MainView) hideDragGhost() {
	if v.dragGhostBg != nil {
		v.dragGhostBg.Hide()
		v.dragGhostBg.Move(fyne.NewPos(0, 0))
		v.dragGhostBg.Resize(fyne.NewSize(1, 1))
		v.dragGhostBg.Refresh()
	}
	if v.dragGhostLabel != nil {
		v.dragGhostLabel.Hide()
		v.dragGhostLabel.Move(fyne.NewPos(0, 0))
		v.dragGhostLabel.Resize(fyne.NewSize(1, 1))
		v.dragGhostLabel.Refresh()
	}
}

func (v *MainView) highlightFolderAt(abs fyne.Position) {
	v.setHighlightedFolder(v.folderIDAt(abs))
}

func (v *MainView) setHighlightedFolder(id string) {
	v.dragMu.Lock()
	if v.highlightedFolder == id {
		v.dragMu.Unlock()
		return
	}
	v.highlightedFolder = id
	nodes := make([]*folderNode, len(v.folderNodes))
	copy(nodes, v.folderNodes)
	v.dragMu.Unlock()

	for _, n := range nodes {
		if n == nil {
			continue
		}
		n.setDropHighlight(id != "" && n.id == id)
	}
}

func (v *MainView) clearFolderHighlight() {
	v.setHighlightedFolder("")
}

func truncateDragName(name string) string {
	if len(name) <= dragGhostMaxName {
		return name
	}
	return name[:dragGhostMaxName-3] + "..."
}
