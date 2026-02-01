# TODO

## High Priority

### UI/UX Improvements
- [ ] Test new 16:9 card layout with various aspect ratios (4:3, portrait, square)
- [ ] Implement UI zoom (Ctrl +/-) for accessibility
- [ ] Add search and filtering for media files
- [ ] Add tag management interface
- [ ] Settings dialog for configuration

### Video Preview System
- [ ] Implement WebM preview alternative to GIF
- [ ] Add progress callback for long video processing
- [ ] Benchmark and optimize preview generation performance

### Media Loading
- [ ] Pass directory argument from main through to app and config layers
- [ ] Support dynamic media directories

## Medium Priority

### Performance
- [ ] Optimize thumbnail generation for large media libraries
- [ ] Add parallel processing for multiple preview generations
- [ ] Implement adaptive quality based on video complexity

### Features
- [ ] Advanced tagging (bulk tagging, color-coded tags)
- [ ] Robust sorting and filtering (by date, size, tags)
- [ ] Add support for additional media formats
- [ ] Duplicate detection improvements

### Service Mode
- [ ] Web UI for managing download rules
- [ ] RSS feed monitoring
- [ ] Sonarr/Radarr integration
- [ ] qBittorrent support
- [ ] Notification system (webhook, email, Discord)

## Low Priority

### GPU Acceleration
- [ ] Add user preference config option for GPU acceleration
- [ ] Test GPU encoding on different hardware configurations

### Documentation
- [ ] Add inline code documentation for complex functions
- [ ] Create user guide for service mode setup
- [ ] Add troubleshooting guide

## Completed

### January 2026
- [x] Media card redesign (16:9 aspect ratio, full-bleed images, overlay labels)
- [x] Scene-based video preview generation
- [x] Two-pass GIF palette optimization
- [x] GPU acceleration support (CUDA, VAAPI)
- [x] Static mosaic preview option

### December 2025
- [x] TV show poster support with automatic detection
- [x] TMDb TV API integration
- [x] Increased poster coverage from 36% to 97%

### Earlier
- [x] Torrent download manager with multi-query search
- [x] Jackett integration
- [x] InfoHash-based deduplication
- [x] Smart filtering (seeders, size, resolution, age)
- [x] Jellyfin integration
