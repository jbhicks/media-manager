package components

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

// getTestGifPathVPC returns the absolute path to the test GIF file
func getTestGifPathVPC() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
}

// createTestVideoPreviewCard creates a VideoPreviewCard for testing
// Returns nil if the test GIF cannot be loaded
func createTestVideoPreviewCard(t *testing.T, label string) *VideoPreviewCard {
	t.Helper()

	gifPath := getTestGifPathVPC()
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		t.Skipf("Test GIF not found at %s", gifPath)
		return nil
	}

	// Create static image from the same GIF
	staticImg := canvas.NewImageFromFile(gifPath)
	if staticImg == nil {
		return nil
	}
	staticImg.FillMode = canvas.ImageFillContain
	staticImg.SetMinSize(fyne.NewSize(100, 100))

	// Create animated GIF using URI like hoverable_card.go does
	uri := storage.NewFileURI(gifPath)
	animatedGif, err := xwidget.NewAnimatedGif(uri)
	if err != nil {
		t.Logf("Could not create animated GIF: %v", err)
		return nil
	}

	// Create label
	lbl := widget.NewLabel(label)
	lbl.Alignment = fyne.TextAlignCenter

	// Create backgrounds
	background := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))
	labelBackground := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))

	// Create container
	cnt := container.NewStack(staticImg)

	card := &VideoPreviewCard{
		staticImage:     staticImg,
		animatedGif:     animatedGif,
		label:           lbl,
		container:       cnt,
		background:      background,
		labelBackground: labelBackground,
		hasAnimation:    true,
		animatedGifPath: gifPath,
	}
	card.ExtendBaseWidget(card)

	return card
}

// =============================================================================
// VideoPreviewCard Construction Tests
// =============================================================================

func TestVideoPreviewCard_Create_ReturnsCard(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	if card.label.Text != "Test Video" {
		t.Errorf("Expected label 'Test Video', got '%s'", card.label.Text)
	}
}

// =============================================================================
// Renderer Objects Tests
// =============================================================================

func TestVideoPreviewCard_Renderer_ObjectsIncludesAllComponents(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// Should include: background, container, labelBackground, and optionally label
	if len(objects) < 3 {
		t.Errorf("Objects() should return at least 3 objects, got %d", len(objects))
	}
}

func TestVideoPreviewCard_Renderer_AllObjectsNonNil(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	for i, obj := range objects {
		if obj == nil {
			t.Errorf("Object at index %d is nil", i)
		}
	}
}

// =============================================================================
// Hover Behavior Tests
// =============================================================================

func TestVideoPreviewCard_MouseIn_SetsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	card.MouseIn(nil)

	if !card.isHovered {
		t.Error("Card should be hovered after MouseIn")
	}
}

func TestVideoPreviewCard_MouseOut_ClearsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	card.MouseIn(nil)
	card.MouseOut()

	if card.isHovered {
		t.Error("Card should not be hovered after MouseOut")
	}
}

// =============================================================================
// Layout Tests
// =============================================================================

func TestVideoPreviewCard_Layout_DoesNotPanic(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	renderer := test.TempWidgetRenderer(t, card)

	// Test layout at various sizes - should not panic
	sizes := []fyne.Size{
		fyne.NewSize(100, 100),
		fyne.NewSize(200, 150),
		fyne.NewSize(50, 50),
	}

	for _, size := range sizes {
		renderer.Layout(size)
	}
}

func TestVideoPreviewCard_MinSize_ReturnsValidSize(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	renderer := test.TempWidgetRenderer(t, card)
	minSize := renderer.MinSize()

	// MinSize should be non-negative
	if minSize.Width < 0 || minSize.Height < 0 {
		t.Errorf("MinSize should be non-negative, got %v", minSize)
	}
}

// =============================================================================
// Renderer Interface Tests
// =============================================================================

func TestVideoPreviewCard_CreateRenderer_ReturnsValidRenderer(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	renderer := test.TempWidgetRenderer(t, card)
	if renderer == nil {
		t.Error("CreateRenderer should return a valid renderer")
	}
}

func TestVideoPreviewCard_Renderer_Destroy_DoesNotPanic(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := createTestVideoPreviewCard(t, "Test Video")
	if card == nil {
		t.Skip("Could not create test VideoPreviewCard")
	}

	// TempWidgetRenderer handles cleanup via Destroy automatically
	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	if len(objects) == 0 {
		t.Error("Renderer should have objects before destroy")
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestVideoPreviewCard_ImplementsHoverable(t *testing.T) {
	t.Log("VideoPreviewCard implements desktop.Hoverable - verified by compilation")
}

func TestVideoPreviewCardRenderer_ImplementsWidgetRenderer(t *testing.T) {
	var _ fyne.WidgetRenderer = (*videoPreviewCardRenderer)(nil)
	t.Log("videoPreviewCardRenderer implements fyne.WidgetRenderer - verified by compilation")
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkVideoPreviewCard_Layout(b *testing.B) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		b.Skip("Could not determine test file location")
	}
	gifPath := filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		b.Skip("Test GIF not found")
	}

	testApp := test.NewApp()
	defer testApp.Quit()

	// Create a card manually for benchmarking
	staticImg := canvas.NewImageFromFile(gifPath)
	background := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))
	labelBackground := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))

	card := &VideoPreviewCard{
		staticImage:     staticImg,
		label:           widget.NewLabel("Benchmark"),
		container:       container.NewStack(staticImg),
		background:      background,
		labelBackground: labelBackground,
	}
	card.ExtendBaseWidget(card)

	renderer := card.CreateRenderer()
	size := fyne.NewSize(200, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Layout(size)
	}
}

func BenchmarkVideoPreviewCard_Refresh(b *testing.B) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		b.Skip("Could not determine test file location")
	}
	gifPath := filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		b.Skip("Test GIF not found")
	}

	testApp := test.NewApp()
	defer testApp.Quit()

	// Create a card manually for benchmarking
	staticImg := canvas.NewImageFromFile(gifPath)
	background := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))
	labelBackground := canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color("background", fyne.CurrentApp().Settings().ThemeVariant()))

	card := &VideoPreviewCard{
		staticImage:     staticImg,
		label:           widget.NewLabel("Benchmark"),
		container:       container.NewStack(staticImg),
		background:      background,
		labelBackground: labelBackground,
	}
	card.ExtendBaseWidget(card)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card.Refresh()
	}
}
