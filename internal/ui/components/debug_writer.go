package components

import (
	"io"
	"sync"

	"fyne.io/fyne/v2/widget"
)

type DebugWriter struct {
	textGrid   *widget.TextGrid
	mu         sync.Mutex
	updateChan chan func()
}

func NewDebugWriter(tg *widget.TextGrid) *DebugWriter {
	writer := &DebugWriter{
		textGrid:   tg,
		updateChan: make(chan func(), 100), // Buffered channel to avoid blocking
	}
	go writer.processUpdates()
	return writer
}

func (w *DebugWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Send the update function to the channel
	w.updateChan <- func() {
		w.textGrid.SetText(w.textGrid.Text() + string(p))
		// Scroll to the bottom
		// This might require a custom scrollable container or more advanced TextGrid usage
	}
	return len(p), nil
}

func (w *DebugWriter) processUpdates() {
	for fn := range w.updateChan {
		fn()
	}
}

var _ io.Writer = (*DebugWriter)(nil)
