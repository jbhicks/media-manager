package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

type HTTPServer struct {
	addr              string
	suggestionService *SuggestionService
	downloadManager   *DownloadManager
	tmdbService       *TMDbService
	db                *db.Database
	server            *http.Server
	sseClients        map[chan string]bool
	sseClientsMutex   sync.RWMutex
}

func NewHTTPServer(addr string, db *db.Database, dm *DownloadManager) *HTTPServer {
	tmdbService := NewTMDbService(db)
	server := &HTTPServer{
		addr:              addr,
		db:                db,
		downloadManager:   dm,
		suggestionService: NewSuggestionService(db, dm, tmdbService),
		tmdbService:       tmdbService,
		sseClients:        make(map[chan string]bool),
	}

	dm.SetUpdateCallback(func() {
		server.broadcastTaskUpdate()
	})

	return server
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// Force IPv4 binding
	if s.addr == ":" {
		s.addr = "0.0.0.0:"
	}

	// Serve static files from React build
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("/home/josh/media-manager/web/dist/assets"))))

	// Serve images from web/images/
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("/home/josh/media-manager/web/images"))))

	// API endpoints
	// Search/torrent API endpoints
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/search/poster", s.handleFetchSearchPoster)
	mux.HandleFunc("/api/search/approve", s.handleApproveSearchResult)
	mux.HandleFunc("/api/search/reject", s.handleRejectSearchResult)
	mux.HandleFunc("/api/search/bulk-approve", s.handleBulkApprove)
	mux.HandleFunc("/api/search/bulk-reject", s.handleBulkReject)
	mux.HandleFunc("/api/search/clear", s.handleClearByStatus)
	mux.HandleFunc("/api/search/clear-rejected", s.handleClearRejected)

	// Suggestions API endpoints
	mux.HandleFunc("/api/suggestions", s.handleSuggestionsAPI)
	mux.HandleFunc("/api/suggestions/stats", s.handleSuggestionsStatsAPI)
	mux.HandleFunc("/api/suggestions/search", s.handleSearchSuggestionsAPI)
	mux.HandleFunc("/api/suggestions/recommendations", s.handleRecommendationsAPI)
	mux.HandleFunc("/api/suggestions/quality-score", s.handleQualityScoreAPI)
	mux.HandleFunc("/api/suggestions/generate", s.handleGenerateSuggestions)
	mux.HandleFunc("/api/suggestions/approve", s.handleApproveSearchResult)
	mux.HandleFunc("/api/suggestions/reject", s.handleRejectSearchResult)
	mux.HandleFunc("/api/suggestions/approve-batch", s.handleBulkApprove)
	mux.HandleFunc("/api/suggestions/reject-batch", s.handleBulkReject)
	mux.HandleFunc("/api/suggestions/clear-rejected", s.handleClearRejected)
	mux.HandleFunc("/api/suggestions/refresh-posters", s.handleRefreshSearchPosters)
	mux.HandleFunc("/api/movie/details", s.handleMovieDetails)

	// Sources, Rules, Tasks, Stats API endpoints
	mux.HandleFunc("/api/sources", s.handleSources)
	mux.HandleFunc("/api/rules", s.handleRules)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/cancel", s.handleCancelTask)
	mux.HandleFunc("/api/tasks/restart", s.handleRestartTask)
	mux.HandleFunc("/api/tasks/delete", s.handleDeleteTask)
	mux.HandleFunc("/api/tasks/clear-completed", s.handleClearCompletedTasks)
	mux.HandleFunc("/api/tasks/clear-failed", s.handleClearFailedTasks)
	mux.HandleFunc("/api/tasks/reprocess", s.handleReprocessCompletedTasks)
	mux.HandleFunc("/api/tasks/stream", s.handleTasksStream)
	mux.HandleFunc("/api/downloads/stream", s.handleDownloadsStream)
	mux.HandleFunc("/api/stats", s.handleStats)

	// VPN status endpoint
	mux.HandleFunc("/api/vpn/status", s.handleVPNStatus)

	// Media library endpoints
	mux.HandleFunc("/api/library/movies", s.handleLibraryMovies)
	mux.HandleFunc("/api/library/poster", s.handleFetchPoster)
	mux.HandleFunc("/api/library/poster-by-title", s.handleFetchPosterByTitle)
	mux.HandleFunc("/api/library/fetch-all-posters", s.handleFetchAllPosters)
	mux.HandleFunc("/api/library/refresh-jellyfin-posters", s.handleRefreshJellyfinPosters)
	mux.HandleFunc("/api/library/reprocess", s.handleReprocessLibrary)
	mux.HandleFunc("/api/library/delete", s.handleDeleteMovie)
	mux.HandleFunc("/api/library/tags", s.handleTags)
	mux.HandleFunc("/api/library/tag/create", s.handleCreateTag)
	mux.HandleFunc("/api/library/tag/assign", s.handleAssignTag)
	mux.HandleFunc("/api/library/tag/remove", s.handleRemoveTag)

	// Catch-all handler for React SPA routing
	mux.HandleFunc("/", s.ServeReactSPA)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.recoveryMiddleware(s.corsMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("[HTTP] Starting HTTP server on %s", s.addr)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP] Server error: %v", err)
		}
	}()

	return nil
}

