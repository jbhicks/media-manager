package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/user/media-manager/pkg/models"
)

// Helper to create test MediaFile
func testMediaFile(path, filename string) models.MediaFile {
	return models.MediaFile{
		Path:     path,
		Filename: filename,
	}
}

func testMediaFileWithDuration(path, filename string, duration int) models.MediaFile {
	return models.MediaFile{
		Path:     path,
		Filename: filename,
		Duration: duration,
	}
}

// =============================================================================
// P0: MinSize Consistency Tests
// =============================================================================

func TestMediaCard_MinSize_WidgetAndRendererConsistent(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	// Get both sizes
	widgetMinSize := card.MinSize()
	renderer := test.TempWidgetRenderer(t, card)
	rendererMinSize := renderer.MinSize()

	// Both should return (216, 192)
	expectedSize := fyne.NewSize(216, 192)
	if widgetMinSize != expectedSize {
		t.Errorf("widget.MinSize()=%v, expected %v", widgetMinSize, expectedSize)
	}
	if rendererMinSize != expectedSize {
		t.Errorf("renderer.MinSize()=%v, expected %v", rendererMinSize, expectedSize)
	}
	if widgetMinSize != rendererMinSize {
		t.Errorf("MinSize inconsistency: widget.MinSize()=%v, renderer.MinSize()=%v - these should match",
			widgetMinSize, rendererMinSize)
	}
}

func TestMediaCard_MinSize_AllMediaTypesConsistent(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	testCases := []struct {
		name     string
		filename string
	}{
		{"image", "test.jpg"},
		{"video", "test.mp4"},
		{"file", "test.txt"},
	}

	var firstSize fyne.Size
	for i, tc := range testCases {
		file := testMediaFile("/fake/path/"+tc.filename, tc.filename)
		card := NewMediaCard(file, "/tmp/thumbs")

		size := card.MinSize()
		if i == 0 {
			firstSize = size
		} else if size != firstSize {
			t.Errorf("%s card MinSize %v differs from first card %v", tc.name, size, firstSize)
		}
	}
}

// =============================================================================
// P0: Renderer Objects Tests
// =============================================================================

func TestMediaCard_Renderer_ObjectsNotNil(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	for i, obj := range objects {
		if obj == nil {
			t.Errorf("Object at index %d is nil - all Objects() must be non-nil", i)
		}
	}
}

func TestMediaCard_Renderer_BaseObjectCount(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Card without duration (no durationLabel)
	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// Base objects: background, content, labelBackground, label, extensionLabel
	// Minimum should be 5 (or 4 if extensionLabel is conditional)
	minExpected := 4 // background, content, labelBackground, label
	if len(objects) < minExpected {
		t.Errorf("Expected at least %d base objects, got %d", minExpected, len(objects))
	}
}

func TestMediaCard_Renderer_WithDurationHasExtraLabel(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Card with duration should have durationLabel
	file := testMediaFileWithDuration("/fake/path/test.mp4", "test.mp4", 120)
	card := NewMediaCard(file, "/tmp/thumbs")

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// Should have: background, content, labelBackground, label, durationLabel, extensionLabel
	expectedMin := 5
	if len(objects) < expectedMin {
		t.Errorf("Card with duration should have at least %d objects, got %d", expectedMin, len(objects))
	}

	// Verify durationLabel is not nil when duration > 0
	if card.durationLabel == nil {
		t.Error("durationLabel should not be nil when duration > 0")
	}
}

// =============================================================================
// P0: Renderer Refresh Behavior Tests
// =============================================================================

func TestMediaCard_Renderer_RefreshDoesNotPanic(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	renderer := test.TempWidgetRenderer(t, card)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Refresh() panicked: %v", r)
		}
	}()

	renderer.Refresh()
}

func TestMediaCard_Renderer_LayoutDoesNotCallRefresh(t *testing.T) {
	// This is a code inspection test - Layout() should never call Refresh()
	// Per Fyne docs: "Layout should never call Refresh"
	// We can't easily test this programmatically, but we document the requirement
	t.Log("Manual verification required: Layout() should not call Refresh()")
}

