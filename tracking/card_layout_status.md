# Card Layout Refactor: Fyne Media Manager

## Summary

This file tracks the status of refactoring media cards to better fit content, especially for widescreen media, using idiomatic Fyne patterns.

---

## Current State

- **MediaCard**: Forces all media into a fixed square (180x180), using `ImageFillStretch` (distorts aspect ratio). Label is always at the bottom.
- **VideoPreviewCard**: Static image uses `ImageFillContain` (preserves aspect), but GIF preview is resized to 100x80 (may distort). Card is not responsive to media aspect ratio.
- **HoverableCard**: GIF fills 80% of card height, but no aspect ratio preservation.

---

## Idiomatic Fyne Card Patterns

- Use `container.NewVBox()` or `container.NewHBox()` to stack image/video and text.
- Use `canvas.Image.FillMode = ImageFillContain` to preserve aspect ratio.
- Allow card width to expand for widescreen content, or set a fixed height and let width be determined by the media’s aspect ratio.
- Add padding and a background rectangle (optionally rounded) for a card-like look.
- Place label below the media preview, with a gradient or solid background for readability.

---

## Recommendations

1. **Preserve Aspect Ratio**
   - Use `ImageFillContain` for all media previews (static and animated).
   - Avoid `ImageFillStretch` unless intentional distortion is desired.

2. **Responsive Card Sizing**
   - Allow card width to expand for widescreen content (fixed height, variable width).
   - Use `container.NewGridWrap()` or a custom layout to size the card based on the media’s natural aspect ratio.

3. **Flexible Layout**
   - Stack image/video and label using `container.NewVBox()` or `container.NewBorder()`.
   - Add padding and a rounded rectangle background for a more idiomatic card appearance.

4. **Consistent Label Placement**
   - Place the label below the media preview, with a readable background.

5. **Hover/Animation**
   - For video previews, swap static image for animated GIF on hover, but keep sizing/aspect ratio consistent.

---

## Implementation Tasks

- [ ] Refactor `MediaCard` and `VideoPreviewCard` to use `ImageFillContain` for all previews.
- [ ] Adjust card layout to allow width to expand for widescreen content (fixed height, variable width).
- [ ] Use `container.NewGridWrap()` or a custom layout to size cards based on media aspect ratio.
- [ ] Add padding and a rounded rectangle background for a more idiomatic card appearance.
- [ ] Ensure label is always below the media preview, with a readable background.
- [ ] Test with various aspect ratios (16:9, 4:3, portrait, square) to ensure content fits well.

---

## UI Zoom (Ctrl - / +) Implementation Plan

- [ ] Step 1: Create a custom theme type (e.g., ZoomableTheme) that wraps the current theme and multiplies all size values (text, icon, padding, etc.) by a zoom factor.
- [ ] Step 2: Add a global variable (e.g., zoomLevel) to track the current zoom factor (default 1.0).
- [ ] Step 3: Implement a function to update the zoom level and re-apply the custom theme using fyne.CurrentApp().Settings().SetTheme().
- [ ] Step 4: Add keyboard shortcut handling at the top-level window or main view to listen for Ctrl +/-, Cmd +/-, and update the zoom level accordingly.
- [ ] Step 5: Ensure that the zoom level is clamped to a reasonable range (e.g., 0.5x to 2.0x).
- [ ] Step 6: (Optional) Persist the zoom level in the app’s config so it is restored on restart.
- [ ] Step 7: Test the feature on all platforms (Linux, Windows, Mac) to ensure that all UI elements (text, icons, padding, etc.) scale smoothly and shortcuts work as expected.
- [ ] Step 8: Add documentation/comments explaining the zoom feature and how to adjust it.

---

## Status

- **Last updated:** 2025-07-15
- **Owner:** Automated agent
- **Next review:** After first implementation PR
