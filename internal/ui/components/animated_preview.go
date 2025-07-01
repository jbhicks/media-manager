package components

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"time"
)

// HoverableVideoCard wraps an AnimatedPreview with hover detection
type HoverableVideoCard struct {
	widget.BaseWidget
	animatedPreview *AnimatedPreview
	label           *widget.Label
	container       *fyne.Container
}

// NewHoverableVideoCard creates a new hoverable video card
func NewHoverableVideoCard(staticImagePath, videoPath, labelText string) *HoverableVideoCard {
	preview := NewAnimatedPreview(staticImagePath, videoPath)
	label := widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})

	// Create a container with the preview and label
	cont := container.NewVBox(preview, label)

	card := &HoverableVideoCard{
		animatedPreview: preview,
		label:           label,
		container:       cont,
	}
	card.ExtendBaseWidget(card)
	return card
}

// Ensure HoverableVideoCard implements desktop.Hoverable
var _ desktop.Hoverable = (*HoverableVideoCard)(nil)

// MouseIn is called when the mouse enters the card area
func (hvc *HoverableVideoCard) MouseIn(*desktop.MouseEvent) {
	fmt.Println("[DEBUG] !!!! HoverableVideoCard MouseIn - starting animation !!!!")
	hvc.animatedPreview.startHoverAnimation()
	// Refresh the container instead of individual components
	hvc.Refresh()
}

// MouseOut is called when the mouse leaves the card area
func (hvc *HoverableVideoCard) MouseOut() {
	fmt.Println("[DEBUG] !!!! HoverableVideoCard MouseOut - stopping animation !!!!")
	hvc.animatedPreview.stopAnimation()
}

// MouseMoved is called when the mouse moves within the card area
func (hvc *HoverableVideoCard) MouseMoved(*desktop.MouseEvent) {
	// Optional: handle mouse movement if needed
}

// Tapped handles tap events
func (hvc *HoverableVideoCard) Tapped(*fyne.PointEvent) {
	fmt.Println("[DEBUG] HoverableVideoCard tapped - starting animation")
	hvc.animatedPreview.startClickAnimation()
}

// MinSize returns the minimum size
func (hvc *HoverableVideoCard) MinSize() fyne.Size {
	return fyne.NewSize(120, 120)
}

// CreateRenderer creates the renderer
func (hvc *HoverableVideoCard) CreateRenderer() fyne.WidgetRenderer {
	return &hoverableVideoCardRenderer{
		card: hvc,
	}
}

// hoverableVideoCardRenderer renders the hoverable video card
type hoverableVideoCardRenderer struct {
	card *HoverableVideoCard
}

func (r *hoverableVideoCardRenderer) Layout(size fyne.Size) {
	r.card.container.Resize(size)
}

func (r *hoverableVideoCardRenderer) MinSize() fyne.Size {
	return r.card.container.MinSize()
}

func (r *hoverableVideoCardRenderer) Refresh() {
	r.card.container.Refresh()
}

func (r *hoverableVideoCardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.card.container}
}

func (r *hoverableVideoCardRenderer) Destroy() {
	r.card.animatedPreview.stopAnimation()
}

// AnimatedPreview creates a simple frame-based animation for video previews
type AnimatedPreview struct {
	widget.BaseWidget
	staticImagePath string
	frameFiles      []string
	displayImage    *canvas.Image
	currentFrame    int
	isAnimating     bool
	isHovered       bool
	frameCount      int
}

// NewAnimatedPreview creates a new animated preview widget
func NewAnimatedPreview(staticImagePath, videoPath string) *AnimatedPreview {
	fmt.Printf("[DEBUG] NewAnimatedPreview called for: %s\n", videoPath)

	// Load static image
	staticImg := canvas.NewImageFromFile(staticImagePath)
	staticImg.FillMode = canvas.ImageFillContain
	staticImg.SetMinSize(fyne.NewSize(100, 80))

	// Get frame file paths
	frameFiles := getFrameFiles(videoPath)

	ap := &AnimatedPreview{
		staticImagePath: staticImagePath,
		frameFiles:      frameFiles,
		displayImage:    staticImg,
		currentFrame:    0,
		isAnimating:     false,
		isHovered:       false,
		frameCount:      len(frameFiles),
	}
	ap.ExtendBaseWidget(ap)

	return ap
}