func (s *HTTPServer) Stop() error {
	log.Println("[HTTP] Stopping HTTP server")

	s.sseClientsMutex.Lock()
	for client := range s.sseClients {
		close(client)
		delete(s.sseClients, client)
	}
	s.sseClientsMutex.Unlock()

	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware catches panics and returns a 500 error
func (s *HTTPServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[HTTP] PANIC recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ServeReactSPA serves the React SPA index.html for all non-API routes
func (s *HTTPServer) ServeReactSPA(w http.ResponseWriter, r *http.Request) {
	// API routes should not be handled here
	if strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.HasPrefix(r.URL.Path, "/assets/") ||
		strings.HasPrefix(r.URL.Path, "/images/") {
		http.NotFound(w, r)
		return
	}

	// Serve the React app's index.html for all other routes
	indexPath := "/home/josh/media-manager/web/dist/index.html"
	html, err := os.ReadFile(indexPath)
	if err != nil {
		log.Printf("[HTTP] Failed to read React index.html: %v", err)
		http.Error(w, "Failed to load application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}

func (s *HTTPServer) broadcastTaskUpdate() {
	s.sseClientsMutex.RLock()
	defer s.sseClientsMutex.RUnlock()

	for client := range s.sseClients {
		select {
		case client <- "update":
		default:
		}
	}
}

func (s *HTTPServer) handleSuggestionsAPI(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	suggestions, total, err := s.suggestionService.ListSuggestions(status, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"suggestions": suggestions,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleSearchSuggestionsAPI(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	minSeedersStr := r.URL.Query().Get("min_seeders")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0
	minSeeders := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	if minSeedersStr != "" {
		if m, err := strconv.Atoi(minSeedersStr); err == nil {
			minSeeders = m
		}
	}

	suggestions, total, err := s.suggestionService.SearchSuggestions(query, status, sortBy, minSeeders, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"suggestions": suggestions,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleRecommendationsAPI(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 5

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	suggestions, err := s.suggestionService.GetTopRecommendations(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	// Format as recommendations with quality scores
	recommendations := make([]map[string]interface{}, 0, len(suggestions))
	for _, suggestion := range suggestions {
		score := s.suggestionService.CalculateTorrentQualityScore(suggestion.Title, suggestion.Size, suggestion.Seeders)
		recommendations = append(recommendations, map[string]interface{}{
			"suggestion":    suggestion,
			"quality_score": score,
		})
	}

	response := map[string]interface{}{
		"recommendations": recommendations,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleQualityScoreAPI returns the quality score for a suggestion
func (s *HTTPServer) handleQualityScoreAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var suggestion models.DownloadSuggestion
	if err := s.db.GetDB().First(&suggestion, id).Error; err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	score := s.suggestionService.CalculateTorrentQualityScore(suggestion.Title, suggestion.Size, suggestion.Seeders)

	response := map[string]interface{}{
		"quality_score": score,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the first enabled download rule
	var rule models.DownloadRule
	if err := s.db.GetDB().Where("enabled = ?", true).First(&rule).Error; err != nil {
		log.Printf("[HTTP] No enabled download rules found: %v", err)
		http.Error(w, "No enabled download rules found", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Generating suggestions using rule: %s", rule.Name)

	created, err := s.suggestionService.GenerateSuggestions(&rule, true)
	if err != nil {
		log.Printf("[HTTP] Failed to generate suggestions: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Generated %d new suggestions", created)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": created,
		"message": fmt.Sprintf("Generated %d new suggestions", created),
	})
}

func (s *HTTPServer) handleSuggestionsStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats, err := s.suggestionService.GetStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *HTTPServer) handleApproveSuggestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	if err := s.suggestionService.ApproveSuggestion(uint(id), ""); err != nil {
		log.Printf("[HTTP] Failed to approve suggestion %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to approve suggestion: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Approved suggestion %d", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"message": "Suggestion approved",
	})
}

func (s *HTTPServer) handleRejectSuggestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	if err := s.suggestionService.RejectSuggestion(uint(id), ""); err != nil {
		log.Printf("[HTTP] Failed to reject suggestion %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to reject suggestion: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Rejected suggestion %d", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      id,
		"message": "Suggestion rejected",
	})
}

func (s *HTTPServer) handleBulkApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body for IDs
	var req struct {
		IDs []uint `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Try form data as fallback
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		ids := r.Form["selected"]
		for _, idStr := range ids {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				continue
			}
			req.IDs = append(req.IDs, uint(id))
		}
	}

	if len(req.IDs) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Bulk approving %d items", len(req.IDs))

	// Approve each selected item
	successCount := 0
	for _, id := range req.IDs {
		if err := s.suggestionService.ApproveSuggestion(id, "Bulk approved"); err != nil {
			log.Printf("[HTTP] Failed to approve suggestion %d: %v", id, err)
		} else {
			successCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"approved": successCount,
		"failed":   len(req.IDs) - successCount,
		"message":  fmt.Sprintf("Approved %d items", successCount),
	})
}

func (s *HTTPServer) handleBulkReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body for IDs
	var req struct {
		IDs []uint `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Try form data as fallback
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		ids := r.Form["selected"]
		for _, idStr := range ids {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				continue
			}
			req.IDs = append(req.IDs, uint(id))
		}
	}

	if len(req.IDs) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Bulk rejecting %d items", len(req.IDs))

	// Reject each selected item
	successCount := 0
	for _, id := range req.IDs {
		if err := s.suggestionService.RejectSuggestion(id, "Bulk rejected"); err != nil {
			log.Printf("[HTTP] Failed to reject suggestion %d: %v", id, err)
		} else {
			successCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"rejected": successCount,
		"failed":   len(req.IDs) - successCount,
		"message":  fmt.Sprintf("Rejected %d items", successCount),
	})
}

func (s *HTTPServer) handleClearByStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := r.URL.Query().Get("status")

	// Build query based on status
	query := s.db.GetDB()
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// Delete suggestions
	result := query.Delete(&models.DownloadSuggestion{})
	if result.Error != nil {
		log.Printf("[HTTP] Failed to clear suggestions: %v", result.Error)
		http.Error(w, "Failed to clear suggestions", http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Cleared %d %s suggestions", result.RowsAffected, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cleared": result.RowsAffected,
		"status":  status,
		"message": fmt.Sprintf("Cleared %d suggestions", result.RowsAffected),
	})
}

func (s *HTTPServer) handleClearRejected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete all rejected suggestions
	result := s.db.GetDB().Where("status = ?", "rejected").Delete(&models.DownloadSuggestion{})
	if result.Error != nil {
		log.Printf("[HTTP] Failed to clear rejected suggestions: %v", result.Error)
		http.Error(w, "Failed to clear rejected suggestions", http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Cleared %d rejected suggestions", result.RowsAffected)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cleared": result.RowsAffected,
		"message": fmt.Sprintf("Cleared %d rejected suggestions", result.RowsAffected),
	})
}

func (s *HTTPServer) handleRefreshSearchPosters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.tmdbService == nil {
		http.Error(w, "TMDb service not available (set TMDB_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	log.Println("[HTTP] Starting search results poster refresh...")

	// Get all suggestions with missing posters
	var suggestions []models.DownloadSuggestion
	result := s.db.GetDB().Where("poster_url IS NULL OR poster_url = ?", "").Find(&suggestions)
	if result.Error != nil {
		log.Printf("[HTTP] Failed to get suggestions: %v", result.Error)
		http.Error(w, fmt.Sprintf("Failed to get suggestions: %v", result.Error), http.StatusInternalServerError)
		return
	}

	totalCount := len(suggestions)
	successCount := 0
	failedCount := 0

	for _, suggestion := range suggestions {
		log.Printf("[HTTP] Processing suggestion: %s", suggestion.Title)

		// Use FetchPosterForTask which automatically detects TV vs Movie
		posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(suggestion.Title)
		if err != nil {
			log.Printf("[HTTP] Failed to find on TMDb: %s - %v", suggestion.Title, err)
			failedCount++
			continue
		}

		if posterURL == "" {
			log.Printf("[HTTP] No poster available for: %s", suggestion.Title)
			failedCount++
			continue
		}

		// Update the suggestion with poster URL and TMDb ID
		updateResult := s.db.GetDB().Model(&models.DownloadSuggestion{}).
			Where("id = ?", suggestion.ID).
			Updates(map[string]interface{}{
				"poster_url": posterURL,
				"tmdb_id":    tmdbID,
			})

		if updateResult.Error != nil {
			log.Printf("[HTTP] Failed to update suggestion %d: %v", suggestion.ID, updateResult.Error)
			failedCount++
			continue
		}

		log.Printf("[HTTP] Updated poster for '%s': %s", suggestion.Title, posterURL)
		successCount++

		// Add delay to avoid rate limiting TMDb
		time.Sleep(300 * time.Millisecond)
	}

	log.Printf("[HTTP] Search poster refresh complete: %d total, %d success, %d failed", totalCount, successCount, failedCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": successCount,
		"failed":  failedCount,
		"total":   totalCount,
		"message": fmt.Sprintf("Updated %d posters (%d failed)", successCount, failedCount),
	})
}

// handleSearch performs torrent search and returns JSON
func (s *HTTPServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// Return empty results if no query
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []models.DownloadSuggestion{},
			"query":   "",
		})
		return
	}

	// Perform real-time torrent search
	log.Printf("[SEARCH] Searching for: %s", query)

	// Create a temporary download rule for searching
	rule := &models.DownloadRule{
		Name:        "Search: " + query,
		SearchQuery: query,
		Enabled:     true,
	}

	// Execute search without downloading
	results, err := s.downloadManager.SearchWithoutDownload(rule)
	if err != nil {
		log.Printf("[SEARCH] Search failed: %v", err)
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[SEARCH] Found %d results for '%s'", len(results), query)

	// Save results to database as suggestions (for approve/reject functionality)
	created := 0
	for _, result := range results {
		suggestion := &models.DownloadSuggestion{
			SourceID:   result.SourceID,
			Title:      result.Title,
			InfoHash:   result.InfoHash,
			MagnetLink: result.MagnetLink,
			TorrentURL: result.TorrentURL,
			Size:       result.Size,
			Seeders:    result.Seeders,
			Leechers:   result.Leechers,
			Category:   result.Category,
			UploadDate: result.UploadDate,
			Status:     "pending",
			PosterURL:  "",
			TMDbID:     0,
		}

		// Use FirstOrCreate to avoid duplicates
		if err := s.db.GetDB().
			Where("info_hash = ?", suggestion.InfoHash).
			FirstOrCreate(suggestion).Error; err != nil {
			log.Printf("[WARNING] Failed to save search result: %v", err)
			continue
		}
		created++
	}

	log.Printf("[SEARCH] Saved %d search results to database", created)

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"query":   query,
		"saved":   created,
	})
}

func (s *HTTPServer) handleFetchSearchPoster(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	if id == 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var suggestion models.DownloadSuggestion
	if err := s.db.GetDB().First(&suggestion, id).Error; err != nil {
		http.Error(w, "Suggestion not found", http.StatusNotFound)
		return
	}

	// Fetch poster from TMDb if not already present
	if suggestion.PosterURL == "" && s.tmdbService != nil {
		posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(suggestion.Title)
		if err == nil && posterURL != "" {
			suggestion.PosterURL = posterURL
			suggestion.TMDbID = tmdbID
			s.db.GetDB().Save(&suggestion)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         id,
		"poster_url": suggestion.PosterURL,
		"title":      suggestion.Title,
	})
}

func (s *HTTPServer) handleApproveSearchResult(w http.ResponseWriter, r *http.Request) {
	// Reuse existing approve suggestion logic
	s.handleApproveSuggestion(w, r)
}

func (s *HTTPServer) handleRejectSearchResult(w http.ResponseWriter, r *http.Request) {
	// Reuse existing reject suggestion logic
	s.handleRejectSuggestion(w, r)
}

func (s *HTTPServer) handleQualityScore(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleMovieDetails(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handleSources returns download sources as JSON
func (s *HTTPServer) handleSources(w http.ResponseWriter, r *http.Request) {
	var sources []models.DownloadSource
	if err := s.db.GetDB().Find(&sources).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch sources: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sources": sources,
	})
}

// handleRules returns download rules as JSON
func (s *HTTPServer) handleRules(w http.ResponseWriter, r *http.Request) {
	var rules []models.DownloadRule
	if err := s.db.GetDB().Find(&rules).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch rules: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
	})
}

