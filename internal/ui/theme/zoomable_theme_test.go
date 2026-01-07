package theme

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestZoomableTheme_Size(t *testing.T) {
	_ = test.NewApp()
	base := theme.DefaultTheme()

	tests := []struct {
		name      string
		zoomLevel float32
		sizeName  fyne.ThemeSizeName
	}{
		{"zoom 1.0x text", 1.0, theme.SizeNameText},
		{"zoom 1.5x text", 1.5, theme.SizeNameText},
		{"zoom 0.5x text", 0.5, theme.SizeNameText},
		{"zoom 2.0x padding", 2.0, theme.SizeNamePadding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zt := NewZoomableTheme(base, tt.zoomLevel)

			baseSize := base.Size(tt.sizeName)
			expectedSize := baseSize * tt.zoomLevel
			actualSize := zt.Size(tt.sizeName)

			if actualSize != expectedSize {
				t.Errorf("Size() = %v, want %v (base: %v, zoom: %v)",
					actualSize, expectedSize, baseSize, tt.zoomLevel)
			}
		})
	}
}

func TestZoomableTheme_SetZoomLevel(t *testing.T) {
	zt := NewZoomableTheme(nil, 1.0)

	zt.SetZoomLevel(1.5)
	if zt.GetZoomLevel() != 1.5 {
		t.Errorf("GetZoomLevel() = %v, want 1.5", zt.GetZoomLevel())
	}

	zt.SetZoomLevel(2.0)
	if zt.GetZoomLevel() != 2.0 {
		t.Errorf("GetZoomLevel() = %v, want 2.0", zt.GetZoomLevel())
	}
}

func TestZoomableTheme_DefaultValues(t *testing.T) {
	// Test nil base theme defaults to DefaultTheme
	zt := NewZoomableTheme(nil, 1.0)
	if zt.GetBaseTheme() == nil {
		t.Error("GetBaseTheme() should not be nil when created with nil")
	}

	// Test invalid zoom level defaults to 1.0
	zt2 := NewZoomableTheme(nil, 0)
	if zt2.GetZoomLevel() != 1.0 {
		t.Errorf("GetZoomLevel() = %v, want 1.0 for invalid input", zt2.GetZoomLevel())
	}

	zt3 := NewZoomableTheme(nil, -1.0)
	if zt3.GetZoomLevel() != 1.0 {
		t.Errorf("GetZoomLevel() = %v, want 1.0 for negative input", zt3.GetZoomLevel())
	}
}

func TestZoomableTheme_Color(t *testing.T) {
	_ = test.NewApp()
	base := theme.DefaultTheme()
	zt := NewZoomableTheme(base, 1.5)

	// Colors should be passed through unchanged (zoom doesn't affect colors)
	baseColor := base.Color(theme.ColorNamePrimary, theme.VariantDark)
	zoomColor := zt.Color(theme.ColorNamePrimary, theme.VariantDark)

	if baseColor != zoomColor {
		t.Errorf("Color() should pass through unchanged")
	}
}

func TestZoomableTheme_Font(t *testing.T) {
	_ = test.NewApp()
	base := theme.DefaultTheme()
	zt := NewZoomableTheme(base, 1.5)

	// Fonts should be passed through unchanged
	baseFont := base.Font(fyne.TextStyle{})
	zoomFont := zt.Font(fyne.TextStyle{})

	if baseFont != zoomFont {
		t.Errorf("Font() should pass through unchanged")
	}
}

func TestZoomableTheme_Icon(t *testing.T) {
	_ = test.NewApp()
	base := theme.DefaultTheme()
	zt := NewZoomableTheme(base, 1.5)

	// Icons should be passed through unchanged
	baseIcon := base.Icon(theme.IconNameHome)
	zoomIcon := zt.Icon(theme.IconNameHome)

	if baseIcon != zoomIcon {
		t.Errorf("Icon() should pass through unchanged")
	}
}
