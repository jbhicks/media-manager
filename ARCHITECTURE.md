# Architecture

## Overview

Media Manager is a native desktop application built with Go and Fyne for browsing, organizing, and managing media files. It includes an optional headless service mode for automated torrent downloads.

## Core Components

### Desktop GUI (`cmd/media-manager`)

```
┌─────────────────────────────────────────────────────────────┐
│                     Fyne Application                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌────────────────────────────────────┐   │
│  │   Sidebar   │  │           Media Grid               │   │
│  │  - Folders  │  │  ┌────────┐ ┌────────┐ ┌────────┐  │   │
│  │  - Tags     │  │  │MediaCard│ │MediaCard│ │MediaCard│  │   │
│  │  - Filters  │  │  │ 288x162 │ │ 288x162 │ │ 288x162 │  │   │
│  └─────────────┘  │  └────────┘ └────────┘ └────────┘  │   │
│                   └────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Service Mode (`cmd/media-manager-service`)

```
┌──────────────────────────────────────────────────────────────┐
│                    Download Service                          │
├──────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────┐    │
│  │ HTTP Server │  │ Download     │  │ Torrent Search   │    │
│  │ (Web UI)    │  │ Manager      │  │ (Jackett)        │    │
│  └─────────────┘  └──────────────┘  └──────────────────┘    │
│         │                │                    │              │
│         └────────────────┼────────────────────┘              │
│                          ▼                                   │
│              ┌───────────────────────┐                       │
│              │   SQLite Database     │                       │
│              │  - Download Rules     │                       │
│              │  - Tasks & History    │                       │
│              │  - Movie Metadata     │                       │
│              └───────────────────────┘                       │
└──────────────────────────────────────────────────────────────┘
```

## Package Structure

### `internal/ui/components`
Reusable Fyne widgets:
- **MediaCard**: 16:9 card with thumbnail, overlay labels, hover effects
- **AnimatedFrames**: Frame-based animation for video hover previews
- **VideoPreviewCard**: Specialized card with animated preview support

### `internal/preview`
Thumbnail and preview generation:
- Static thumbnail generation (images and videos)
- Animated GIF/frame preview generation
- Scene detection using FFmpeg
- GPU acceleration support (CUDA, VAAPI)
- Two-pass palette optimization for GIFs

### `internal/service`
Business logic for service mode:
- **DownloadManager**: Manages download tasks and rules
- **TMDbService**: Movie/TV metadata and poster fetching
- **HTTPServer**: Web UI and REST API
- **SuggestionService**: Download suggestion management

### `internal/torrent`
Torrent search providers:
- **Jackett**: Multi-indexer search aggregation
- Multi-query strategy (21 queries for movies)
- InfoHash-based deduplication

### `internal/db`
Database layer:
- GORM ORM with SQLite
- Models for media files, tags, download tasks
- Migration support

## Data Flow

### Thumbnail Generation
```
Media File → FFmpeg Probe (duration) → Generate Thumbnail
                                            │
                    ┌───────────────────────┴───────────────────────┐
                    ▼                                               ▼
              Static Image                                    Video Preview
              (mosaic grid)                              (scene-based frames)
                    │                                               │
                    └───────────────────────┬───────────────────────┘
                                            ▼
                              ~/.media-manager/thumbnails/
```

### Download Workflow
```
Download Rule → Search Jackett (21 queries) → Aggregate Results
                                                      │
                                            ┌─────────┴─────────┐
                                            ▼                   ▼
                                    Filter & Sort      Deduplicate by
                                    (seeders, size)    InfoHash & Title
                                            │                   │
                                            └─────────┬─────────┘
                                                      ▼
                                              Download Task
                                                      │
                                                      ▼
                                    Torrent Client (Transmission)
                                                      │
                                                      ▼
                                    Jellyfin Library Refresh
```

## Fyne Threading Model

**Critical**: All UI updates must run on the main thread.

```go
// Correct: Use fyne.Do() for UI updates from goroutines
go func() {
    result := expensiveOperation()
    fyne.Do(func() {
        widget.SetText(result)
        widget.Refresh()
    })
}()

// Incorrect: Direct UI updates from goroutines will crash
go func() {
    widget.SetText("text")  // CRASH!
}()
```

## Database Schema

### Core Tables
- `media_files` - Scanned media file metadata
- `tags` - Tag definitions
- `media_file_tags` - Many-to-many relationship
- `download_tasks` - Active and completed downloads
- `download_rules` - Automated download rules
- `download_suggestions` - Pending download suggestions
- `movie_metadata` - Cached TMDb data (movies and TV shows)
- `service_configs` - Service configuration

## External Dependencies

### Required
- **FFmpeg**: Video thumbnail and preview generation
- **SQLite**: Local database (via CGO)

### Optional
- **Jackett**: Torrent indexer aggregation
- **Transmission**: Torrent client for downloads
- **Jellyfin**: Media server integration
- **TMDb API**: Movie/TV metadata and posters

## Configuration Files

| File | Purpose |
|------|---------|
| `.env` | Environment variables (API keys) |
| `.air.toml` | Air auto-reload config (dev) |
| `docker-compose.yml` | Container orchestration |
| `Makefile` | Build and development tasks |

## Performance Considerations

1. **Thumbnail Caching**: All thumbnails cached in `~/.media-manager/thumbnails/`
2. **Lazy Loading**: Media cards load thumbnails on demand
3. **Frame Animation**: Pre-extracted frames for smooth hover previews
4. **Database Indexing**: SQLite indexes on frequently queried columns
5. **GPU Acceleration**: Optional CUDA/VAAPI for video processing
