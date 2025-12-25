# Torrent Download Manager

## Overview

The Media Manager includes an automated torrent download manager with multi-query search aggregation for building large, high-quality torrent datasets.

## Key Features

### 1. Multi-Query Search Aggregation

For movie searches without a specific query, the system automatically executes **21 separate searches**:
- 1 empty query (popular torrents)
- 20 genre-specific queries: action, adventure, comedy, drama, thriller, horror, sci-fi, fantasy, romance, crime, mystery, animation, documentary, family, superhero, war, western, musical, biography, sport

**Benefits:**
- Aggregates ~10x more results than single-query searches
- Better coverage across different genres and release groups
- Higher probability of finding optimal torrents
- Previous session demonstrated: 1085+ results vs ~108 from single query

### 2. InfoHash-Based Deduplication

All results are deduplicated by InfoHash across queries to prevent duplicate downloads.

### 3. Smart Filtering

Filter results by:
- **Seeders**: Min/max seeder counts
- **Size**: Min/max file size in bytes
- **Resolution**: 1080p, 720p, 4k, etc.
- **Upload Age**: Min/max age in days
- **Quality Tags**: Match specific quality patterns in titles

### 4. Flexible Sorting

Sort results by:
- `seeders`: Highest seeders first (most popular)
- `size`: Largest files first
- `upload_date`: Most recent uploads first
- `balanced`: Balanced score combining seeders and recency

### 5. Title Deduplication

Deduplicate results by normalized title to control how many versions of the same movie appear in the final list.

## Configuration

### Download Rule Structure

```go
type DownloadRule struct {
    SearchQuery        string  // Specific search term (empty = multi-query for movies)
    MediaType          string  // "movie" or "tv"
    Resolution         string  // "1080p", "720p", "4k", etc.
    MinSeeders         int     // Minimum seeders required
    MaxSeeders         int     // Maximum seeders (0 = unlimited)
    MinSize            int64   // Minimum size in bytes
    MaxSize            int64   // Maximum size in bytes
    MinUploadAge       int     // Minimum age in days
    MaxUploadAge       int     // Maximum age in days (0 = unlimited)
    MaxResults         int     // Maximum results to return
    MaxResultsPerTitle int     // Max results per unique title (deduplication)
    SortBy             string  // "seeders", "size", "upload_date", "balanced"
}
```

### Example Configuration

```go
rule := &models.DownloadRule{
    SearchQuery:        "",      // Empty = multi-query for movies
    MediaType:          "movie",
    Resolution:         "1080p",
    MinSeeders:         10,
    MinSize:            1024 * 1024 * 1024,      // 1 GB
    MaxSize:            20 * 1024 * 1024 * 1024, // 20 GB
    MaxUploadAge:       180,     // 6 months
    MaxResults:         100,
    MaxResultsPerTitle: 1,       // Only one version per movie
    SortBy:             "seeders",
}
```

## Usage

### Running the Demo Script

```bash
# Set your Jackett API key
export JACKETT_API_KEY=your-api-key-here

# Optionally customize Jackett URL (default: http://localhost:9117)
export JACKETT_URL=http://localhost:9117

# Run the demo
go run examples/show_download_list.go
```

The demo will:
1. Execute all 21 genre queries for movies
2. Aggregate and deduplicate by InfoHash
3. Filter by configured criteria
4. Sort by seeders (most popular first)
5. Deduplicate by title (max 1 per title)
6. Display top 100 results with detailed stats

### Output Format