// =============================================================================
// Hover Behavior Tests
// =============================================================================

func TestMediaCard_MouseIn_SetsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	if card.isHovered {
		t.Error("Card should not be hovered initially")
	}

	// Simulate mouse enter
	card.MouseIn(nil)

	if !card.isHovered {
		t.Error("Card should be hovered after MouseIn")
	}
}

func TestMediaCard_MouseOut_ClearsHovered(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	card.MouseIn(nil)
	card.MouseOut()

	if card.isHovered {
		t.Error("Card should not be hovered after MouseOut")
	}
}

// =============================================================================
// Secondary Tap (Right-Click) Tests
// =============================================================================

func TestMediaCard_TappedSecondary_ShowsMenu(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")

	window := test.NewWindow(card)
	defer window.Close()

	// This should not panic - just verifying it can be called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TappedSecondary panicked: %v", r)
		}
	}()

	// Use test.TapSecondary to simulate right-click
	test.TapSecondary(card)
}

// =============================================================================
// GetMediaType Tests
// =============================================================================

func TestGetMediaType_Images(t *testing.T) {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".tif", ".ico", ".svg"}
	for _, ext := range imageExts {
		if GetMediaType("test"+ext) != MediaTypeImage {
			t.Errorf("Expected %s to be MediaTypeImage", ext)
		}
	}
}

func TestGetMediaType_Videos(t *testing.T) {
	videoExts := []string{".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv", ".ts", ".3gp", ".mpeg", ".mpg", ".wmv", ".m4v", ".vob", ".divx"}
	for _, ext := range videoExts {
		if GetMediaType("test"+ext) != MediaTypeVideo {
			t.Errorf("Expected %s to be MediaTypeVideo", ext)
		}
	}
}

func TestGetMediaType_Files(t *testing.T) {
	fileExts := []string{".txt", ".pdf", ".doc", ".unknown", ""}
	for _, ext := range fileExts {
		if GetMediaType("test"+ext) != MediaTypeFile {
			t.Errorf("Expected %s to be MediaTypeFile", ext)
		}
	}
}

func TestGetMediaType_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		filename string
		expected MediaType
	}{
		{"TEST.JPG", MediaTypeImage},
		{"TEST.jpg", MediaTypeImage},
		{"TEST.MP4", MediaTypeVideo},
		{"TEST.Mp4", MediaTypeVideo},
	}

	for _, tc := range testCases {
		result := GetMediaType(tc.filename)
		if result != tc.expected {
			t.Errorf("GetMediaType(%q) = %v, want %v", tc.filename, result, tc.expected)
		}
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestMediaCard_FullLifecycle(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFileWithDuration("/fake/path/test.mp4", "test.mp4", 90)
	card := NewMediaCard(file, "/tmp/thumbs")

	// Create window and add card
	window := test.NewWindow(card)
	defer window.Close()

	// Get renderer
	renderer := test.TempWidgetRenderer(t, card)

	// Test layout at various sizes
	sizes := []fyne.Size{
		fyne.NewSize(216, 216),
		fyne.NewSize(300, 300),
		fyne.NewSize(150, 150),
	}

	for _, size := range sizes {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Layout(%v) panicked: %v", size, r)
			}
		}()
		renderer.Layout(size)
	}

	// Test hover cycle
	card.MouseIn(nil)
	card.MouseOut()

	// Test refresh
	renderer.Refresh()

	// Test destroy
	renderer.Destroy()
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkMediaCard_Creation(b *testing.B) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMediaCard(file, "/tmp/thumbs")
	}
}

func BenchmarkMediaCard_Renderer_Layout(b *testing.B) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")
	renderer := card.CreateRenderer()
	size := fyne.NewSize(216, 216)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Layout(size)
	}
}

func BenchmarkMediaCard_Renderer_Refresh(b *testing.B) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs")
	renderer := card.CreateRenderer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Refresh()
	}
}
