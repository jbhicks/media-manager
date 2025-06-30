# Media Manager - Architecture Plan

## Overview
A Go-based media management application that provides a native desktop interface for browsing, sorting, and tagging media files with animated previews.

## Key Architectural Decisions

### 1. Frontend Architecture
**Desktop Application with Fyne**
- Native desktop GUI using Fyne framework
- Benefits: True native performance, no browser overhead, excellent image handling
- Cross-platform: Windows, macOS, Linux
- Direct file system access and native dialogs

### 2. Database Choice
**SQLite**
- Single file, no setup required
- Good performance for local file management
- SQL queries for complex filtering/sorting
- GORM ORM for Go integration

### 3. Preview Generation Strategy
**On-demand generation with caching**
- Generate previews when first requested
- Cache to disk for subsequent requests
- Pro: Fast startup, Con: Slower first load
- Background processing for large directories

### 4. File Format Support
**Images**: JPEG, PNG, GIF, WebP, TIFF, BMP
**Videos**: MP4, AVI, MOV, MKV (generate thumbnail previews)
**Preview format**: JPEG for thumbnails, preserve original for display

## Proposed Architecture

### Backend Components
```
cmd/
├── media-manager/       # Main application entry point

internal/
├── app/                 # Application setup and main window
├── ui/                  # Fyne UI components and layouts
│   ├── components/      # Reusable UI components
│   ├── dialogs/         # Modal dialogs
│   └── views/           # Main application views
├── db/                  # Database layer and models
├── scanner/             # File system scanning
├── preview/             # Preview generation (thumbnails)
├── tagger/              # Tagging system
└── config/              # Configuration management

pkg/
├── models/              # Shared data structures
└── utils/               # Utility functions
```

### Data Models
```go
type MediaFile struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    Path        string    `json:"path" gorm:"uniqueIndex"`
    Filename    string    `json:"filename"`
    Size        int64     `json:"size"`
    ModTime     time.Time `json:"mod_time"`
    FileType    string    `json:"file_type"` // image, video
    MimeType    string    `json:"mime_type"`
    PreviewPath string    `json:"preview_path"`
    Width       int       `json:"width"`
    Height      int       `json:"height"`
    Duration    int       `json:"duration"` // for videos, in seconds
    Tags        []Tag     `json:"tags" gorm:"many2many:file_tags;"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
    ID    int64  `json:"id" gorm:"primaryKey"`
    Name  string `json:"name" gorm:"uniqueIndex"`
    Color string `json:"color"` // hex color for UI
}

type Folder struct {
    ID          int64     `json:"id" gorm:"primaryKey"`
    Path        string    `json:"path" gorm:"uniqueIndex"`
    Name        string    `json:"name"`
    LastScanned time.Time `json:"last_scanned"`
    FileCount   int       `json:"file_count"`
}
```

### UI Layout Structure
```go
// Main application layout
app := app.New()
window := app.NewWindow("Media Manager")

// Main layout: Border container
content := container.NewBorder(
    toolbar,     // Top: search bar, view options, filters
    statusBar,   // Bottom: selection info, progress
    sidebar,     // Left: folder tree, tags panel
    nil,         // Right: (empty for now)
    mediaGrid,   // Center: scrollable media thumbnail grid
)

window.SetContent(content)
```

### Technology Stack
- **GUI Framework**: Fyne v2 (native desktop GUI)
- **Database**: SQLite with GORM ORM
- **Preview Generation**: 
  - Images: Go's image package for thumbnails
  - Videos: FFmpeg via exec.Command for thumbnails
- **File Watching**: fsnotify for real-time updates
- **Image Loading**: Fyne's resource system + custom loaders

## Implementation Phases

### Phase 1: Core Desktop App
- Basic Fyne application structure
- File scanning and database storage
- Simple media grid view with thumbnails
- Basic folder navigation

### Phase 2: Enhanced UI
- Advanced thumbnail grid with smooth scrolling
- Tag management UI with drag & drop
- Search and filtering interface
- Preview panel for selected media

### Phase 3: Advanced Features
- Bulk operations (tagging, moving, deleting)
- Performance optimizations for large libraries
- Settings and preferences
- Keyboard shortcuts and hotkeys

## Key Implementation Details

### Preview Generation
- Generate 300x300px JPEG thumbnails on demand
- Cache in `~/.media-manager/thumbnails/` directory
- Use SHA-256 hash of file path as thumbnail filename
- Background worker for batch thumbnail generation

### File Scanning
- Watch specified directories with fsnotify
- Support for recursive directory scanning
- Skip hidden files and system directories
- Handle file renames and moves gracefully

### Performance Considerations
- Virtual scrolling for large media collections
- Lazy loading of thumbnails
- Background database operations
- Efficient image resource management

## Configuration
Environment variables and config file support:
- `MEDIA_DIRS`: Comma-separated list of directories to scan
- `DB_PATH`: SQLite database file path (default: `~/.media-manager/db.sqlite`)
- `THUMBNAIL_DIR`: Thumbnail cache directory
- `THUMBNAIL_SIZE`: Thumbnail dimensions (default: 300x300)