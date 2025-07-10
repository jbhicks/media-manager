package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestMediaCardLayout(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	imageCard := NewMediaCard("/fake/path/test.jpg", "test.jpg", MediaTypeImage, "")

	// Test card minimum size
	expectedCardSize := fyne.NewSize(180, 101)
	if imageCard.MinSize() != expectedCardSize {
		t.Errorf("Expected card MinSize to be %v, got %v", expectedCardSize, imageCard.MinSize())
	}

	// Create renderer to test layout
	renderer := imageCard.CreateRenderer()
	cardSize := fyne.NewSize(180, 180)

	// Apply layout
	renderer.Layout(cardSize)

	// Get the objects from renderer
	objects := renderer.Objects()
	if len(objects) != 4 {
		t.Errorf("Expected 4 objects (background, content, labelBackground, label), got %d", len(objects))
	}
}

func TestMediaCardUniformSizing(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// Test that all card types have identical sizes
	imageCard := NewMediaCard("/fake/path/test.jpg", "test.jpg", MediaTypeImage, "")
	videoCard := NewMediaCard("/fake/path/test.mp4", "test.mp4", MediaTypeVideo, "")
	fileCard := NewMediaCard("/fake/path/test.txt", "test.txt", MediaTypeFile, "")

	cards := []*MediaCard{imageCard, videoCard, fileCard}
	expectedSize := fyne.NewSize(180, 101)

	for i, card := range cards {
		if card.MinSize() != expectedSize {
			t.Errorf("Card %d has different MinSize: expected %v, got %v", i, expectedSize, card.MinSize())
		}

		// Test renderer layout
		renderer := imageCard.CreateRenderer()
		renderer.Layout(fyne.NewSize(180, 180))

		objects := renderer.Objects()
		if len(objects) != 4 {
			t.Errorf("Card %d has wrong number of objects: expected 4, got %d", i, len(objects))
		}
	}
}
