package components

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// AnimatedFrames is a widget that cycles through a sequence of images
// Used for video preview animation on hover
type AnimatedFrames struct {
	widget.BaseWidget

	framePaths   []string        // Paths to frame files
	currentFrame int             // Current frame index
	fps          int             // Frames per second for animation
	running      bool            // Whether animation is running
	stopChan     chan struct{}   // Channel to signal stop
	mu           sync.Mutex      // Protects running state
	
	// Single image that we update the file path on
	displayImage *canvas.Image
}

// NewAnimatedFrames creates a new animated frames widget from a list of image paths
func NewAnimatedFrames(framePaths []string) *AnimatedFrames {
	if len(framePaths) == 0 {
		return nil
	}

	// Create single display image starting with first frame
	img := canvas.NewImageFromFile(framePaths[0])
	img.FillMode = canvas.ImageFillContain
	
	af := &AnimatedFrames{
		framePaths:   framePaths,
		currentFrame: 0,
		fps:          8, // 8 FPS for smooth but efficient animation
		running:      false,
		displayImage: img,
	}

	af.ExtendBaseWidget(af)
	return af
}

// SetFPS sets the animation frame rate
func (af *AnimatedFrames) SetFPS(fps int) {
	af.fps = fps
}

// Start begins the animation loop
func (af *AnimatedFrames) Start() {
	af.mu.Lock()
	if af.running || len(af.framePaths) <= 1 {
		af.mu.Unlock()
		return
	}
	af.running = true
	af.stopChan = make(chan struct{})
	af.mu.Unlock()

	fmt.Printf("[DEBUG] AnimatedFrames: Starting animation with %d frames at %d FPS\n", len(af.framePaths), af.fps)

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(af.fps))
		defer ticker.Stop()

		frameCount := 0
		for {
			select {
			case <-af.stopChan:
				fmt.Println("[DEBUG] AnimatedFrames: Animation stopped")
				return
			case <-ticker.C:
				af.mu.Lock()
				if !af.running {
					af.mu.Unlock()
					return
				}
				af.currentFrame = (af.currentFrame + 1) % len(af.framePaths)
				newPath := af.framePaths[af.currentFrame]
				af.mu.Unlock()

				frameCount++
				if frameCount <= 3 || frameCount%10 == 0 {
					fmt.Printf("[DEBUG] AnimatedFrames: Frame %d -> %s\n", af.currentFrame, filepath.Base(newPath))
				}

				// Update image by loading new resource - File property change doesn't trigger reload
				fyne.Do(func() {
					// Load image from file as resource to force refresh
					res, err := fyne.LoadResourceFromPath(newPath)
					if err == nil {
						af.displayImage.Resource = res
						af.displayImage.File = "" // Clear file to use resource
						af.displayImage.Refresh()
					}
				})
			}
		}
	}()
}

// Stop halts the animation
func (af *AnimatedFrames) Stop() {
	af.mu.Lock()
	defer af.mu.Unlock()

	if !af.running {
		return
	}

	af.running = false
	close(af.stopChan)
	
	// Reset to first frame
	af.currentFrame = 0
	if len(af.framePaths) > 0 {
		af.displayImage.File = af.framePaths[0]
	}
}

// IsRunning returns whether the animation is currently playing
func (af *AnimatedFrames) IsRunning() bool {
	af.mu.Lock()
	defer af.mu.Unlock()
	return af.running
}

// CurrentImage returns the display image
func (af *AnimatedFrames) CurrentImage() *canvas.Image {
	return af.displayImage
}

// FirstFrame returns the first frame path
func (af *AnimatedFrames) FirstFramePath() string {
	if len(af.framePaths) == 0 {
		return ""
	}
	return af.framePaths[0]
}

// FrameCount returns the number of frames
func (af *AnimatedFrames) FrameCount() int {
	return len(af.framePaths)
}

// MinSize returns the minimum size of the widget
func (af *AnimatedFrames) MinSize() fyne.Size {
	return fyne.NewSize(200, 150) // Default size, will be resized by parent
}

// CreateRenderer returns the renderer for this widget
func (af *AnimatedFrames) CreateRenderer() fyne.WidgetRenderer {
	return &animatedFramesRenderer{
		af:      af,
		objects: []fyne.CanvasObject{af.displayImage},
	}
}

// animatedFramesRenderer implements fyne.WidgetRenderer
type animatedFramesRenderer struct {
	af      *AnimatedFrames
	objects []fyne.CanvasObject
}

func (r *animatedFramesRenderer) Layout(size fyne.Size) {
	if r.af.displayImage != nil {
		r.af.displayImage.Resize(size)
		r.af.displayImage.Move(fyne.NewPos(0, 0))
	}
}

func (r *animatedFramesRenderer) MinSize() fyne.Size {
	return r.af.MinSize()
}

func (r *animatedFramesRenderer) Refresh() {
	if r.af.displayImage != nil {
		canvas.Refresh(r.af.displayImage)
	}
}

func (r *animatedFramesRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *animatedFramesRenderer) Destroy() {
	r.af.Stop()
}
