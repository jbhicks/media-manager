package components

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestMediaCardLayout(t *testing.T) {
	// Create a test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Create a test image file to avoid file system dependencies
	testImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			testImage.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red test image
		}
	}

	// Create media cards of different types
	imageCard := NewMediaCard("/fake/path/test.jpg", "test.jpg", MediaTypeImage)
	_ = NewMediaCard("/fake/path/test.txt", "test.txt", MediaTypeFile) // For completeness

	// Test card minimum size
	expectedCardSize := fyne.NewSize(90, 90)
	if imageCard.MinSize() != expectedCardSize {
		t.Errorf("Expected card MinSize to be %v, got %v", expectedCardSize, imageCard.MinSize())
	}

	// Create renderer to test layout
	renderer := imageCard.CreateRenderer()
	cardSize := fyne.NewSize(90, 90)

	// Apply layout
	renderer.Layout(cardSize)

	// Get the objects from renderer
	objects := renderer.Objects()
	if len(objects) != 3 {
		t.Errorf("Expected 3 objects (background, content, label), got %d", len(objects))
	}

	// Test background sizing - should fill entire card
	background := objects[0] // First object should be background
	if background.Size() != cardSize {
		t.Errorf("Expected background to fill entire card %v, got %v", cardSize, background.Size())
	}

	// Test content sizing - should take most of the space with minimal padding
	content := objects[1] // Second object should be content
	contentSize := content.Size()

	// Content should be almost the full width (minus padding)
	expectedContentWidth := float32(88) // 90-2 padding
	if contentSize.Width != expectedContentWidth {
		t.Errorf("Expected content width to be %f, got %f", expectedContentWidth, contentSize.Width)
	}

	// Content should take most of the height
	expectedContentHeight := float32(77) // 90 total - 12 label - 1 padding
	if contentSize.Height != expectedContentHeight {
		t.Errorf("Expected content height to be %f, got %f", expectedContentHeight, contentSize.Height)
	}

	// Test label sizing - should be at bottom with fixed height
	label := objects[2] // Third object should be label
	labelSize := label.Size()
	labelPos := label.Position()

	expectedLabelHeight := float32(12)
	if labelSize.Height != expectedLabelHeight {
		t.Errorf("Expected label height to be %f, got %f", expectedLabelHeight, labelSize.Height)
	}

	expectedLabelY := float32(78) // 90 - 12
	if labelPos.Y != expectedLabelY {
		t.Errorf("Expected label Y position to be %f, got %f", expectedLabelY, labelPos.Y)
	}

	// Test that content takes up most of the vertical space (should be at least 85% of card)
	contentHeightRatio := contentSize.Height / float32(90)
	minExpectedRatio := float32(0.85) // Content should take at least 85% of card height for tight layout
	if contentHeightRatio < minExpectedRatio {
		t.Errorf("Content height ratio is too small: %f (expected at least %f). This indicates excessive vertical spacing!",
			contentHeightRatio, minExpectedRatio)
	}

	// Test that label height is minimal
	if labelSize.Height > 12 {
		t.Errorf("Label height is too large: %f (expected 12). Label is taking too much space!",
			labelSize.Height)
	}
}

func TestMediaCardUniformSizing(t *testing.T) {
	// Test that all card types have identical sizes
	imageCard := NewMediaCard("/fake/path/test.jpg", "test.jpg", MediaTypeImage)
	videoCard := NewMediaCard("/fake/path/test.mp4", "test.mp4", MediaTypeVideo)
	fileCard := NewMediaCard("/fake/path/test.txt", "test.txt", MediaTypeFile)

	cards := []*MediaCard{imageCard, videoCard, fileCard}
	expectedSize := fyne.NewSize(90, 90)

	for i, card := range cards {
		if card.MinSize() != expectedSize {
			t.Errorf("Card %d has different MinSize: expected %v, got %v", i, expectedSize, card.MinSize())
		}

		// Test renderer layout
		renderer := card.CreateRenderer()
		renderer.Layout(expectedSize)

		objects := renderer.Objects()
		if len(objects) != 3 {
			t.Errorf("Card %d has wrong number of objects: expected 3, got %d", i, len(objects))
		}

		// All content objects should have the same size
		content := objects[1]
		expectedContentSize := fyne.NewSize(88, 77) // 90-2 width, 90-12-1 height
		if content.Size() != expectedContentSize {
			t.Errorf("Card %d content has wrong size: expected %v, got %v", i, expectedContentSize, content.Size())
		}
	}
}
