# Changelog

## [Unreleased]
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

### Changed
- Improved error handling for video thumbnail generation
- Enhanced UI for folder addition and refresh functionality

### Fixed
- Resolved redundant `cmd.Run()` calls in video thumbnail generation
