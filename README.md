<p align="center">
  <img src="web/images/logo.svg" alt="Media Manager Logo" width="200">
</p>

# Media Manager

A native desktop media management application built with Go and Fyne for browsing, organizing, and tagging image and video files.

## Features

- **Native Desktop UI**: Built with Fyne for cross-platform compatibility (Windows, macOS, Linux)
- **Media File Support**: Images (JPEG, PNG, GIF, WebP, TIFF, BMP) and Videos (MP4, AVI, MOV, MKV, WebM)
- **Real-time File Scanning**: Automatic detection of new media files
- **Thumbnail Generation**: Automatic thumbnail creation for fast browsing
- **Tagging System**: Organize files with custom tags and colors
- **SQLite Database**: Local storage with no external dependencies
- **Background Service**: Automated torrent downloading and media management
- **Torrent Download Manager**: Automated torrent searching and downloading with multi-query aggregation

## Components

### Desktop Application
The main GUI application for browsing and managing media files.

### Background Service  
Automated service for torrent downloads and media processing. See [SERVICE.md](docs/SERVICE.md) for details.

## Installation

### Prerequisites
- Go 1.21+
- C/C++ compiler (for Fyne dependencies)

### Build
```bash
git clone <repository>
cd media-manager
go mod tidy

# Build GUI application
go build -o bin/media-manager cmd/media-manager/main.go

# Build background service
go build -o bin/media-manager-service cmd/media-manager-service/main.go

# Create Windows installer (includes both applications)
make installer
```

### Installers
- **Windows**: `make installer` creates `dist/media-manager_installer.exe` - a complete Windows installer with shortcuts and uninstaller
- **Binaries**: Individual executables for Windows, macOS, and Linux available in `dist/` directory
```bash
# GUI application (runs without console window)
./bin/media-manager

# Background service (console application for logging)
./bin/media-manager-service
```

### Application Types
- **GUI Application**: Built with `-H=windowsgui` flag - runs as a pure Windows GUI application with no console window
- **Background Service**: Built as console application - shows terminal for logging and debugging

### Development Flags

- `--dev-reset`  
  **Purpose:** Deletes the entire thumbnail cache and database on startup, forcing a full rescan and regeneration of all previews and metadata.  
  **Usage:**
  ```bash
  ./bin/media-manager --dev-reset
  ```
  **When to use:**
  - After changing preview/thumbnail generation logic
  - When you want to clear all cached previews and start fresh
  - For development and debugging to ensure all media is reprocessed
  
  This flag is intended for development and testing. It will remove `~/.media-manager/thumbnails/` and `~/.media-manager/media.db` before launching the app normally.

## Architecture

- **Frontend**: Fyne-based native desktop GUI
- **Backend**: Go with SQLite database (GORM ORM)
- **File Scanning**: Real-time monitoring with fsnotify
- **Thumbnails**: On-demand generation with disk caching
- **Storage**: Local SQLite database and thumbnail cache

## Project Structure

```
cmd/
├── media-manager/          # GUI application entry point
│   ├── main.go            # Fyne desktop app
│   └── main_test.go       # GUI tests
├── media-manager-service/  # Background service entry point
│   └── main.go            # HTTP API service
└── clear-previews/        # Utility for clearing previews
    └── main.go

internal/
├── app/                   # Main application logic
├── config/                # Configuration management
├── db/                    # Database layer (SQLite + GORM)
├── ffmpeg/                # Video processing utilities
├── jellyfin/              # Jellyfin integration
├── preview/               # Thumbnail and preview generation
├── scanner/               # File system scanning
├── service/               # Background service logic
│   ├── download_manager.go # Torrent download management
│   ├── http_server.go     # REST API endpoints
│   ├── service.go          # Main service orchestration
│   ├── suggestion_service.go # Media suggestions
│   └── tmdb_service.go     # TMDB API integration
├── tagger/                # File tagging system
├── torrent/               # Torrent search providers
└── ui/                    # GUI components and views

pkg/models/                # Shared data models
web/                       # Web UI assets (HTML, CSS, JS)
docs/                      # Documentation
reference/                 # Reference materials (Primer CSS)
scripts/                   # Build and utility scripts
```

## FFmpeg Handling

The application automatically downloads ffmpeg binaries on first run:
- Binaries are cached in `~/.media-manager/bin/`
- Downloads from GitHub releases (one-time, ~150MB total)
- Falls back to system PATH if download fails
- Supports: Windows, macOS (Intel & Apple Silicon), Linux

## Configuration

