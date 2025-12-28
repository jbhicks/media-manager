package components

import (
	"io"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type DebugWriter struct {
	textGrid *widget.TextGrid
	mu       sync.Mutex
}

func NewDebugWriter(tg *widget.TextGrid) *DebugWriter {
	return &DebugWriter{
		textGrid: tg,
	}
}

func (w *DebugWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Use Fyne's thread-safe Do function for UI updates
	fyne.Do(func() {
		w.textGrid.SetText(w.textGrid.Text() + string(p))
		w.textGrid.Refresh()
	})

	return len(p), nil
}

var _ io.Writer = (*DebugWriter)(nil)
