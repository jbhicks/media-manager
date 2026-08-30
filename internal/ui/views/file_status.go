package views

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (v *MainView) ensureStatusBar() fyne.CanvasObject {
	if v.statusBar != nil {
		return v.statusBar
	}
	v.moveStatusLabel = widget.NewLabel("")
	v.moveStatusLabel.Truncation = fyne.TextTruncateEllipsis
	v.moveProgressBar = widget.NewProgressBar()
	v.moveProgressBar.Min = 0
	v.moveProgressBar.Max = 1
	v.moveProgressBar.Hide()
	inner := container.NewBorder(nil, nil, nil, v.moveProgressBar, v.moveStatusLabel)
	v.statusBar = container.NewPadded(inner)
	v.statusBar.Hide()
	return v.statusBar
}

func (v *MainView) beginMoveProgress(text string) {
	v.ensureStatusBar()
	v.moveStatusLabel.SetText(text)
	v.moveProgressBar.SetValue(0)
	v.moveProgressBar.Show()
	v.statusBar.Show()
	v.statusBar.Refresh()
}

func (v *MainView) updateMoveProgress(frac float64) {
	if v.moveProgressBar == nil {
		return
	}
	fyne.Do(func() {
		v.moveProgressBar.SetValue(frac)
	})
}

func (v *MainView) endMoveProgress() {
	fyne.Do(func() {
		if v.moveProgressBar != nil {
			v.moveProgressBar.SetValue(1)
			v.moveProgressBar.Hide()
		}
		if v.moveStatusLabel != nil {
			v.moveStatusLabel.SetText("")
		}
		if v.statusBar != nil {
			v.statusBar.Hide()
			v.statusBar.Refresh()
		}
	})
}

func (v *MainView) wrapRoot(inner fyne.CanvasObject) fyne.CanvasObject {
	return v.wrapWithDragGhost(container.NewBorder(nil, v.ensureStatusBar(), nil, nil, inner))
}