The application stores data in `~/.media-manager/`:
- `media.db` - SQLite database
- `thumbnails/` - Generated thumbnail cache (16:9 aspect ratio thumbnails)
- `bin/` - Downloaded ffmpeg binaries (cached on first run)
- `frames/` - Animated preview frames for video hover effects

Environment variables:
- `DB_PATH` - Custom database path
- `THUMBNAIL_DIR` - Custom thumbnail directory  
- `THUMBNAIL_SIZE` - Thumbnail dimensions (default: 200)

**Development Note:** Air automatically clears the thumbnail cache on rebuild to ensure uniform sizing after generation logic changes. Use `make clear-cache` to manually clear thumbnails.

## Torrent Download Manager

The application includes an automated torrent download manager with advanced search capabilities:

### Features
- **Multi-Query Search Aggregation**: For movies, executes 21 separate queries (empty + 20 genre keywords) to build larger datasets
- **Jackett Integration**: Search across multiple torrent indexers simultaneously
- **Smart Filtering**: Filter by seeders, size, resolution, upload age
- **InfoHash Deduplication**: Prevents duplicate results across queries
- **Configurable Rules**: Define custom download rules with specific criteria
- **Automatic Execution**: Rules can be scheduled to run automatically

### Multi-Query Search Strategy

When searching for movies without a specific query, the download manager automatically executes 21 searches:
1. Empty query (popular torrents)
2. 20 genre queries: action, adventure, comedy, drama, thriller, horror, sci-fi, fantasy, romance, crime, mystery, animation, documentary, family, superhero, war, western, musical, biography, sport

This strategy aggregates **~10x more results** than single-query searches, providing better coverage and higher-quality results.

### Example Usage

```bash
# Run the demo to see multi-query search in action
JACKETT_API_KEY=your-key go run examples/show_download_list.go
```

The demo will:
1. Execute all 21 genre queries
2. Aggregate and deduplicate results by InfoHash
3. Filter by your criteria (seeders, size, resolution, age)
4. Display the top 100 movies ready for download

### Configuration

Download rules support the following parameters:
- `SearchQuery`: Specific search term (empty = multi-query for movies)
- `MediaType`: "movie" or "tv"
- `Resolution`: "1080p", "720p", "4k", etc.
- `MinSeeders`: Minimum seeders required
- `MinSize`/`MaxSize`: Size range in bytes
- `MaxUploadAge`: Maximum age in days
- `MaxResults`: Maximum results to return
- `MaxResultsPerTitle`: Maximum results per unique title (deduplication)
- `SortBy`: "seeders", "size", "balanced"

## Current Status

✅ **Phase 1 Complete**: Core desktop application
- [x] Fyne application framework
- [x] Database models and SQLite integration
- [x] File scanner with real-time monitoring
- [x] Basic UI layout with sidebar and media grid
- [x] Preview/thumbnail generation system

✅ **Phase 2 Complete**: Torrent Download Manager
- [x] Jackett integration for torrent search
- [x] Multi-query search aggregation (21 genre queries for movies)
- [x] Smart filtering (seeders, size, resolution, age)
- [x] InfoHash-based deduplication
- [x] Download rules with automatic execution
- [x] Comprehensive test suite

✅ **Phase 3 Complete**: Media Presentation
- [x] TV show poster support with automatic detection (97% coverage)
- [x] Modern 16:9 card design with full-bleed images
- [x] Animated hover previews for videos
- [x] Scene-based preview generation with GPU acceleration
- [x] Jellyfin integration for media streaming

🚧 **Current Phase**: UI Polish and Features
- [ ] Search and filtering
- [ ] Tag management interface
- [ ] Settings dialog
- [ ] UI zoom (accessibility)

## Development

```bash
# Run GUI application (shows console during development)
go run cmd/media-manager/main.go

# Run background service (always shows console for logging)
go run cmd/media-manager-service/main.go

# Run tests
go test ./...

# Build for different platforms
GOOS=windows go build -ldflags="-H=windowsgui" -o bin/media-manager.exe cmd/media-manager/main.go
GOOS=windows go build -o bin/media-manager-service.exe cmd/media-manager-service/main.go
GOOS=darwin go build -o bin/media-manager-mac cmd/media-manager/main.go
GOOS=darwin go build -o bin/media-manager-service-mac cmd/media-manager-service/main.go
GOOS=linux go build -o bin/media-manager-linux cmd/media-manager/main.go
GOOS=linux go build -o bin/media-manager-service-linux cmd/media-manager-service/main.go
```

**Note**: The GUI application is built with `-H=windowsgui` to hide the console window in production. During development with `go run`, the console remains visible for debugging.
