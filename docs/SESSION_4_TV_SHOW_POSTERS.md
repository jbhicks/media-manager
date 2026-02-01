# Session 4: TV Show Poster Support

**Date**: 2025-12-25  
**Status**: ✅ Completed  
**Impact**: Increased poster coverage from 36% to 97% (86 out of 89 suggestions)

## Problem

The web service only supported **movie poster** fetching from TMDb. Out of 89 download suggestions:
- **32 with posters** (36%) - Movies only
- **57 without posters** (64%) - TV shows not supported

## Solution Implemented

Added comprehensive TV show support to the TMDb service with automatic detection.

### 1. TV Show Detection (`internal/service/tmdb_service.go`)

**Added `MediaInfo` struct** (lines 113-120):
```go
type MediaInfo struct {
    Name      string
    Year      int
    Season    int
    Episode   int
    IsTV      bool
}
```

**Added `ExtractMediaInfo()` function** to detect TV vs Movie:
- Detects patterns: `S01E01`, `S01 E01`, `1x01`, `Season 1 Episode 1`
- Removes episode titles after pattern (e.g., "Shoresy S05E01 Keep It Simple" → "Shoresy")
- Preserves year in TV show titles (e.g., "All Creatures Great and Small 2020")
- Returns `MediaInfo` with `IsTV` flag, season, and episode numbers

### 2. TMDb TV API Integration

**Added TV show data structures** (lines 93-111):
```go
type TMDbTVSearchResult struct {
    Page         int        `json:"page"`
    Results      []TMDbShow `json:"results"`
    TotalPages   int        `json:"total_pages"`
    TotalResults int        `json:"total_results"`
}

type TMDbShow struct {
    ID           int     `json:"id"`
    Name         string  `json:"name"`
    PosterPath   string  `json:"poster_path"`
    BackdropPath string  `json:"backdrop_path"`
    FirstAirDate string  `json:"first_air_date"`
    Overview     string  `json:"overview"`
    VoteAverage  float64 `json:"vote_average"`
}
```

**Added `SearchTV()` function** (lines 297-348):
- Queries TMDb TV search API: `GET /search/tv`
- Uses `first_air_date_year` parameter for year filtering
- **Retry logic**: If search fails with year in title, retries without year
  - Example: "All Creatures Great and Small 2020" → "All Creatures Great and Small"

**Added helper functions**:
- `GetTVPosterURL()` - Returns poster URL for TV shows
- `GetTVBackdropURL()` - Returns backdrop URL for TV shows

### 3. Unified Poster Fetching

**Updated `FetchPosterForTask()`** (lines 393-476):
- Automatically detects if title is TV show or movie
- Routes to `SearchTV()` or `SearchMovie()` based on detection
- Caches results in `movie_metadata` table (works for both movies and TV)
- Returns poster URL and TMDb ID

**Flow**:
```
Title → ExtractMediaInfo() → IsTV?
                              ├─ Yes: SearchTV()
                              └─ No:  SearchMovie()
                                      ↓
                            Cache in database
                                      ↓
                          Return poster URL + TMDb ID
```

### 4. Updated HTTP Handler

**Modified `handleRefreshSearchPosters()`** (`internal/service/http_server.go:643-679`):
- **Before**: Only used `SearchMovie()` - movies only
- **After**: Uses `FetchPosterForTask()` - auto-detects TV vs movies
- Simplified from 50 lines to 35 lines of code

## Test Results

### Initial Test (5 samples)
```
✅ Shoresy S05E01                               → Found poster
✅ Adventure Time Fionna and Cake S02E10       → Found poster
✅ The Kardashians S07E10                      → Found poster
✅ All Creatures Great and Small 2020 S06E00   → Found poster (with retry)
✅ Die My Love 2025 (movie)                    → Found poster
```
**Success rate**: 100%

### Full Database Test (89 suggestions)
```bash
curl -X POST http://localhost:8080/api/suggestions/refresh-posters
```
**Results**:
- **86 posters found** (96.6% success rate)
- **3 failed**:
  1. "Mortimer and Whitehouse Gone Fishing S08E07" - Edge case
  2. "The Price Is Right 2025 12 24" (2 duplicates) - Daily game show

**Coverage by status**:
- Approved: 16/16 (100%)
- Pending: 54/57 (94.7%)
- Rejected: 16/16 (100%)

**Overall**: 86/89 = **96.6% coverage** (up from 36%)

## Files Modified

### Core Changes
1. **`internal/service/tmdb_service.go`**
   - Added `MediaInfo`, `TMDbShow`, `TMDbTVSearchResult` structs
   - Added `ExtractMediaInfo()` function (replaces old `ExtractMovieInfo`)
   - Added `SearchTV()` function with retry logic
   - Added `GetTVPosterURL()` and `GetTVBackdropURL()` helpers
   - Updated `FetchPosterForTask()` to support TV shows
   - Kept `ExtractMovieInfo()` for backward compatibility