// getFrameFiles returns the file paths for animation frames
func getFrameFiles(videoPath string) []string {
	fmt.Printf("[DEBUG] Getting frame files for: %s\n", videoPath)

	baseFileName := filepath.Base(videoPath)
	homeDir, _ := os.UserHomeDir()
	frameDir := filepath.Join(homeDir, ".media-manager", "thumbnails", baseFileName+"_frames")

	// Check if we have a complete frame sequence
	markerPath := filepath.Join(frameDir, ".complete")
	if _, err := os.Stat(markerPath); err != nil {
		fmt.Printf("[DEBUG] No complete frame sequence found for: %s\n", videoPath)
		return nil
	}

	var frameFiles []string

	// Get paths for all 8 frames
	for i := 0; i < 8; i++ {
		framePath := filepath.Join(frameDir, fmt.Sprintf("frame_%d.jpg", i))
		if _, err := os.Stat(framePath); err == nil {
			frameFiles = append(frameFiles, framePath)
		} else {
			fmt.Printf("[DEBUG] Missing frame: %s\n", framePath)
			return nil
		}
	}

	fmt.Printf("[DEBUG] Successfully found %d frame files\n", len(frameFiles))
	return frameFiles
}

// Tapped handles tap events to trigger animation
func (ap *AnimatedPreview) Tapped(*fyne.PointEvent) {
	fmt.Println("[DEBUG] AnimatedPreview tapped - starting animation")
	ap.startClickAnimation()
}

// startClickAnimation begins a click-triggered animation (3 cycles)
func (ap *AnimatedPreview) startClickAnimation() {
	if len(ap.frameFiles) == 0 {
		fmt.Println("[DEBUG] No animation frames available")
		return
	}

	ap.stopAnimation() // Stop any existing animation
	ap.isAnimating = true
	ap.isHovered = false
	ap.currentFrame = 0
	ap.frameCount = 3 // 3 cycles for click

	ap.animateNextFrame()
}

// startHoverAnimation begins a hover-triggered animation (continuous loop)
func (ap *AnimatedPreview) startHoverAnimation() {
	if len(ap.frameFiles) == 0 {
		fmt.Println("[DEBUG] No animation frames available")
		return
	}

	if ap.isAnimating && ap.isHovered {
		return // Already animating from hover
	}

	ap.isAnimating = true
	ap.isHovered = true
	ap.currentFrame = 0
	ap.frameCount = -1 // Infinite loop for hover

	ap.animateNextFrame()
}

// animateNextFrame shows the next frame and schedules the next one
func (ap *AnimatedPreview) animateNextFrame() {
	if !ap.isAnimating || len(ap.frameFiles) == 0 {
		return
	}

	// Update to next frame
	framePath := ap.frameFiles[ap.currentFrame]
	ap.displayImage.File = framePath

	// Don't refresh here - let the parent container handle it
	fmt.Printf("[DEBUG] Showing frame %d: %s\n", ap.currentFrame, framePath)

	ap.currentFrame = (ap.currentFrame + 1) % len(ap.frameFiles)

	// Handle frame counting for non-hover animations
	if ap.currentFrame == 0 && ap.frameCount > 0 {
		ap.frameCount--
		if ap.frameCount <= 0 {
			ap.stopAnimation()
			return
		}
	}

	// Schedule next frame after 300ms
	time.AfterFunc(300*time.Millisecond, ap.animateNextFrame)
}

// stopAnimation stops the animation and returns to static image
func (ap *AnimatedPreview) stopAnimation() {
	fmt.Println("[DEBUG] Stopping animation")
	ap.isAnimating = false
	ap.isHovered = false
	ap.frameCount = 3 // Reset for next click animation

	// Return to static image
	ap.displayImage.File = ap.staticImagePath
	// Don't refresh here - let the parent container handle it
}

// MinSize returns the minimum size
func (ap *AnimatedPreview) MinSize() fyne.Size {
	return fyne.NewSize(100, 80)
}

// CreateRenderer creates the renderer
func (ap *AnimatedPreview) CreateRenderer() fyne.WidgetRenderer {
	return &animatedPreviewRenderer{
		preview: ap,
	}
}

// animatedPreviewRenderer renders the animated preview
type animatedPreviewRenderer struct {
	preview *AnimatedPreview
}

func (r *animatedPreviewRenderer) Layout(size fyne.Size) {
	r.preview.displayImage.Resize(size)
}

func (r *animatedPreviewRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 80)
}

func (r *animatedPreviewRenderer) Refresh() {
	r.preview.displayImage.Refresh()
}

func (r *animatedPreviewRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.preview.displayImage}
}

func (r *animatedPreviewRenderer) Destroy() {
	r.preview.stopAnimation()
}