```
═══════════════════════════════════════════════════════════════
  MULTI-QUERY TORRENT SEARCH - TARGETED DOWNLOAD LIST
═══════════════════════════════════════════════════════════════

Search Configuration:
  • Media Type: movie
  • Resolution: 1080p
  • Min Seeders: 10
  • Size Range: 1.0 GB - 20.0 GB
  • Max Age: 180 days (6.0 months)
  • Max Results: 100
  • Max Per Title: 1
  • Sort By: seeders

Generated 21 search queries:
   1. <empty> (popular torrents)
   2. action
   3. adventure
   ...

═══════════════════════════════════════════════════════════════
  EXECUTING SEARCHES...
═══════════════════════════════════════════════════════════════

[ 1/21] Searching: <popular>      ... ✓ Found 108, Added 108 unique (total: 108)
[ 2/21] Searching: action         ... ✓ Found 95, Added 87 unique (total: 195)
...

Total unique results collected: 1085

═══════════════════════════════════════════════════════════════
  FILTERING & PROCESSING...
═══════════════════════════════════════════════════════════════

After filtering (seeders, size, age, resolution): 453 results
After sorting by seeders: 453 results
After deduplication (max 1 per title): 312 results

═══════════════════════════════════════════════════════════════
  TOP 100 MOVIES READY FOR DOWNLOAD
═══════════════════════════════════════════════════════════════

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[1] Movie Title 1080p BluRay x264-GROUP
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  📊 Seeders: 523   | Leechers: 42    | Ratio: 12.45
  💾 Size: 8.2 GB
  📅 Upload Date: 2024-10-15 (age: 68 days / 2.3 months)
  🔑 InfoHash: abc123def456...
  🧲 Magnet: magnet:?xt=urn:btih:abc123...

...
```

## Implementation Details

### Multi-Query Search Flow

1. **Query Generation** (`generateSearchQueries()`):
   - Custom query → single query
   - Movies with empty query → 21 queries (empty + 20 genres)
   - TV shows with empty query → single empty query

2. **Search Execution**:
   - Execute all queries in sequence
   - Collect results from each query
   - Track InfoHash for deduplication

3. **Aggregation**:
   - Combine results from all queries
   - Deduplicate by InfoHash (prevents duplicate torrents)

4. **Filtering** (`filterResults()`):
   - Apply seeder constraints (min/max)
   - Apply size constraints (min/max)
   - Apply upload age constraints (min/max)
   - Match resolution patterns in titles

5. **Sorting** (`sortResults()`):
   - Sort by specified criteria
   - Balanced mode uses weighted score

6. **Title Deduplication** (`deduplicateResults()`):
   - Normalize titles (remove quality tags, years, etc.)
   - Limit results per normalized title
   - Keeps first N results per title (already sorted)

### Title Normalization

Removes these patterns for deduplication:
- Quality: 1080p, 720p, 4k, BluRay, WEB-DL, etc.
- Codecs: x264, x265, HEVC, etc.
- Audio: DTS, AC3, AAC, etc.
- Groups: YIFY, RARBG, etc.
- Years: 2010-2025
- Punctuation: dots, dashes, underscores

Example:
```
Input:  "The.Movie.2024.1080p.BluRay.x264-YIFY"
Output: "the movie"
```

## Known Issues

### go-jackett Library Date Parsing

**Issue**: The go-jackett library (v0.0.0-20220630233612) occasionally fails with "EOF" errors on date parsing.

**Impact**: 
- ~2/21 queries (9.5%) may fail
- 90%+ success rate still provides 1000+ results
- Minimal impact on final output quality

**Status**: **Accepted as-is**
- 19/21 queries succeeding = sufficient data
- Can revisit if success rate drops below 80%

**Alternatives Considered**:
1. ✅ Accept current state (recommended)
2. Fork and fix go-jackett (high effort)
3. Direct Torznab XML parsing (very high effort)
4. Find alternative library (uncertain availability)

## Testing

### Unit Tests

```bash
# Run all service tests
go test ./internal/service -v

# Run specific test
go test ./internal/service -run TestGenerateSearchQueries -v
```

### Integration Tests

```bash
# Run integration tests (requires Jackett setup)
INTEGRATION_TEST=1 go test ./internal/service -v
```

### Test Coverage

- ✅ Query generation (custom, movies, TV)
- ✅ Template expansion (year, month, date)
- ✅ Filtering (seeders, size, age, resolution)
- ✅ Sorting (all modes)
- ✅ Deduplication (InfoHash and title)
- ✅ Title normalization
- ✅ Balanced score calculation

## Future Enhancements

Potential improvements:
- Add more genre categories
- Support dynamic query templates
- Implement query priority/weighting
- Add cache warming strategies
- Support custom genre lists per user
- Add query performance metrics
- Implement adaptive query expansion
