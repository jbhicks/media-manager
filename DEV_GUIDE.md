# Development Guide for Media Manager

## Quick Start
```bash
# Initialize and run
go mod tidy
go build -o bin/media-manager.exe ./cmd/media-manager
.\bin\media-manager.exe [media-directory]

# Development commands
go test ./...                           # Run all tests
go test ./internal/ui/components/...    # Test UI components
go test ./internal/preview/...          # Test preview generation
go build ./...                          # Build all packages
```

## Project Structure
```
├── cmd/
│   ├── media-manager/main.go          # Desktop GUI entry point
│   └── clear-previews/main.go         # Utility to clear preview cache
├── internal/                          # Private application code
│   ├── app/                           # Application setup
│   ├── ui/                           # Fyne UI components
│   │   ├── components/               # Reusable components (MediaCard, etc.)
│   │   ├── dialogs/                  # Modal dialogs
│   │   └── views/                    # Main views
│   ├── db/                           # Database layer (GORM)
│   ├── scanner/                      # File scanning
│   ├── preview/                      # Thumbnail/preview generation
│   ├── ffmpeg/                       # FFmpeg binary management
│   ├── service/                      # Download manager, TMDb, HTTP server
│   ├── torrent/                      # Torrent search providers
│   ├── jellyfin/                     # Jellyfin API client
│   └── config/                       # Configuration
├── pkg/                              # Public packages
│   ├── models/                       # Data structures
│   └── utils/                        # Utilities
├── web/                              # Web UI assets
├── docs/                             # Additional documentation
├── tracking/                         # Feature tracking docs
└── examples/                         # Demo scripts
```

## Key Technologies
- **Framework**: Fyne v2 (native desktop GUI)
- **Database**: SQLite with GORM ORM
- **Thumbnails**: FFmpeg for video, Go image packages for images
- **File Watching**: fsnotify for real-time updates
- **Testing**: Go standard testing, fyne.io/fyne/v2/test for GUI tests
- **Web UI**: HTMX + custom CSS (Primer CSS as reference)

## Configuration

Data stored in `~/.media-manager/`:
- `media.db` - SQLite database
- `thumbnails/` - Generated thumbnail cache
- `bin/` - Downloaded ffmpeg binaries
- `frames/` - Animated preview frames

Environment variables:
- `MEDIA_DIRS`: Directories to scan for media files
- `DB_PATH`: SQLite database file path
- `THUMBNAIL_DIR`: Thumbnail cache directory
- `TMDB_API_KEY`: TMDb API key for poster fetching
- `JACKETT_API_KEY`: Jackett API key for torrent search
- `JACKETT_URL`: Jackett server URL (default: http://localhost:9117)

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/ui/components/... -v

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./internal/ui/components/...
```

## Build

```bash
# Build for current platform
go build -o bin/media-manager.exe ./cmd/media-manager

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o bin/media-manager-linux ./cmd/media-manager
GOOS=darwin GOARCH=arm64 go build -o bin/media-manager-mac ./cmd/media-manager
```

## Development Flags

- `--dev-reset`: Deletes thumbnail cache and database on startup for fresh testing
