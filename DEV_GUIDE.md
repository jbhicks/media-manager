# Development Guide for Media Manager

## Quick Start
```bash
# Initialize and run
go mod tidy
go run cmd/media-manager/main.go

# Development commands
go test ./...                           # Run all tests
go test ./internal/scanner              # Test specific package
go build -o bin/media-manager cmd/media-manager/main.go
```

## Project Structure
```
├── cmd/
│   └── media-manager/main.go          # Application entry point
├── internal/                          # Private application code
│   ├── app/                           # Application setup
│   ├── ui/                           # Fyne UI components
│   │   ├── components/               # Reusable components
│   │   ├── dialogs/                  # Modal dialogs
│   │   └── views/                    # Main views
│   ├── db/                           # Database layer
│   ├── scanner/                      # File scanning
│   ├── preview/                      # Thumbnail generation
│   └── config/                       # Configuration
├── pkg/                              # Public packages
│   ├── models/                       # Data structures
│   └── utils/                        # Utilities
├── go.mod
├── schema.sql                        # Database schema
└── ARCHITECTURE.md                   # Detailed architecture
```

## Key Technologies
- **Framework**: Fyne v2 (native desktop GUI)
- **Database**: SQLite with GORM ORM
- **Thumbnails**: Go image packages, FFmpeg for videos
- **File Watching**: fsnotify for real-time updates
- **Testing**: Go standard testing, testify for assertions

## Configuration
Environment variables:
- `MEDIA_DIRS`: Directories to scan for media files
- `DB_PATH`: SQLite database file path (default: ~/.media-manager/db.sqlite)
- `THUMBNAIL_DIR`: Thumbnail cache directory (default: ~/.media-manager/thumbnails)
- `THUMBNAIL_SIZE`: Thumbnail dimensions (default: 300)