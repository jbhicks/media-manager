// Package theme provides custom theme implementations for the media manager
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ZoomableTheme wraps a base theme and multiplies all size values by a zoom factor
type ZoomableTheme struct {
	base      fyne.Theme
	zoomLevel float32
}

// NewZoomableTheme creates a new ZoomableTheme wrapping the given base theme
func NewZoomableTheme(base fyne.Theme, zoomLevel float32) *ZoomableTheme {
	if base == nil {
		base = theme.DefaultTheme()
	}
	if zoomLevel <= 0 {
		zoomLevel = 1.0
	}
	return &ZoomableTheme{
		base:      base,
		zoomLevel: zoomLevel,
	}
}

// SetZoomLevel updates the zoom level
func (z *ZoomableTheme) SetZoomLevel(level float32) {
	z.zoomLevel = level
}

// GetZoomLevel returns the current zoom level
func (z *ZoomableTheme) GetZoomLevel() float32 {
	return z.zoomLevel
}

// SetBaseTheme updates the underlying base theme
func (z *ZoomableTheme) SetBaseTheme(base fyne.Theme) {
	if base == nil {
		base = theme.DefaultTheme()
	}
	z.base = base
}

// GetBaseTheme returns the underlying base theme
func (z *ZoomableTheme) GetBaseTheme() fyne.Theme {
	return z.base
}

// Color returns the named color for the current theme
func (z *ZoomableTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return z.base.Color(name, variant)
}

// Font returns the named font for the current theme
func (z *ZoomableTheme) Font(style fyne.TextStyle) fyne.Resource {
	return z.base.Font(style)
}

// Icon returns the named icon for the current theme
func (z *ZoomableTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return z.base.Icon(name)
}

// Size returns the named size, scaled by the zoom level
func (z *ZoomableTheme) Size(name fyne.ThemeSizeName) float32 {
	return z.base.Size(name) * z.zoomLevel
}

// Ensure ZoomableTheme implements fyne.Theme
var _ fyne.Theme = (*ZoomableTheme)(nil)
