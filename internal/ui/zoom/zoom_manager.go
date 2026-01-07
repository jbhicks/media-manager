// Package zoom provides UI scaling/zoom functionality for the media manager
package zoom

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"

	xtheme "fyne.io/x/fyne/theme"

	customtheme "github.com/user/media-manager/internal/ui/theme"
)

const (
	// MinZoom is the minimum zoom level (50%)
	MinZoom = 0.5
	// MaxZoom is the maximum zoom level (200%)
	MaxZoom = 2.0
	// DefaultZoom is the default zoom level (100%)
	DefaultZoom = 1.0
	// ZoomStep is the amount to change zoom level per step
	ZoomStep = 0.1
)

// Available theme names
const (
	ThemeDefault = "Default"
	ThemeDark    = "Dark"
	ThemeLight   = "Light"
	ThemeAdwaita = "Adwaita"
)

// Manager handles zoom level and theme management
type Manager struct {
	app            fyne.App
	zoomLevel      float32
	themeName      string
	zoomTheme      *customtheme.ZoomableTheme
	onZoomChanged  func(float32)
	onThemeChanged func(string)
}

// NewManager creates a new zoom manager
func NewManager(app fyne.App, zoomLevel float32, themeName string) *Manager {
	if zoomLevel < MinZoom || zoomLevel > MaxZoom {
		zoomLevel = DefaultZoom
	}
	if themeName == "" {
		themeName = ThemeDefault
	}

	m := &Manager{
		app:       app,
		zoomLevel: zoomLevel,
		themeName: themeName,
	}

	// Create the zoomable theme with the selected base theme
	baseTheme := m.getBaseTheme(themeName)
	m.zoomTheme = customtheme.NewZoomableTheme(baseTheme, zoomLevel)

	return m
}

// getBaseTheme returns the fyne.Theme for a given theme name
func (m *Manager) getBaseTheme(name string) fyne.Theme {
	switch name {
	case ThemeDark:
		return &forcedVariantTheme{base: theme.DefaultTheme(), variant: theme.VariantDark}
	case ThemeLight:
		return &forcedVariantTheme{base: theme.DefaultTheme(), variant: theme.VariantLight}
	case ThemeAdwaita:
		return xtheme.AdwaitaTheme()
	default:
		return theme.DefaultTheme()
	}
}

// Apply sets the zoomable theme on the application
func (m *Manager) Apply() {
	m.app.Settings().SetTheme(m.zoomTheme)
	fmt.Printf("[DEBUG] Applied theme: %s at zoom level: %.1f%%\n", m.themeName, m.zoomLevel*100)
}

// ZoomIn increases the zoom level by one step
func (m *Manager) ZoomIn() {
	newLevel := m.zoomLevel + ZoomStep
	if newLevel > MaxZoom {
		newLevel = MaxZoom
	}
	m.SetZoomLevel(newLevel)
}

// ZoomOut decreases the zoom level by one step
func (m *Manager) ZoomOut() {
	newLevel := m.zoomLevel - ZoomStep
	if newLevel < MinZoom {
		newLevel = MinZoom
	}
	m.SetZoomLevel(newLevel)
}

// ResetZoom resets the zoom level to default
func (m *Manager) ResetZoom() {
	m.SetZoomLevel(DefaultZoom)
}

// SetZoomLevel sets the zoom level directly
func (m *Manager) SetZoomLevel(level float32) {
	if level < MinZoom {
		level = MinZoom
	}
	if level > MaxZoom {
		level = MaxZoom
	}
	m.zoomLevel = level
	m.zoomTheme.SetZoomLevel(level)
	m.Apply()
	fmt.Printf("[DEBUG] Zoom level changed to: %.0f%%\n", level*100)

	if m.onZoomChanged != nil {
		m.onZoomChanged(level)
	}
}

// GetZoomLevel returns the current zoom level
func (m *Manager) GetZoomLevel() float32 {
	return m.zoomLevel
}

// SetTheme changes the base theme
func (m *Manager) SetTheme(name string) {
	m.themeName = name
	baseTheme := m.getBaseTheme(name)
	m.zoomTheme.SetBaseTheme(baseTheme)
	m.Apply()
	fmt.Printf("[DEBUG] Theme changed to: %s\n", name)

	if m.onThemeChanged != nil {
		m.onThemeChanged(name)
	}
}

// GetThemeName returns the current theme name
func (m *Manager) GetThemeName() string {
	return m.themeName
}

// GetAvailableThemes returns a list of available theme names
func (m *Manager) GetAvailableThemes() []string {
	return []string{ThemeDefault, ThemeDark, ThemeLight, ThemeAdwaita}
}

// SetOnZoomChanged sets a callback for when zoom level changes
func (m *Manager) SetOnZoomChanged(callback func(float32)) {
	m.onZoomChanged = callback
}

// SetOnThemeChanged sets a callback for when theme changes
func (m *Manager) SetOnThemeChanged(callback func(string)) {
	m.onThemeChanged = callback
}

// RegisterShortcuts registers keyboard shortcuts for zoom control
func (m *Manager) RegisterShortcuts(window fyne.Window) {
	canvas := window.Canvas()

	// Ctrl+= (Plus) - Zoom In
	ctrlPlus := &desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierControl}
	canvas.AddShortcut(ctrlPlus, func(shortcut fyne.Shortcut) {
		m.ZoomIn()
	})

	// Ctrl+- (Minus) - Zoom Out
	ctrlMinus := &desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierControl}
	canvas.AddShortcut(ctrlMinus, func(shortcut fyne.Shortcut) {
		m.ZoomOut()
	})

	// Ctrl+0 - Reset Zoom
	ctrlZero := &desktop.CustomShortcut{KeyName: fyne.Key0, Modifier: fyne.KeyModifierControl}
	canvas.AddShortcut(ctrlZero, func(shortcut fyne.Shortcut) {
		m.ResetZoom()
	})

	fmt.Println("[DEBUG] Registered zoom shortcuts: Ctrl+= (zoom in), Ctrl+- (zoom out), Ctrl+0 (reset)")
}

// forcedVariantTheme wraps a theme and forces a specific variant
type forcedVariantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (f *forcedVariantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.base.Color(name, f.variant)
}

func (f *forcedVariantTheme) Font(style fyne.TextStyle) fyne.Resource {
	return f.base.Font(style)
}

func (f *forcedVariantTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return f.base.Icon(name)
}

func (f *forcedVariantTheme) Size(name fyne.ThemeSizeName) float32 {
	return f.base.Size(name)
}

var _ fyne.Theme = (*forcedVariantTheme)(nil)