// handleTasks returns download tasks as JSON
func (s *HTTPServer) handleTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.downloadManager.GetAllTasks()
	stats := s.getDownloadStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks": tasks,
		"stats": stats,
	})
}

func (s *HTTPServer) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID uint `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.downloadManager.CancelTask(req.ID); err != nil {
		log.Printf("[HTTP] Failed to cancel task %d: %v", req.ID, err)
		http.Error(w, fmt.Sprintf("Failed to cancel task: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Cancelled task %d", req.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      req.ID,
		"message": "Task cancelled",
	})
}

func (s *HTTPServer) handleRestartTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID uint `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.downloadManager.RestartTask(req.ID); err != nil {
		log.Printf("[HTTP] Failed to restart task %d: %v", req.ID, err)
		http.Error(w, fmt.Sprintf("Failed to restart task: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Restarted task %d", req.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      req.ID,
		"message": "Task restarted",
	})
}

func (s *HTTPServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskID uint

	// Try to parse as JSON first
	var req struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		taskID = req.ID
	} else {
		// Try form data as fallback
		if err := r.ParseForm(); err == nil {
			if idStr := r.FormValue("id"); idStr != "" {
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					taskID = uint(id)
				}
			}
		}
	}

	if taskID == 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := s.downloadManager.DeleteTask(taskID); err != nil {
		log.Printf("[HTTP] Failed to delete task %d: %v", taskID, err)
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Successfully deleted task %d", taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      taskID,
		"message": "Task deleted",
	})
}

