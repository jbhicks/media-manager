# Current Task: Video Preview Generation (PornHub-style Animated Previews)

**Status**: 🚧 In Progress  
**Last Updated**: 2026-01-06

## Goal

Implement animated video preview generation that shows a smooth cross-section of various scenes throughout a video - similar to how adult video sites generate hover previews. The output should be an optimized animated GIF or WebM that provides a quick visual summary of the video content.

## Requirements

### Functional Requirements
1. Generate animated preview from any video file
2. Extract clips from multiple points throughout the video (8-15 segments)
3. Scale down to thumbnail size (160x90 to 320x180)
4. Reduce framerate for bandwidth (8-12 fps)
5. Generate optimized output (GIF with palette, or WebM)
6. Support scene-based detection as an option

### Quality Requirements
- **Robust Unit Testing**: ALL functionality must have unit tests
- **Build Verification**: `go build ./...` must pass before feature is complete
- **Test Verification**: `go test ./...` must pass before feature is complete
- **No Regressions**: Existing functionality must continue to work

## Algorithm

1. **Get video duration** - Use `ffprobe` to determine total length
2. **Calculate sample points** - Divide video into N equal segments
3. **Extract short clips** from each segment (0.5-1.5 seconds each)
4. **Scale down** to thumbnail dimensions
5. **Reduce framerate** to 8-12 fps
6. **Concatenate clips** into single output
7. **Generate optimized palette** (for GIF) or encode as WebM

## FFmpeg Implementation Reference

### Method 1: Scene-Based Preview (Auto-detect interesting scenes) ✅ IMPLEMENTED
```bash
# Extract frames at scene changes, tile them into a preview image
ffmpeg -i video.mp4 -vf "select='gt(scene,0.4)',scale=160:120,tile" -frames:v 1 preview.png
```

**Usage in Go:**
```go
opts := &preview.ScenePreviewOptions{
    SceneThreshold: 0.4,  // 0.0-1.0, higher = fewer scenes
    TileWidth:      160,  // Width per tile
    TileHeight:     120,  // Height per tile
    Cols:           4,    // 4 columns
    Rows:           4,    // 4 rows = 16 scenes total
}
err := preview.GenerateSceneBasedPreview(videoPath, outputPath, opts)
```

**Features:**
- Auto-detects scene changes above threshold
- Falls back to lower threshold if not enough scenes found
- Falls back to evenly distributed frames if no scene changes detected
- Outputs a static tiled image (not animated GIF)
- Configurable tile size and grid dimensions

### Method 2: Animated GIF Preview (Primary Method)

**Step 1: Extract clips at intervals**
```bash
ffmpeg -i input.mp4 \
  -vf "select='between(t,duration*0.10,duration*0.10+1)+between(t,duration*0.25,duration*0.25+1)+between(t,duration*0.40,duration*0.40+1)+between(t,duration*0.55,duration*0.55+1)+between(t,duration*0.70,duration*0.70+1)+between(t,duration*0.85,duration*0.85+1)',setpts=N/FRAME_RATE/TB,scale=320:-1,fps=10" \
  -t 6 preview_temp.mp4
```

**Step 2: Generate optimal palette for GIF**
```bash
ffmpeg -i preview_temp.mp4 -vf "palettegen=max_colors=256:stats_mode=diff" palette.png
```

**Step 3: Create GIF using palette**
```bash
ffmpeg -i preview_temp.mp4 -i palette.png -lavfi "paletteuse=dither=bayer:bayer_scale=3" -loop 0 preview.gif
```

### Method 3: All-in-One Command
```bash
DURATION=$(ffprobe -v error -show_entries format=duration -of csv=p=0 input.mp4)

ffmpeg -i input.mp4 \
  -vf "select='lt(mod(t,$DURATION/10),0.5)',setpts=N/10/TB,scale=320:-1:flags=lanczos,fps=10,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse" \
  -loop 0 preview.gif
```

### Method 4: WebM Preview (Better quality, smaller size)
```bash
ffmpeg -i input.mp4 \
  -vf "select='lt(mod(t,duration/12),0.8)',setpts=N/10/TB,scale=320:-1,fps=12" \
  -c:v libvpx-vp9 -crf 35 -b:v 0 -an \
  -t 8 preview.webm
```

### Sprite Sheet (For Scrubbing UI)
```bash
ffmpeg -i input.mp4 \
  -vf "select='not(mod(n,300))',scale=160:90,tile=10x10" \
  -frames:v 1 sprite.jpg
```

## Key FFmpeg Filters

