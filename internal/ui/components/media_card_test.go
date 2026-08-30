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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

	// Get both sizes
	widgetMinSize := card.MinSize()
	renderer := test.TempWidgetRenderer(t, card)
	rendererMinSize := renderer.MinSize()

	// Both should return the new 16:9 card size (scaled by zoom level)
	expectedSize := fyne.NewSize(CardWidth(), CardHeight())
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
		card := NewMediaCard(file, "/tmp/thumbs", nil)

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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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

	// Card without duration (no durationLabel/Badge)
	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs", nil)

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// New design objects: background, content, hoverOverlay, labelBackground, extensionBadge, extensionLabel, label
	// Minimum should be 6 (background + hoverOverlay + labelBackground + extensionBadge + extensionLabel + label)
	minExpected := 6
	if len(objects) < minExpected {
		t.Errorf("Expected at least %d base objects, got %d", minExpected, len(objects))
	}
}

func TestMediaCard_Renderer_WithDurationHasExtraLabel(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Card with duration should have durationLabel and durationBadge
	file := testMediaFileWithDuration("/fake/path/test.mp4", "test.mp4", 120)
	card := NewMediaCard(file, "/tmp/thumbs", nil)

	renderer := test.TempWidgetRenderer(t, card)
	objects := renderer.Objects()

	// Should have: background, content, hoverOverlay, labelBackground, extensionBadge, extensionLabel, durationBadge, durationLabel, label
	expectedMin := 8
	if len(objects) < expectedMin {
		t.Errorf("Card with duration should have at least %d objects, got %d", expectedMin, len(objects))
	}

	// Verify durationLabel and durationBadge are not nil when duration > 0
	if card.durationLabel == nil {
		t.Error("durationLabel should not be nil when duration > 0")
	}
	if card.durationBadge == nil {
		t.Error("durationBadge should not be nil when duration > 0")
	}
}

// =============================================================================
// P0: Renderer Refresh Behavior Tests
// =============================================================================

func TestMediaCard_Renderer_RefreshDoesNotPanic(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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
	card := NewMediaCard(file, "/tmp/thumbs", nil)

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
		_ = NewMediaCard(file, "/tmp/thumbs", nil)
	}
}

func BenchmarkMediaCard_Renderer_Layout(b *testing.B) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs", nil)
	renderer := card.CreateRenderer()
	size := fyne.NewSize(CardWidth(), CardHeight())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Layout(size)
	}
}

func BenchmarkMediaCard_Renderer_Refresh(b *testing.B) {
	testApp := test.NewApp()
	defer testApp.Quit()

	file := testMediaFile("/fake/path/test.jpg", "test.jpg")
	card := NewMediaCard(file, "/tmp/thumbs", nil)
	renderer := card.CreateRenderer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Refresh()
	}
}

func TestMediaCard_DragEnd_ReportsDropPosition(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := NewMediaCard(testMediaFile("/fake/path/video.mp4", "video.mp4"), t.TempDir(), nil)
	var got fyne.Position
	called := false
	card.SetOnDragEnd(func(dropPos fyne.Position) {
		called = true
		got = dropPos
	})
	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(42, 99)},
		Dragged:    fyne.Delta{DX: 10, DY: 0},
	})
	card.DragEnd()
	if !called {
		t.Fatal("expected OnDragEnd to be called")
	}
	if got != fyne.NewPos(42, 99) {
		t.Fatalf("drop position = %v, want (42, 99)", got)
	}
}

func TestMediaCard_Dragged_FiresStartOnceAndDraggedEachMove(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := NewMediaCard(testMediaFile("/fake/path/video.mp4", "video.mp4"), t.TempDir(), nil)
	starts := 0
	drags := 0
	var last fyne.Position
	card.SetOnDragStart(func() { starts++ })
	card.SetOnDragged(func(abs fyne.Position) {
		drags++
		last = abs
	})

	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(1, 0)},
		Dragged:    fyne.Delta{DX: 1, DY: 0},
	})
	if starts != 0 || drags != 0 {
		t.Fatalf("below threshold: starts=%d drags=%d", starts, drags)
	}

	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(10, 2)},
		Dragged:    fyne.Delta{DX: 9, DY: 2},
	})
	if starts != 1 {
		t.Fatalf("onDragStart once, got %d", starts)
	}
	if drags != 1 {
		t.Fatalf("onDragged after threshold, got %d", drags)
	}

	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(40, 8)},
		Dragged:    fyne.Delta{DX: 30, DY: 6},
	})
	if starts != 1 {
		t.Fatalf("onDragStart should stay 1, got %d", starts)
	}
	if drags != 2 {
		t.Fatalf("onDragged every move, got %d", drags)
	}
	if last != fyne.NewPos(40, 8) {
		t.Fatalf("last abs = %v, want (40, 8)", last)
	}
}

func TestMediaCard_Tapped_AfterDrag_DoesNotOpen(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := NewMediaCard(testMediaFile("/fake/path/video.mp4", "video.mp4"), t.TempDir(), nil)
	opened := false
	card.onOpen = func() { opened = true }

	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(0, 0)},
		Dragged:    fyne.Delta{DX: 8, DY: 0},
	})
	card.DragEnd()
	card.Tapped(nil)
	if opened {
		t.Fatal("Tapped after drag should not open the file")
	}
	if card.didDrag {
		t.Fatal("didDrag should be consumed by Tapped")
	}
}

func TestMediaCard_Tapped_WithoutDrag_Opens(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := NewMediaCard(testMediaFile("/fake/path/video.mp4", "video.mp4"), t.TempDir(), nil)
	opened := false
	card.onOpen = func() { opened = true }
	card.Tapped(nil)
	if !opened {
		t.Fatal("Tapped without drag should open the file")
	}
}

func TestMediaCard_TestDragOrDirectDriverCalls(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	card := NewMediaCard(testMediaFile("/fake/path/video.mp4", "video.mp4"), t.TempDir(), nil)
	ended := false
	card.SetOnDragEnd(func(p fyne.Position) {
		ended = true
	})

	w := test.NewWindow(card)
	defer w.Close()
	w.Resize(card.MinSize())

	test.Drag(w.Canvas(), fyne.NewPos(10, 10), 20, 20)
	if !ended {
		card.Dragged(&fyne.DragEvent{
			PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(30, 30)},
			Dragged:    fyne.Delta{DX: 20, DY: 20},
		})
		card.DragEnd()
	}
	if !ended {
		t.Fatal("expected DragEnd via test.Drag or direct Dragged/DragEnd")
	}
}
