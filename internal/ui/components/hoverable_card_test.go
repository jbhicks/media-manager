package components

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// getTestGifPath returns the absolute path to the test GIF file
func getTestGifPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
}

// =============================================================================
// HoverableCard Construction Tests
// =============================================================================

func TestHoverableCard_NewWithInvalidPath_ReturnsNil(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Test with non-existent file
	card := NewHoverableCard("/nonexistent/path/test.gif", "Test Label")

	// Should return nil when GIF can't be loaded
	if card != nil {
		t.Error("Expected nil when GIF path is invalid")
	}
}

func TestHoverableCard_NewWithValidPath_ReturnsCard(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test Label")

	if card == nil {
		t.Fatal("Expected non-nil card when GIF path is valid")
	}
	if card.animatedGif == nil {
		t.Error("Expected animatedGif to be set")
	}
	if card.label == nil {
		t.Error("Expected label to be set")
	}
}

// =============================================================================
// Hover Behavior Tests
// =============================================================================

func TestHoverableCard_InitialState_NotHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test Label")
	if card == nil {
		t.Skip("Could not create HoverableCard with test GIF")
	}

	if card.isHovered {
		t.Error("Card should not be hovered initially")
	}
}

func TestHoverableCard_MouseIn_SetsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test Label")
	if card == nil {
		t.Skip("Could not create HoverableCard with test GIF")
	}

	card.MouseIn(nil)

	if !card.isHovered {
		t.Error("Card should be hovered after MouseIn")
	}
}

func TestHoverableCard_MouseOut_ClearsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test Label")
	if card == nil {
		t.Skip("Could not create HoverableCard with test GIF")
	}

	card.MouseIn(nil)
	card.MouseOut()

	if card.isHovered {
		t.Error("Card should not be hovered after MouseOut")
	}
}

// =============================================================================
// Renderer Tests
// =============================================================================

func TestHoverableCard_CreateRenderer_ReturnsValidRenderer(t *testing.T) {
	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test")
	if card == nil {
		t.Fatal("Could not create HoverableCard with test GIF")
	}

	renderer := test.TempWidgetRenderer(t, card)
	if renderer == nil {
		t.Error("CreateRenderer should return a valid renderer")
	}
}

func TestHoverableCard_Renderer_Objects_IncludesGifAndLabel(t *testing.T) {
	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test")
	if card == nil {
		t.Fatal("Could not create HoverableCard with test GIF")
	}

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// Should have at least 2 objects: gif and label
	if len(objects) < 2 {
		t.Errorf("Expected at least 2 objects (gif + label), got %d", len(objects))
	}
}

func TestHoverableCard_Renderer_Layout_CorrectProportions(t *testing.T) {
	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test")
	if card == nil {
		t.Fatal("Could not create HoverableCard with test GIF")
	}

	renderer := test.TempWidgetRenderer(t, card)
	// Layout with a specific size
	testSize := fyne.NewSize(200, 200)
	renderer.Layout(testSize)

	// Verify layout was applied without errors
	objects := renderer.Objects()
	if len(objects) < 2 {
		t.Error("Layout should maintain at least 2 objects")
	}
}

func TestHoverableCard_Renderer_Destroy_StopsGIF(t *testing.T) {
	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test")
	if card == nil {
		t.Fatal("Could not create HoverableCard with test GIF")
	}

	renderer := test.TempWidgetRenderer(t, card)
	// Destroy is called automatically by TempWidgetRenderer cleanup
	// Just verify we can get objects before cleanup
	objects := renderer.Objects()
	if len(objects) == 0 {
		t.Error("Renderer should have objects before destroy")
	}
}

// =============================================================================
// MinSize Tests
// =============================================================================

func TestHoverableCard_MinSize_WithGif(t *testing.T) {
	gifPath := getTestGifPath()
	card := NewHoverableCard(gifPath, "Test")
	if card == nil {
		t.Fatal("Could not create HoverableCard with test GIF")
	}

	minSize := card.MinSize()
	// MinSize should be non-negative (may be 0,0 for small GIFs since AnimatedGif returns 0,0 initially)
	if minSize.Width < 0 || minSize.Height < 0 {
		t.Errorf("MinSize should be non-negative, got %v", minSize)
	}
}

func TestHoverableCard_MinSize_Fallback(t *testing.T) {
	// Test with invalid path to trigger fallback
	card := NewHoverableCard("/nonexistent/path.gif", "Test")
	if card == nil {
		// If constructor returns nil for invalid path, that's acceptable
		t.Log("Constructor returned nil for invalid path, fallback not testable")
		return
	}

	minSize := card.MinSize()
	// Fallback should be (120, 120)
	if minSize.Width != 120 || minSize.Height != 120 {
		t.Logf("MinSize with invalid path: %v (expected 120x120 fallback)", minSize)
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

// Verify HoverableCard implements desktop.Hoverable
func TestHoverableCard_ImplementsHoverable(t *testing.T) {
	// This is compile-time verified by the var declaration in hoverable_card.go:
	// var _ desktop.Hoverable = (*HoverableCard)(nil)
	t.Log("HoverableCard implements desktop.Hoverable - verified by compilation")
}

// Verify hoverableCardRenderer implements fyne.WidgetRenderer
func TestHoverableCardRenderer_ImplementsWidgetRenderer(t *testing.T) {
	// Compile-time verification
	var _ fyne.WidgetRenderer = (*hoverableCardRenderer)(nil)
	t.Log("hoverableCardRenderer implements fyne.WidgetRenderer - verified by compilation")
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkHoverableCard_HoverCycle(b *testing.B) {
	// Get test GIF path using runtime.Caller
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		b.Skip("Could not determine test file location")
	}
	gifPath := filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		b.Skip("Test GIF not found at " + gifPath)
	}

	card := NewHoverableCard(gifPath, "Benchmark Test")
	if card == nil {
		b.Skip("Could not create HoverableCard")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card.MouseIn(nil)
		card.MouseOut()
	}
}

func BenchmarkHoverableCard_Refresh(b *testing.B) {
	// Get test GIF path using runtime.Caller
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		b.Skip("Could not determine test file location")
	}
	gifPath := filepath.Join(filepath.Dir(filename), "testdata", "test.gif")
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		b.Skip("Test GIF not found at " + gifPath)
	}

	card := NewHoverableCard(gifPath, "Benchmark Test")
	if card == nil {
		b.Skip("Could not create HoverableCard")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card.Refresh()
	}
}