| Filter | Purpose |
|--------|---------|
| `select` | Pick frames at specific times or scene changes |
| `thumbnail` | Auto-select representative frame from N frames |
| `scale` | Resize to thumbnail dimensions |
| `fps` | Reduce framerate for smaller file size |
| `setpts` | Reset timestamps after frame selection |
| `palettegen` | Generate optimal 256-color palette for GIF |
| `paletteuse` | Apply palette with dithering options |
| `tile` | Arrange frames in grid (for sprite sheets) |

## Configuration Parameters

| Parameter | Recommended | Notes |
|-----------|-------------|-------|
| Segment count | 8-15 | More segments = better coverage |
| Clip duration | 0.5-1.5s | Per segment |
| Output FPS | 8-12 | Lower = smaller file |
| Output size | 320x180 | 16:9 aspect ratio |
| GIF colors | 128-256 | Trade quality vs size |
| Dither method | `sierra2_4a` or `bayer` | For GIF quality |

## Implementation Plan

### Phase 1: Core Preview Generator
- [x] Create `internal/preview/video_preview.go` - (Methods in generator.go)
- [x] Implement `GetVideoDuration()` using ffprobe - (getVideoDuration)
- [x] Implement `GenerateAnimatedPreview()` - (existing)
- [x] Implement `GeneratePreviewGIF()` with palette optimization - (generateSceneBasedGIFWithCPU)
- [ ] Implement `GeneratePreviewWebM()` alternative
- [x] Implement `GenerateSceneBasedPreview()` - Method 1 ✅
- [x] **Unit tests for all functions**

### Phase 2: Configuration & Options
- [ ] Create `PreviewConfig` struct with configurable options
- [ ] Add segment count, clip duration, output size options
- [ ] Add scene-detection mode as alternative
- [ ] **Unit tests for configuration handling**

### Phase 3: Integration
- [ ] Integrate with existing preview generator
- [ ] Add CLI flags or config file options
- [ ] Update web UI to display animated previews on hover
- [ ] **Integration tests**

### Phase 4: Optimization
- [ ] Add caching to avoid regenerating existing previews
- [ ] Add progress callback for long videos
- [ ] Benchmark and optimize for performance
- [ ] **Performance tests**

## Testing Requirements

### Unit Tests (REQUIRED)
```go
// Example test structure
func TestGetVideoDuration(t *testing.T) { ... }
func TestGenerateAnimatedPreview(t *testing.T) { ... }
func TestGeneratePreviewGIF(t *testing.T) { ... }
func TestGeneratePreviewWebM(t *testing.T) { ... }
func TestPreviewConfig_Defaults(t *testing.T) { ... }
func TestPreviewConfig_Validation(t *testing.T) { ... }
func TestSceneDetection(t *testing.T) { ... }
```

### Test Files
- Use `media/` directory test videos (big_buck_bunny variants)
- Create `internal/preview/testdata/` for test-specific files

### Validation Before Completion
```bash
# MUST pass before marking feature complete
go vet ./...
go build ./...
go test ./... -v
```

## Files to Create/Modify

### New Files
- `internal/preview/video_preview.go` - Core implementation
- `internal/preview/video_preview_test.go` - Unit tests
- `internal/preview/config.go` - Configuration struct
- `internal/preview/config_test.go` - Config tests

### Modified Files
- `internal/preview/generator.go` - Integration point
- `AGENTS.md` - Document new preview functionality

## Success Criteria

- [ ] All unit tests pass
- [ ] `go build ./...` succeeds with no errors
- [ ] `go test ./...` succeeds with no failures
- [ ] `go vet ./...` reports no issues
- [ ] Can generate animated GIF preview from test video
- [ ] Can generate WebM preview from test video
- [ ] Preview shows clips from multiple points in video
- [ ] Output file size is reasonable (<2MB for 5-second preview)
- [ ] Preview generation completes in reasonable time (<30s for 1-hour video)

## Reference Documentation

- FFmpeg Documentation: https://ffmpeg.org/documentation.html
- FFmpeg Filters: https://ffmpeg.org/ffmpeg-filters.html
- Context7 library ID: `/websites/ffmpeg_documentation`

---

## Previous Task (Completed)

<details>
<summary>TV Show Poster Support - ✅ Completed 2025-12-25</summary>

Added comprehensive TV show poster support with automatic TV vs Movie detection.
- 96.6% poster coverage (up from 36%)
- Automatic TV show detection via patterns
- TMDb TV API integration with retry logic
- Full details in: `SESSION_4_TV_SHOW_POSTERS.md`

</details>
