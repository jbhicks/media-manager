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
- FFmpeg (for video thumbnail generation)

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
│   └── config/           # Configuration management
├── pkg/
│   ├── models/           # Shared data structures
│   └── utils/            # Utility functions
└── bin/                  # Built executables
```

## Configuration

The application stores data in `~/.media-manager/`:
- `media.db` - SQLite database
- `thumbnails/` - Generated thumbnail cache (uniform 200×200 images)

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