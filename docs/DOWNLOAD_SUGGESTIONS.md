# Download Suggestions System

## Overview
The download suggestions system replaces automatic torrent downloads with a manual approval workflow. Users can review search results in a web UI and approve/reject downloads before they start.

## Architecture

### Database Layer
- **Model**: `DownloadSuggestion` in `pkg/models/download_models.go`
- **Fields**: 
  - `SourceID`, `Title`, `InfoHash` (unique), `MagnetLink`
  - `Size`, `Seeders`, `Leechers`, `Category`
  - `Status` (pending/approved/rejected)
  - `Notes`, `ApprovedAt`, `RejectedAt`, `DownloadedAt`
- **Indexes**: info_hash, seeders, upload_date, status

### Service Layer
- **SuggestionService** (`internal/service/suggestion_service.go`)
  - `GenerateSuggestions()` - Run multi-query search, save as pending
  - `ListSuggestions(status, limit, offset)` - Paginated listing
  - `ApproveSuggestion(id, notes)` - Creates DownloadTask in transaction
  - `RejectSuggestion(id, notes)` - Marks rejected
  - `BulkApprove(ids)`, `BulkReject(ids)` - Batch operations
  - `GetStats()` - Returns counts by status
  - `ClearRejected()` - Delete rejected items

- **DownloadManager** (`internal/service/download_manager.go`)
  - `SearchWithoutDownload(rule)` - Multi-query search without downloads

### HTTP API
- **Server**: `internal/service/http_server.go` (runs on port 8080)
- **Endpoints**:
  - `GET /api/suggestions` - List (params: status, limit, offset)
  - `POST /api/suggestions/generate` - Trigger generation
  - `GET /api/suggestions/stats` - Get counts
  - `POST /api/suggestions/approve` - Approve single (body: {id, notes})
  - `POST /api/suggestions/reject` - Reject single (body: {id, notes})
  - `POST /api/suggestions/approve-batch` - Bulk approve (body: {ids})
  - `POST /api/suggestions/reject-batch` - Bulk reject (body: {ids})
  - `POST /api/suggestions/clear-rejected` - Delete rejected
  - `GET /api/sources`, `/api/rules`, `/api/tasks` - Supporting data

### Web UI
- **Location**: `web/suggestions.html`
- **Features**:
  - Stats dashboard (pending/approved/rejected/total)
  - Suggestion cards with title, seeders, size, category
  - Individual approve/reject buttons
  - Batch select and approve/reject all
  - Status filtering (pending/all/approved/rejected)
  - Pagination (20 items/page)
  - Auto-refresh stats every 30s
  - Generate new suggestions button
  - Clear rejected items
  - Dark theme, responsive design

### CLI Tools
- **generate-suggestions** (`cmd/generate-suggestions/main.go`)
  - Standalone tool to generate suggestions
  - Uses first enabled download rule
  - Outputs JSON stats

## Workflow

1. **Generate Suggestions**:
   - User clicks "Generate New Suggestions" in web UI
   - OR runs `./generate-suggestions` from command line
   - System runs multi-query search (e.g., 21 genre queries for movies)
   - Results are filtered, sorted, deduplicated
   - Saved to `download_suggestions` table with status='pending'

2. **Review in Web UI**:
   - Access `http://localhost:8080/web/suggestions.html`
   - View pending suggestions with metadata
   - Filter by status, paginate through results

3. **Approve/Reject**:
   - **Approve**: Creates `DownloadTask`, updates suggestion status
   - **Reject**: Marks suggestion as rejected with timestamp
   - Both actions support optional notes

4. **Download Execution**:
   - Only approved tasks are downloaded
   - Download manager processes tasks as usual
   - No automatic downloads from search results

## Usage

### Start Service
```bash
./media-manager-service
```
- HTTP server runs on `http://localhost:8080`
- Access web UI at `http://localhost:8080/web/suggestions.html`

### Generate Suggestions (CLI)
```bash
./generate-suggestions
```
Output:
```json
{
  "suggestions_created": 50,
  "pending": 50,
  "approved": 0,
  "rejected": 0,
  "total": 50
}
```

### API Examples

**Generate suggestions**:
```bash
curl -X POST http://localhost:8080/api/suggestions/generate
```

**List pending**:
```bash
curl "http://localhost:8080/api/suggestions?status=pending&limit=20&offset=0"
```

**Approve**:
```bash
curl -X POST http://localhost:8080/api/suggestions/approve \
  -H "Content-Type: application/json" \
  -d '{"id": 123, "notes": "Looks good"}'
```

**Get stats**:
```bash
curl http://localhost:8080/api/suggestions/stats
```

## Testing

Run tests:
```bash
go test ./internal/service/... -run TestSuggestionService -v
```

All tests pass:
- `TestSuggestionService_CreateAndList`
- `TestSuggestionService_ApproveSuggestion`
- `TestSuggestionService_RejectSuggestion`
- `TestSuggestionService_GetStats`

## Files

### New Files
- `internal/service/suggestion_service.go` (236 lines)
- `internal/service/suggestion_service_test.go` (215 lines)
- `internal/service/http_server.go` (399 lines)
- `web/suggestions.html` (585 lines)
- `cmd/generate-suggestions/main.go` (93 lines)

### Modified Files
- `pkg/models/download_models.go` - Added DownloadSuggestion model
- `internal/db/database.go` - Added auto-migration
- `internal/service/download_manager.go` - Added SearchWithoutDownload()
- `internal/service/service.go` - Integrated HTTP server

## Configuration

No additional configuration needed. The system uses:
- Existing download rules from database
- Existing download sources (Jackett, RARBG, etc.)
- Existing filter/sort settings

## Security Notes

- CORS enabled for cross-origin access
- No authentication (runs locally)
- Database transactions for approve operations
- Unique constraint on info_hash prevents duplicates

## Future Enhancements

- Search/filter suggestions in web UI
- Torrent details modal with full metadata
- Scheduled suggestion generation (cron job)
- Email/notifications for new suggestions
- Quality scoring and auto-recommendations
- Bulk import from external sources
