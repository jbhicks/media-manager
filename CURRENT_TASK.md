# Current Task: Fix video previews

**Status**: Completed
**Last Updated**: 2025-07-08

## Details
Investigated and fixed the issue where video file previews were not being generated.

## Sub-tasks / Progress
- [x] Verified current state of media loading and preview generation.
- [x] Identified root cause: `ensureVideoThumbnail` was not being called, and `media_card.go` was attempting complex animated GIF generation.
- [x] Removed animated GIF generation logic from `internal/ui/components/media_card.go`.
- [x] Modified `internal/ui/views/main.go` to call `ensureVideoThumbnail` for video files.
- [x] Updated `internal/ui/components/media_card.go` to display static video thumbnails.
- [x] Updated `internal/ui/components/media_card_test.go` to reflect `NewMediaCard` signature changes and adjusted layout expectations.
- [x] Ensured project builds successfully.
- [x] Ensured all tests pass.

## Notes
The primary issue was a disconnect between the static thumbnail generation logic in `main.go` and the animated GIF attempt in `media_card.go`. Simplified to static thumbnails for initial fix.

## Next Steps
Ready for user to verify the fix by running the application. Further enhancements (e.g., animated previews) can be considered as a separate task.