2. **`internal/service/http_server.go`**
   - Simplified `handleRefreshSearchPosters()` to use `FetchPosterForTask()`
   - Removed manual movie-only logic

### Test Files
3. **`test_tv_poster.go`** (created for testing)
   - Tests TV show detection and poster fetching
   - Can be run with: `go run test_tv_poster.go`

## API Behavior

### Endpoint: `POST /api/suggestions/refresh-posters`

**Before**:
- Only searched TMDb movies API
- Ignored all TV shows
- 36% success rate

**After**:
- Auto-detects TV shows vs movies
- Searches appropriate TMDb API
- 96.6% success rate

**Example response**:
```html
<div style="color: var(--accent-green); ...">
    ✅ Updated 86 posters (3 failed)<br>
    <small>Refresh the page to see updated posters</small>
</div>
```

### API Output Example
```json
{
    "id": 123,
    "title": "Shoresy S05E01 Keep It Simple 1080p AMZN WEB-DL...",
    "poster_url": "https://image.tmdb.org/t/p/w500/aOnZyGIJCkG3gCMLWInGWRpM3HT.jpg",
    "tmdb_id": 131378,
    "status": "pending"
}
```

## Edge Cases Handled

### 1. Year in TV Show Title
**Problem**: "All Creatures Great and Small 2020" includes year as part of name  
**Solution**: Retry search without year if first attempt fails  
**Result**: ✅ Successfully found

### 2. Episode Titles
**Problem**: "Shoresy S05E01 Keep It Simple" includes episode title  
**Solution**: Extract everything before S##E## pattern  
**Result**: ✅ Clean title "Shoresy" → Found poster

### 3. Special Episodes
**Problem**: "S06E00 Christmas Special"  
**Solution**: Pattern handles S##E00 format  
**Result**: ✅ Detected as TV show

### 4. Movie Backward Compatibility
**Problem**: Existing movie poster code might break  
**Solution**: Kept `ExtractMovieInfo()` function for compatibility  
**Result**: ✅ Movies still work

## Performance

- **API Rate Limiting**: 300ms delay between requests (same as before)
- **Database Caching**: Results cached in `movie_metadata` table
- **Processing Time**: ~30 seconds for 89 suggestions (300ms × 89 + API time)

## Database Schema

**Table**: `movie_metadata`  
**Usage**: Caches TMDb results for both movies and TV shows

| Field | Type | Notes |
|-------|------|-------|
| `tmdb_id` | int | TMDb ID (movie or TV show) |
| `title` | string | Clean title |
| `clean_title` | string | Normalized for matching |
| `year` | int | Release/air year |
| `poster_url` | string | Full poster URL |
| `backdrop_url` | string | Full backdrop URL |
| `overview` | text | Description |
| `rating` | float | TMDb vote average |

## Web UI Impact

**Before**:
- 57 pending suggestions showed placeholder images
- Users couldn't identify shows visually

**After**:
- 54 pending suggestions show actual posters
- Visual browsing experience improved dramatically
- Only 3 edge cases remain without posters

## Future Enhancements

### Potential Improvements
1. **Fuzzy Matching**: Handle more edge cases like "Mortimer and Whitehouse"
2. **Episode-Level Posters**: Fetch specific episode stills from TMDb
3. **Season Posters**: Use season-specific posters instead of show posters
4. **Manual Override**: UI to manually assign TMDb ID for failed matches

### Known Limitations
1. Daily shows with date-based episodes (e.g., "The Price Is Right 2025 12 24")
2. Shows with non-standard naming patterns
3. Very new shows not yet in TMDb database

## Testing Commands

```bash
# Test TV show detection
go run test_tv_poster.go

# Trigger poster refresh via API
curl -X POST http://localhost:8080/api/suggestions/refresh-posters

# Check poster coverage
sqlite3 ~/.media-manager/media.db \
  "SELECT status, COUNT(*) as total, 
   SUM(CASE WHEN poster_url IS NOT NULL THEN 1 ELSE 0 END) as with_posters 
   FROM download_suggestions GROUP BY status;"

# View suggestions with posters
curl "http://localhost:8080/api/suggestions?status=pending&limit=5" | python3 -m json.tool
```

## Conclusion

**Mission Accomplished!** 🎉

- Increased poster coverage from **36% to 97%**
- Added TV show support with automatic detection
- Simplified codebase (reduced handler by 15 lines)
- Maintained backward compatibility
- All tests passing ✅

The web UI now provides a much better visual experience for browsing download suggestions, with nearly universal poster coverage across both movies and TV shows.
