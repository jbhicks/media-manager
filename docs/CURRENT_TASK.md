# Current Task: Content Area Scrolling Fix

**Status**: 🚧 In Progress  
**Last Updated**: 2026-01-06

## Problem Statement

When opening a folder with a large number of videos/media files, the window expands vertically past the available desktop space (below the taskbar). There is no scroll bar for the content area, making it impossible to see all content.

## Root Cause Analysis

In [main.go](internal/ui/views/main.go), the media grid is created using:
```go
v.mediaGridContainer = container.NewGridWrap(fyne.NewSize(cardWidth, cardHeight), cards...)
v.mediaGridWrapper = container.NewMax(v.mediaGridContainer)
```

The issue is that `container.NewMax()` does NOT constrain size - it expands to fit all content. The media grid wrapper is then placed in an `HSplit` without any scroll container wrapping the right side.

## Solution: Wrap Content in Scroll Container

According to Fyne documentation, the proper pattern is to use `container.NewVScroll()` or `container.NewScroll()` to wrap content that may exceed the available space. Key points from the docs:

1. **`container.NewScroll(content)`** - Creates a scrollable container in both directions
2. **`container.NewVScroll(content)`** - Creates a vertical-only scrollable container (ideal for grids)
3. **MinSize behavior**: Scroll containers can have a smaller MinSize than their content, enabling scrolling
4. **`scroll.SetMinSize()`** - Can set minimum visible size before scrolling kicks in

### Implementation Plan

1. **Wrap the media grid in a vertical scroll container**:
   ```go
   // Before (current - causes window to expand indefinitely):
   v.mediaGridWrapper = container.NewMax(v.mediaGridContainer)
   
   // After (fixed - enables scrolling):
   v.mediaGridWrapper = container.NewVScroll(v.mediaGridContainer)
   ```

2. **Files to modify**:
   - `internal/ui/views/main.go` - Wrap media grid in scroll container

3. **Considerations**:
   - The folder tree on the left already uses `container.NewScroll()` - this is correct
   - The media grid on the right needs the same treatment
   - `NewVScroll` is preferred over `NewScroll` since we only need vertical scrolling (horizontal is handled by GridWrap flowing content)

## Implementation Steps

- [x] In `createMediaGrid()`: Change `container.NewMax()` to `container.NewVScroll()`
- [x] In `RefreshMediaGridWithForce()`: Ensure wrapper remains a scroll container when refreshing
- [x] Update `mediaGridWrapper` type from `*fyne.Container` to `*container.Scroll`
- [x] Update `createMediaGrid()` return type to `fyne.CanvasObject`
- [ ] Test with large folder (100+ files) to verify scroll bar appears
- [ ] Test that window no longer expands past screen bounds
- [ ] Verify scroll bar styling matches theme

## Testing Checklist

- [ ] Window stays within screen bounds with many files
- [ ] Vertical scroll bar appears when content exceeds viewport
- [ ] Scroll bar works with mouse wheel
- [ ] Scroll bar works with drag
- [ ] GridWrap still wraps cards correctly at different window widths
- [ ] Zoom in/out still works with scroll container
- [ ] Filter still works (reduces/expands content)

## Fyne Scroll Container Reference

From Fyne v2.x documentation:

```go
// General scroll (both directions)
container.NewScroll(content fyne.CanvasObject) *Scroll

// Vertical only (what we need for media grid)
container.NewVScroll(content fyne.CanvasObject) *Scroll

// Horizontal only  
container.NewHScroll(content fyne.CanvasObject) *Scroll

// Set minimum size (optional - useful if grid should show at least N rows)
scroll.SetMinSize(fyne.NewSize(width, height))
```

**ScrollDirection constants** (if needed for finer control):
- `container.ScrollBoth` - horizontal and vertical
- `container.ScrollVerticalOnly` - top to bottom only
- `container.ScrollHorizontalOnly` - left to right only
- `container.ScrollNone` - disable scrolling

---

## Previous Completions (for reference)

### Media Card Redesign (January 2026) ✅
- 16:9 aspect ratio (288x162), full-bleed images, overlay labels, badges

### Video Preview System ✅
- Scene-based preview generation, GPU acceleration, animated hover previews

### UI Zoom & Theme Feature ✅
- Ctrl+=/Ctrl+- zoom, theme selector, settings persistence

## Quick Commands

```bash
# Build and run
go build -o bin/media-manager.exe ./cmd/media-manager
.\bin\media-manager.exe .\media

# Run tests
go test ./internal/ui/views/... -v

# Full test suite
go test ./...
```
