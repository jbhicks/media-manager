# Card Layout Refactor: Fyne Media Manager

## Summary

This file tracks the status of refactoring media cards to better fit content, especially for widescreen media, using idiomatic Fyne patterns.

---

## Current State (COMPLETED - January 2026)

- **MediaCard**: Now uses 16:9 aspect ratio (288x162), full-bleed `ImageFillCover`, rounded corners (8px), and overlay labels with gradient background.
- **Design**: Modern Netflix/YouTube style cards with image filling the entire card area.
- **Labels**: Filename displayed over semi-transparent gradient at bottom, extension badge top-right, duration badge bottom-right.
- **Hover**: Subtle white overlay on hover, animated GIF preview for video files.

---

## Idiomatic Fyne Card Patterns

- Use `canvas.Image.FillMode = ImageFillCover` for full-bleed images that fill the card.
- Use `canvas.Rectangle.CornerRadius` for rounded corners.
- Use `canvas.LinearGradient` for overlay gradients that ensure text legibility.
- Place labels as overlay elements on top of the image.

---

## Implementation Tasks (COMPLETED)

- [x] Refactor `MediaCard` to use `ImageFillCover` for full-bleed previews.
- [x] Change card dimensions to 288x162 (16:9 aspect ratio).
- [x] Add rounded corners (CornerRadius: 8) to card background.
- [x] Implement overlay gradient for filename label at bottom.
- [x] Add pill-shaped extension badge (top-right).
- [x] Add pill-shaped duration badge (bottom-right) for videos.
- [x] Implement hover overlay effect.
- [x] Update grid layout in main.go to use new dimensions.
- [x] Update tests with new expected sizes and object counts.
- [ ] Test with various aspect ratios (16:9, 4:3, portrait, square) to ensure content fits well.

---

## UI Zoom (Ctrl - / +) Implementation Plan (COMPLETED - January 2026)

- [x] Step 1: Create a custom theme type (ZoomableTheme) that wraps the current theme and multiplies all size values by a zoom factor.
  - Created `internal/ui/theme/zoomable_theme.go`
- [x] Step 2: Add zoom manager to track the current zoom factor (default 1.0).
  - Created `internal/ui/zoom/zoom_manager.go`
- [x] Step 3: Implement SetTheme() integration to update zoom level and re-apply theme.
- [x] Step 4: Add keyboard shortcut handling: Ctrl+= (zoom in), Ctrl+- (zoom out), Ctrl+0 (reset).
- [x] Step 5: Zoom level clamped to 0.5x - 2.0x range.
- [x] Step 6: Persist zoom level and theme name in config.json.
- [x] Step 7: Added theme selector (Default, Dark, Light, Adwaita from fyne-x).
- [x] Step 8: Added View > Theme menu and zoom controls.

---

## Status

- **Last updated:** 2026-01-20
- **Owner:** Automated agent
- **Next review:** After visual testing complete
