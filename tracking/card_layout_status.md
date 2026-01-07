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

## UI Zoom (Ctrl - / +) Implementation Plan

- [ ] Step 1: Create a custom theme type (e.g., ZoomableTheme) that wraps the current theme and multiplies all size values (text, icon, padding, etc.) by a zoom factor.
- [ ] Step 2: Add a global variable (e.g., zoomLevel) to track the current zoom factor (default 1.0).
- [ ] Step 3: Implement a function to update the zoom level and re-apply the custom theme using fyne.CurrentApp().Settings().SetTheme().
- [ ] Step 4: Add keyboard shortcut handling at the top-level window or main view to listen for Ctrl +/-, Cmd +/-, and update the zoom level accordingly.
- [ ] Step 5: Ensure that the zoom level is clamped to a reasonable range (e.g., 0.5x to 2.0x).
- [ ] Step 6: (Optional) Persist the zoom level in the app's config so it is restored on restart.
- [ ] Step 7: Test the feature on all platforms (Linux, Windows, Mac) to ensure that all UI elements (text, icons, padding, etc.) scale smoothly and shortcuts work as expected.
- [ ] Step 8: Add documentation/comments explaining the zoom feature and how to adjust it.

---

## Status

- **Last updated:** 2026-01-20
- **Owner:** Automated agent
- **Next review:** After visual testing complete
