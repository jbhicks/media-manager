# Media Manager

A native desktop media management application built with Go and Fyne for browsing, organizing, and tagging image and video files.

## Features

- **Native Desktop UI**: Built with Fyne for cross-platform compatibility (Windows, macOS, Linux)
- **Media File Support**: Images (JPEG, PNG, GIF, WebP, TIFF, BMP) and Videos (MP4, AVI, MOV, MKV, WebM)
- **Real-time File Scanning**: Automatic detection of new media files
- **Thumbnail Generation**: Automatic thumbnail creation for fast browsing
- **Tagging System**: Organize files with custom tags and colors
- **SQLite Database**: Local storage with no external dependencies

## Installation

### Prerequisites
- Go 1.24+ 
- C/C++ compiler (for Fyne dependencies)

**Note:** FFmpeg is automatically downloaded on first run for video thumbnail generation. No manual installation required.

### Build
```bash
git clone <repository>
cd media-manager
go mod tidy
go build -o bin/media-manager cmd/media-manager/main.go
```

### Run
```bash
./bin/media-manager
```

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
├── cmd/media-manager/     # Application entry point
├── internal/
│   ├── app/              # Application setup and main window
│   ├── ui/               # Fyne UI components and views
│   ├── db/               # Database layer and models
│   ├── scanner/          # File system scanning
│   ├── preview/          # Thumbnail generation
│   ├── ffmpeg/           # FFmpeg binary management (auto-download)
│   └── config/           # Configuration management
├── pkg/
│   ├── models/           # Shared data structures
│   └── utils/            # Utility functions
└── bin/                  # Built executables
```

## FFmpeg Handling

The application automatically downloads ffmpeg binaries on first run:
- Binaries are cached in `~/.media-manager/bin/`
- Downloads from GitHub releases (one-time, ~150MB total)
- Falls back to system PATH if download fails
- Supports: macOS ARM64 (Apple Silicon)
- Coming soon: Linux, Windows, macOS Intel

## Configuration

The application stores data in `~/.media-manager/`:
- `media.db` - SQLite database
- `thumbnails/` - Generated thumbnail cache (uniform 200×200 images)
- `bin/` - Downloaded ffmpeg binaries (cached on first run)
- `previews/` - Generated video preview GIFs

Environment variables:
- `DB_PATH` - Custom database path
- `THUMBNAIL_DIR` - Custom thumbnail directory  
- `THUMBNAIL_SIZE` - Thumbnail dimensions (default: 200)

**Development Note:** Air automatically clears the thumbnail cache on rebuild to ensure uniform sizing after generation logic changes. Use `make clear-cache` to manually clear thumbnails.

## Current Status

✅ **Phase 1 Complete**: Core desktop application structure
- [x] Fyne application framework
- [x] Database models and SQLite integration
- [x] File scanner with real-time monitoring
- [x] Basic UI layout with sidebar and media grid
- [x] Preview/thumbnail generation system

🚧 **Next Phase**: Enhanced UI and functionality
- [ ] Actual media file loading and display
- [ ] Tag management interface
- [ ] Search and filtering
- [ ] Settings dialog

## Development

```bash
# Run application
go run cmd/media-manager/main.go

# Run tests
go test ./...

# Build for different platforms
GOOS=windows go build -o bin/media-manager.exe cmd/media-manager/main.go
GOOS=darwin go build -o bin/media-manager-mac cmd/media-manager/main.go
GOOS=linux go build -o bin/media-manager-linux cmd/media-manager/main.go
```
