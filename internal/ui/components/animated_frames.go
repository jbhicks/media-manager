package components

import (
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

	framePaths     []string        // Paths to frame files
	frameResources []fyne.Resource // Pre-loaded frame resources
	currentFrame   int             // Current frame index
	fps            int             // Frames per second for animation
	running        bool            // Whether animation is running
	stopChan       chan struct{}   // Channel to signal stop
	mu             sync.Mutex      // Protects running state

	// Single image that we update the resource on
	displayImage *canvas.Image
}

// NewAnimatedFrames creates a new animated frames widget from pre-loaded resources
// IMPORTANT: Use NewAnimatedFramesAsync for loading from file paths to avoid UI blocking
func NewAnimatedFrames(framePaths []string) *AnimatedFrames {
	if len(framePaths) == 0 {
		return nil
	}

	// Load resources - this blocks, so caller should use NewAnimatedFramesAsync instead
	frameResources := loadFrameResources(framePaths)
	if len(frameResources) == 0 {
		return nil
	}

	return newAnimatedFramesFromResources(framePaths, frameResources)
}

// NewAnimatedFramesAsync loads frame resources in background and calls callback with result
// This avoids blocking the UI thread during disk I/O
func NewAnimatedFramesAsync(framePaths []string, callback func(*AnimatedFrames)) {
	if len(framePaths) == 0 {
		callback(nil)
		return
	}

	go func() {
		// Load resources in background goroutine (NOT on UI thread)
		frameResources := loadFrameResources(framePaths)
		if len(frameResources) == 0 {
			fyne.Do(func() { callback(nil) })
			return
		}

		// Create widget on UI thread
		fyne.Do(func() {
			af := newAnimatedFramesFromResources(framePaths, frameResources)
			callback(af)
		})
	}()
}

// loadFrameResources loads all frame images from disk (blocking operation)
func loadFrameResources(framePaths []string) []fyne.Resource {
	frameResources := make([]fyne.Resource, 0, len(framePaths))
	for _, path := range framePaths {
		res, err := fyne.LoadResourceFromPath(path)
		if err != nil {
			continue
		}
		frameResources = append(frameResources, res)
	}
	return frameResources
}

// newAnimatedFramesFromResources creates AnimatedFrames from pre-loaded resources
func newAnimatedFramesFromResources(framePaths []string, frameResources []fyne.Resource) *AnimatedFrames {
	// Create single display image starting with first frame resource
	img := &canvas.Image{Resource: frameResources[0]}
	img.FillMode = canvas.ImageFillContain

	af := &AnimatedFrames{
		framePaths:     framePaths,
		frameResources: frameResources,
		currentFrame:   0,
		fps:            8, // 8 FPS for smooth but efficient animation
		running:        false,
		displayImage:   img,
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
	if af.running || len(af.frameResources) <= 1 {
		af.mu.Unlock()
		return
	}
	af.running = true
	af.stopChan = make(chan struct{})
	af.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(af.fps))
		defer ticker.Stop()

		for {
			select {
			case <-af.stopChan:
				return
			case <-ticker.C:
				af.mu.Lock()
				if !af.running {
					af.mu.Unlock()
					return
				}
				af.currentFrame = (af.currentFrame + 1) % len(af.frameResources)
				res := af.frameResources[af.currentFrame]
				af.mu.Unlock()

				// Update image using pre-loaded resource - no disk I/O!
				fyne.Do(func() {
					af.displayImage.Resource = res
					af.displayImage.Refresh()
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
	if len(af.frameResources) > 0 {
		af.displayImage.Resource = af.frameResources[0]
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
	return len(af.frameResources)
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