func (s *HTTPServer) handleClearCompletedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := s.downloadManager.DeleteTasksByStatus("completed")
	if err != nil {
		log.Printf("[HTTP] Failed to clear completed tasks: %v", err)
		http.Error(w, fmt.Sprintf("Failed to clear completed tasks: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Cleared %d completed tasks", count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cleared": count,
		"message": fmt.Sprintf("Cleared %d completed tasks", count),
	})
}

func (s *HTTPServer) handleClearFailedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := s.downloadManager.DeleteTasksByStatus("failed")
	if err != nil {
		log.Printf("[HTTP] Failed to clear failed tasks: %v", err)
		http.Error(w, fmt.Sprintf("Failed to clear failed tasks: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Cleared %d failed tasks", count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cleared": count,
		"message": fmt.Sprintf("Cleared %d failed tasks", count),
	})
}

func (s *HTTPServer) handleReprocessCompletedTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := s.downloadManager.ReprocessCompletedDownloads()
	if err != nil {
		log.Printf("[HTTP] Failed to reprocess completed tasks: %v", err)
		http.Error(w, fmt.Sprintf("Failed to reprocess: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		Count   int    `json:"count"`
		Message string `json:"message"`
	}{
		Count:   count,
		Message: fmt.Sprintf("Reprocessed %d completed downloads", count),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleReprocessLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("[HTTP] Starting library reprocessing (filename cleanup)...")

	count, err := s.downloadManager.ReprocessLibraryFiles()
	if err != nil {
		log.Printf("[HTTP] Failed to reprocess library: %v", err)
		http.Error(w, fmt.Sprintf("Failed to reprocess library: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Library reprocessing completed: %d folders renamed", count)

	response := struct {
		Count   int    `json:"count"`
		Message string `json:"message"`
	}{
		Count:   count,
		Message: fmt.Sprintf("Reprocessed %d library folders with enhanced filename cleaning", count),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleTasksStream(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Add client to the map
	s.sseClientsMutex.Lock()
	s.sseClients[clientChan] = true
	s.sseClientsMutex.Unlock()

	// Remove client when connection closes
	defer func() {
		s.sseClientsMutex.Lock()
		delete(s.sseClients, clientChan)
		close(clientChan)
		s.sseClientsMutex.Unlock()
	}()

	// Send initial data
	tasks := s.downloadManager.GetAllTasks()
	if data, err := json.Marshal(tasks); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", data)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	// Listen for updates
	for {
		select {
		case update := <-clientChan:
			if update == "update" {
				tasks := s.downloadManager.GetAllTasks()
				if data, err := json.Marshal(tasks); err == nil {
					fmt.Fprintf(w, "data: %s\n\n", data)
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}
			}
		case <-r.Context().Done():
			return
		}
	}
}

func (s *HTTPServer) getDownloadStats() map[string]int {
	tasks := s.downloadManager.GetAllTasks()

	stats := map[string]int{
		"pending":     0,
		"downloading": 0,
		"completed":   0,
		"failed":      0,
	}

	for _, task := range tasks {
		stats[task.Status]++
	}

	return stats
}

func (s *HTTPServer) handleDownloadsStream(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Register client
	s.sseClientsMutex.Lock()
	s.sseClients[clientChan] = true
	s.sseClientsMutex.Unlock()

	// Send initial stats
	stats := s.getDownloadStats()
	statsJSON, _ := json.Marshal(stats)
	fmt.Fprintf(w, "data: %s\n\n", statsJSON)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Listen for updates
	ctx := r.Context()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			s.sseClientsMutex.Lock()
			delete(s.sseClients, clientChan)
			close(clientChan)
			s.sseClientsMutex.Unlock()
			return
		case <-clientChan:
			// Send updated stats when notified
			stats := s.getDownloadStats()
			statsJSON, _ := json.Marshal(stats)
			fmt.Fprintf(w, "data: %s\n\n", statsJSON)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-ticker.C:
			// Periodic keepalive and stats update
			stats := s.getDownloadStats()
			statsJSON, _ := json.Marshal(stats)
			fmt.Fprintf(w, "data: %s\n\n", statsJSON)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// handleStats returns server statistics as JSON
func (s *HTTPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP] handleStats called from %s", r.RemoteAddr)
	
	// Safe defaults
	response := map[string]interface{}{
		"downloads": map[string]int{
			"pending":     0,
			"downloading": 0,
			"completed":   0,
			"failed":      0,
		},
		"suggestions": map[string]int64{
			"pending":  0,
			"approved": 0,
			"rejected": 0,
			"total":    0,
		},
	}

	// Get download stats if available
	if s.downloadManager != nil {
		response["downloads"] = s.getDownloadStats()
	} else {
		log.Printf("[HTTP] Warning: downloadManager is nil")
	}

	// Get suggestion stats if available
	if s.suggestionService != nil {
		suggestionStats, err := s.suggestionService.GetStats()
		if err == nil && suggestionStats != nil {
			response["suggestions"] = suggestionStats
		} else if err != nil {
			log.Printf("[HTTP] Warning: GetStats error: %v", err)
		}
	} else {
		log.Printf("[HTTP] Warning: suggestionService is nil")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[HTTP] Error encoding stats response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	} else {
		log.Printf("[HTTP] handleStats completed successfully")
	}
}

// handleVPNStatus returns the current VPN connection status
func (s *HTTPServer) handleVPNStatus(w http.ResponseWriter, r *http.Request) {
	if s.downloadManager == nil {
		http.Error(w, "Download manager not available", http.StatusServiceUnavailable)
		return
	}

	vpnStatus := s.downloadManager.GetVPNStatus()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(vpnStatus); err != nil {
		log.Printf("[HTTP] Error encoding VPN status response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleLibraryMovies returns the media library as JSON
func (s *HTTPServer) handleLibraryMovies(w http.ResponseWriter, r *http.Request) {
	// Try to use Jellyfin first
	jellyfinClient := s.downloadManager.GetJellyfinClient()

	type MediaItem struct {
		Title     string `json:"title"`
		Year      int    `json:"year"`
		PosterURL string `json:"poster_url"`
		Size      string `json:"size"`
		Overview  string `json:"overview"`
		Rating    string `json:"rating"`
		Path      string `json:"path"`
	}

	var mediaItems []MediaItem
	var dataSource string

	if jellyfinClient != nil {
		// Use Jellyfin API
		log.Println("[LIBRARY] Using Jellyfin API to fetch library")
		items, err := jellyfinClient.GetLibraryItems("Movie")
		if err != nil {
			log.Printf("[LIBRARY] Failed to get Jellyfin items: %v", err)
			dataSource = "error"
		} else {
			dataSource = "jellyfin"
			for _, item := range items {
				posterURL := "/images/placeholder-movie.jpg"
				if imageTag, ok := item.ImageTags["Primary"]; ok {
					posterURL = jellyfinClient.GetImageURL(item.Id, imageTag)
				}

				// Calculate file size from media sources
				var totalSize int64
				for _, source := range item.MediaSources {
					totalSize += source.Size
				}
				sizeGB := float64(totalSize) / (1024 * 1024 * 1024)
				sizeStr := fmt.Sprintf("%.2f GB", sizeGB)
				if sizeGB == 0 {
					sizeStr = "Unknown"
				}

				// Format rating
				ratingStr := ""
				if item.CommunityRating > 0 {
					ratingStr = fmt.Sprintf("⭐ %.1f", item.CommunityRating)
				}

				// Clean the title using TMDb service to remove quality tags
				displayTitle := item.Name
				year := item.ProductionYear
				if s.tmdbService != nil {
					cleanName, extractedYear := s.tmdbService.ExtractMovieInfo(item.Name)
					if cleanName != "" {
						displayTitle = cleanName
					}
					if extractedYear > 0 && year == 0 {
						year = extractedYear
					}
				}

				mediaItems = append(mediaItems, MediaItem{
					Title:     displayTitle,
					Year:      year,
					PosterURL: posterURL,
					Size:      sizeStr,
					Overview:  item.Overview,
					Rating:    ratingStr,
					Path:      item.Path,
				})
			}
		}
	} else {
		dataSource = "filesystem"
	}

	// Fallback to filesystem if Jellyfin not available or failed
	if dataSource == "filesystem" || dataSource == "error" {
		log.Println("[LIBRARY] Using filesystem scan (Jellyfin not configured)")
		mediaDir := "/mnt/media/Movies"

		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".mkv" || ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".m4v" {
					if !strings.HasSuffix(path, ".part") {
						fileName := filepath.Base(path)
						fileDir := filepath.Base(filepath.Dir(path))

						title := fileDir
						if title == "Movies" || title == "." {
							title = strings.TrimSuffix(fileName, filepath.Ext(fileName))
						}

						sizeGB := float64(info.Size()) / (1024 * 1024 * 1024)
						sizeStr := fmt.Sprintf("%.2f GB", sizeGB)

						posterURL := "/images/placeholder-movie.jpg"
						movieName, year := s.tmdbService.ExtractMovieInfo(title)
						cleanMovieTitle := s.tmdbService.CleanTitle(movieName)

						if cleanMovieTitle != "" {
							var metadata models.MovieMetadata
							if err := s.db.GetDB().Where("clean_title = ?", cleanMovieTitle).
								First(&metadata).Error; err == nil && metadata.PosterURL != "" {
								posterURL = metadata.PosterURL
							}
						}

						mediaItems = append(mediaItems, MediaItem{
							Title:     title,
							Year:      year,
							PosterURL: posterURL,
							Size:      sizeStr,
							Path:      path,
						})
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("[LIBRARY] Error scanning media directory: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"movies":     mediaItems,
		"count":      len(mediaItems),
		"dataSource": dataSource,
	})
}

func (s *HTTPServer) handleFetchPoster(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleFetchPosterByTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Title parameter required", http.StatusBadRequest)
		return
	}

	if s.tmdbService == nil {
		http.Error(w, "TMDb service not available", http.StatusServiceUnavailable)
		return
	}

	posterURL, _, err := s.tmdbService.FetchPosterForTask(title)
	if err != nil {
		log.Printf("[HTTP] Failed to fetch poster for %s: %v", title, err)
		http.Error(w, "Failed to fetch poster", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"poster_url": posterURL,
	})
}

func (s *HTTPServer) handleFetchAllPosters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.tmdbService == nil {
		http.Error(w, "TMDb service not available", http.StatusServiceUnavailable)
		return
	}

	// Scan the media directory for all video files
	mediaDir := "/mnt/media/Downloads"

	type MediaFile struct {
		Path  string
		Title string
	}

	var mediaFiles []MediaFile
	err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// Check if it's a video file
			if ext == ".mkv" || ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".m4v" {
				// Skip partial downloads
				if !strings.HasSuffix(path, ".part") {
					fileName := filepath.Base(path)
					fileDir := filepath.Base(filepath.Dir(path))

					// Extract title from filename or directory
					title := fileDir
					if title == "Downloads" || title == "." {
						title = strings.TrimSuffix(fileName, filepath.Ext(fileName))
					}

					mediaFiles = append(mediaFiles, MediaFile{
						Path:  path,
						Title: title,
					})
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[HTTP] Error scanning media directory: %v", err)
		http.Error(w, "Failed to scan media directory", http.StatusInternalServerError)
		return
	}

	// Fetch posters for each file (with caching)
	fetchedCount := 0
	cachedCount := 0
	failedCount := 0

	for _, media := range mediaFiles {
		// Extract clean movie name and year using the same logic as TMDb service
		movieName, _ := s.tmdbService.ExtractMovieInfo(media.Title)
		cleanTitle := s.tmdbService.CleanTitle(movieName)

		// Check if already cached
		var metadata models.MovieMetadata
		result := s.db.GetDB().Where("clean_title = ?", cleanTitle).First(&metadata)

		if result.Error == nil && metadata.PosterURL != "" {
			// Already cached
			cachedCount++
			continue
		}

		// Fetch from TMDb (this will also cache it)
		posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(media.Title)
		if err != nil {
			log.Printf("[HTTP] Failed to fetch poster for '%s': %v", media.Title, err)
			failedCount++
			continue
		}

		if posterURL != "" && tmdbID > 0 {
			fetchedCount++
			log.Printf("[HTTP] Fetched poster for '%s': %s", media.Title, posterURL)
		}

		// Add a small delay to avoid rate limiting (250ms between requests)
		time.Sleep(250 * time.Millisecond)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"fetched": fetchedCount,
		"cached":  cachedCount,
		"failed":  failedCount,
		"total":   len(mediaFiles),
	})
}

func (s *HTTPServer) handleRefreshJellyfinPosters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jellyfinClient := s.downloadManager.GetJellyfinClient()
	if jellyfinClient == nil {
		http.Error(w, "Jellyfin not configured", http.StatusServiceUnavailable)
		return
	}

	if s.tmdbService == nil {
		http.Error(w, "TMDb service not available (set TMDB_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	log.Println("[HTTP] Starting Jellyfin poster refresh...")

	// Get all movie items from Jellyfin
	items, err := jellyfinClient.GetLibraryItems("Movie")
	if err != nil {
		log.Printf("[HTTP] Failed to get Jellyfin items: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get Jellyfin items: %v", err), http.StatusInternalServerError)
		return
	}

	successCount := 0
	failedCount := 0
	skippedCount := 0

	for _, item := range items {
		// Clean the title to search TMDb
		cleanName, year := s.tmdbService.ExtractMovieInfo(item.Name)
		if cleanName == "" {
			log.Printf("[HTTP] Skipping item with no clean name: %s", item.Name)
			skippedCount++
			continue
		}

		log.Printf("[HTTP] Processing: %s -> cleaned: %s (%d)", item.Name, cleanName, year)

		// Search TMDb for the correct movie
		movie, err := s.tmdbService.SearchMovie(cleanName, year)
		if err != nil || movie == nil {
			log.Printf("[HTTP] Failed to find movie on TMDb: %s (%d) - %v", cleanName, year, err)
			failedCount++
			continue
		}

		// Get the poster URL from TMDb
		posterURL := ""
		if movie.PosterPath != "" {
			posterURL = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", movie.PosterPath)
		}

		if posterURL == "" {
			log.Printf("[HTTP] No poster available for: %s", cleanName)
			failedCount++
			continue
		}

		// Download the poster image
		resp, err := http.Get(posterURL)
		if err != nil {
			log.Printf("[HTTP] Failed to download poster from %s: %v", posterURL, err)
			failedCount++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[HTTP] Poster download returned status %d", resp.StatusCode)
			failedCount++
			continue
		}

		// Determine where to save the poster
		// Jellyfin looks for poster.jpg/poster.png in the movie folder
		// Map Jellyfin's Docker path (/media/movies/) to actual path (/mnt/media/Movies/)
		movieDir := filepath.Dir(item.Path)
		movieDir = strings.Replace(movieDir, "/media/movies", "/mnt/media/Movies", 1)
		posterPath := filepath.Join(movieDir, "poster.jpg")

		// Create the poster file
		posterFile, err := os.Create(posterPath)
		if err != nil {
			log.Printf("[HTTP] Failed to create poster file %s: %v", posterPath, err)
			failedCount++
			continue
		}

		// Copy the poster data
		_, err = io.Copy(posterFile, resp.Body)
		posterFile.Close()
		if err != nil {
			log.Printf("[HTTP] Failed to write poster file %s: %v", posterPath, err)
			os.Remove(posterPath) // Clean up partial file
			failedCount++
			continue
		}

		log.Printf("[HTTP] Downloaded poster to: %s", posterPath)

		// Trigger Jellyfin to refresh this item's metadata
		// This will make Jellyfin detect the new poster.jpg file
		err = jellyfinClient.RefreshItemMetadata(item.Id, false)
		if err != nil {
			log.Printf("[HTTP] Failed to refresh Jellyfin metadata for %s: %v", item.Name, err)
			// Don't count as failed since we got the poster
		}

		successCount++

		// Add delay to avoid rate limiting TMDb
		time.Sleep(300 * time.Millisecond)
	}

	log.Printf("[HTTP] Jellyfin poster refresh complete: %d success, %d failed, %d skipped", successCount, failedCount, skippedCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": successCount,
		"failed":  failedCount,
		"skipped": skippedCount,
		"message": fmt.Sprintf("Downloaded %d posters, %d failed, %d skipped", successCount, failedCount, skippedCount),
	})
}

func (s *HTTPServer) handleDeleteMovie(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleTags(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleAssignTag(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *HTTPServer) handleRemoveTag(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Helper function for ternary operator
func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// capitalize returns a string with the first letter capitalized
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
