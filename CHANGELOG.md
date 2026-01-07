# Changelog

## [Unreleased]

## [0.3.0] - 2026-01-06
### Added
- **Modern Media Card Design**: 16:9 aspect ratio cards (288x162) with full-bleed images
- Overlay labels with gradient background for better legibility
- Pill-shaped badges for file extension and video duration
- Hover effects with subtle white overlay
- Rounded corners (8px) on all cards
- Exported CardWidth/CardHeight constants for grid layouts

### Changed
- Card dimensions changed from 216x192 to 288x162 (true 16:9 widescreen)
- Image fill mode changed to ImageFillStretch for full-bleed display
- Labels now overlay images instead of appearing below

### Fixed
- Updated AGENTS.md with modern Fyne card design patterns
- Cleaned up stray files and updated .gitignore

## [0.2.0] - 2025-12-25
### Added
- **TV Show Poster Support**: Automatic TV vs Movie detection for poster fetching
- TMDb TV API integration with retry logic for year-in-title cases
- `ExtractMediaInfo()` function for parsing TV show patterns (S01E01, 1x01, etc.)
- Poster coverage increased from 36% to 97% (86/89 suggestions)

### Changed
- Simplified `handleRefreshSearchPosters()` handler using unified `FetchPosterForTask()`
- Both movies and TV shows now cached in `movie_metadata` table

## [0.1.0] - 2025-07-01
### Added
- Initial implementation of thumbnail generation
- Basic tagging system with SQLite database
- Real-time file scanning using fsnotify
- **Torrent Download Manager with multi-query search aggregation**
  - Multi-query search strategy: executes 21 separate queries for movies (empty + 20 genre keywords)
  - Jackett integration for multi-indexer torrent searching
  - InfoHash-based deduplication across query results
  - Smart filtering by seeders, size, resolution, and upload age
  - Configurable download rules with automatic execution
  - Comprehensive test suite (unit + integration tests)
  - Demo script (`examples/show_download_list.go`) showcasing multi-query aggregation
  - Achieves ~10x more results than single-query searches
- **Video Preview System**
  - Scene-based preview generation using FFmpeg scene detection
  - Two-pass GIF encoding with optimal color palettes
  - Static mosaic preview option (2x2 thumbnail grid)
  - GPU acceleration support (CUDA, VAAPI)
  - Animated frame-based hover previews

### Changed
- Improved error handling for video thumbnail generation
- Enhanced UI for folder addition and refresh functionality

### Fixed
- Resolved redundant `cmd.Run()` calls in video thumbnail generation
