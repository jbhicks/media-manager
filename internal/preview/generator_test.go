package preview

import (
	"path/filepath"
	"testing"
)

type MockConfig struct{}

var _ ConfigProvider = (*MockConfig)(nil)

func (m *MockConfig) GetThumbnailDir() string {
	return "./mock_thumbnails"
}

func (m *MockConfig) GetThumbnailSize() int {
	return 300
}

func TestGeneratePreview(t *testing.T) {
	mockConfig := &MockConfig{}
	generator := NewPreviewGenerator(mockConfig)

	// Test image file
	imagePath := "test.jpg"
	thumbPath, err := generator.GeneratePreview(imagePath)
	if err != nil {
		t.Errorf("Failed to generate preview for image: %v", err)
	}
	if filepath.Base(thumbPath) != "test.jpg" {
		t.Errorf("Unexpected thumbnail path: %s", thumbPath)
	}

	// Test unsupported file
	unsupportedPath := "test.txt"
	_, err = generator.GeneratePreview(unsupportedPath)
	if err == nil {
		t.Errorf("Expected error for unsupported file type")
	}
}

func TestIsImageFile(t *testing.T) {
	generator := &PreviewGenerator{config: &MockConfig{}}
	if !generator.isImageFile(".jpg") {
		t.Errorf("Expected .jpg to be recognized as an image file")
	}
	if generator.isImageFile(".txt") {
		t.Errorf("Expected .txt not to be recognized as an image file")
	}
}
