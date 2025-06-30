# Media Manager Project Log

## Feature Implementation Status (as of 2025-06-29)

### Implemented
- Native Desktop UI: Fyne-based UI is present (internal/ui/views/main.go)
- SQLite Database: GORM with SQLite is implemented (internal/db/database.go)
- Tagging System: Tag model and DB methods exist (pkg/models/models.go, internal/db/database.go)
- Real-time File Scanning: fsnotify-based scanner implemented (internal/scanner/scanner.go)
- Thumbnail Generation: Preview generator for images/videos (internal/preview/generator.go)
- Configuration: Config struct and env overrides (internal/config/config.go)

### Partially Implemented / TODO
- Media File Support: Image/video extensions handled, but UI/UX for all formats may need review
- Tagging UI: Tag display present, but add/edit features may be incomplete (see TODOs in main.go)
- Folder Addition: UI button exists, but not implemented (main.go)
- Refresh Button: UI exists, not implemented (main.go)
- Thumbnail Caching: Code for thumbnail pathing exists, but cache management may need review

### Missing/Unknown
- Video thumbnail generation via ffmpeg: ffmpeg command present, but integration and error handling need review
- Full-featured search/filter: Search bar in UI, but backend logic not implemented
- Export/Import: No evidence of export/import features
- User documentation: No user-facing docs beyond README

---

Update this log as features are completed or new ones are added.
