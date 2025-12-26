package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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

	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("/home/josh/media-manager/web"))))

	// Clean URL routes that return full page or partial based on HX-Request header
	mux.HandleFunc("/library", s.handleLibraryPage)
	mux.HandleFunc("/search", s.handleSearchPage)
	mux.HandleFunc("/downloads", s.handleDownloadsPage)
	mux.HandleFunc("/settings", s.handleSettingsPage)

	// Search/torrent API endpoints
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/search/poster", s.handleFetchSearchPoster)
	mux.HandleFunc("/api/search/approve", s.handleApproveSearchResult)
	mux.HandleFunc("/api/search/reject", s.handleRejectSearchResult)
	mux.HandleFunc("/api/search/bulk-approve", s.handleBulkApprove)
	mux.HandleFunc("/api/search/bulk-reject", s.handleBulkReject)
	mux.HandleFunc("/api/search/clear", s.handleClearByStatus)
	mux.HandleFunc("/api/search/clear-rejected", s.handleClearRejected)

	// Suggestions API endpoints (used by suggestions.html)
	mux.HandleFunc("/suggestions", s.handleSuggestionsPage)
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

	// Legacy HTMX partial endpoints (kept for backwards compatibility)
	mux.HandleFunc("/partials/library", s.handleLibraryPartial)
	mux.HandleFunc("/partials/suggestions", s.handleSuggestionsPartial)
	mux.HandleFunc("/partials/downloads", s.handleDownloadsPartial)
	mux.HandleFunc("/partials/downloads/list", s.handleDownloadsListPartial)
	mux.HandleFunc("/partials/settings", s.handleSettingsPartial)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.corsMiddleware(mux),
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

func (s *HTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Check if HTMX request for partial content
	if r.Header.Get("HX-Request") == "true" {
		// Return home content partial
		fmt.Fprintf(w, `
		<div class="content-header">
			<h1>🏠 Welcome to Media Manager</h1>
			<p class="subtitle">Your personal media downloading and management system</p>
		</div>
		
		<div class="dashboard-grid">
			<a href="/downloads" class="dashboard-card" hx-get="/downloads" hx-target="#content" hx-push-url="true">
				<div class="card-icon">⬇️</div>
				<div class="card-title">Downloads</div>
				<div class="card-subtitle">Manage your download queue</div>
			</a>
			
			<a href="/library" class="dashboard-card" hx-get="/library" hx-target="#content" hx-push-url="true">
				<div class="card-icon">📚</div>
				<div class="card-title">Library</div>
				<div class="card-subtitle">Browse your collection</div>
			</a>
			
			<a href="/search" class="dashboard-card" hx-get="/search" hx-target="#content" hx-push-url="true">
				<div class="card-icon">🔍</div>
				<div class="card-title">Search</div>
				<div class="card-subtitle">Find movies and TV shows</div>
			</a>
			
			<div class="dashboard-card">
				<div class="card-icon">⚙️</div>
				<div class="card-title">Settings</div>
				<div class="card-subtitle">Configure your system</div>
			</div>
		</div>
		`)
		return
	}

	// Serve full page - read and return the HTML file directly
	html, err := os.ReadFile("/home/josh/media-manager/web/index.html")
	if err != nil {
		http.Error(w, "Failed to read HTML file", http.StatusInternalServerError)
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

// Stub handlers for compilation - these need to be implemented
func (s *HTTPServer) handleLibraryPage(w http.ResponseWriter, r *http.Request) {
	s.servePageOrPartial(w, r, s.handleLibraryPartial, "library")
}
func (s *HTTPServer) handleSuggestionsPage(w http.ResponseWriter, r *http.Request) {
	s.servePageOrPartial(w, r, s.handleSuggestionsPartial, "suggestions")
}
func (s *HTTPServer) handleDownloadsPage(w http.ResponseWriter, r *http.Request) {
	s.servePageOrPartial(w, r, s.handleDownloadsPartial, "downloads")
}
func (s *HTTPServer) handleSettingsPage(w http.ResponseWriter, r *http.Request) {
	s.servePageOrPartial(w, r, s.handleSettingsPartial, "settings")
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

	// Redirect to pending suggestions page to show new items
	w.Header().Set("HX-Redirect", "/suggestions?status=pending")
	w.WriteHeader(http.StatusOK)
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

	// Return empty response to remove the card
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, ``)
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

	// Return empty response to remove the card
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, ``)
}
func (s *HTTPServer) handleBulkApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form to get selected IDs
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	ids := r.Form["selected"]
	if len(ids) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Bulk approving %d items", len(ids))

	// Approve each selected item
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Printf("[HTTP] Invalid ID: %s", idStr)
			continue
		}

		if err := s.suggestionService.ApproveSuggestion(uint(id), "Bulk approved"); err != nil {
			log.Printf("[HTTP] Failed to approve suggestion %d: %v", id, err)
		}
	}

	// Refresh the search results
	w.Header().Set("HX-Redirect", "/search?status=approved")
	w.WriteHeader(http.StatusOK)
}
func (s *HTTPServer) handleBulkReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form to get selected IDs
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	ids := r.Form["selected"]
	if len(ids) == 0 {
		http.Error(w, "No items selected", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Bulk rejecting %d items", len(ids))

	// Reject each selected item
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Printf("[HTTP] Invalid ID: %s", idStr)
			continue
		}

		if err := s.suggestionService.RejectSuggestion(uint(id), "Bulk rejected"); err != nil {
			log.Printf("[HTTP] Failed to reject suggestion %d: %v", id, err)
		}
	}

	// Refresh the search results
	w.Header().Set("HX-Redirect", "/search?status=rejected")
	w.WriteHeader(http.StatusOK)
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

	// Redirect to search page
	w.Header().Set("HX-Redirect", "/search")
	w.WriteHeader(http.StatusOK)
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

	// Redirect to search page
	w.Header().Set("HX-Redirect", "/search")
	w.WriteHeader(http.StatusOK)
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

	w.Header().Set("Content-Type", "text/html")
	if successCount > 0 {
		fmt.Fprintf(w, `<div style="color: var(--accent-green); padding: 12px; background: rgba(46, 160, 67, 0.1); border-radius: 6px; margin-top: 12px;">
			✅ Updated %d posters (%d failed)<br>
			<small>Refresh the page to see updated posters</small>
		</div>`, successCount, failedCount)
	} else {
		fmt.Fprintf(w, `<div style="color: var(--text-muted); padding: 12px; background: rgba(128, 128, 128, 0.1); border-radius: 6px; margin-top: 12px;">
			ℹ️ No posters updated (%d missing, %d failed)<br>
			<small>Some movies may not be available on TMDb</small>
		</div>`, totalCount, failedCount)
	}
}

// New Search Page Handlers
func (s *HTTPServer) handleSearchPage(w http.ResponseWriter, r *http.Request) {
	s.servePageOrPartial(w, r, s.handleSearchPartial, "search")
}

func (s *HTTPServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		// If no query, show recent/saved searches from database
		s.handleSearchPartial(w, r)
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

	// Redirect to search results view with pending filter (query preserved in form)
	w.Header().Set("HX-Redirect", fmt.Sprintf("/search?status=pending&q=%s", url.QueryEscape(query)))
	w.WriteHeader(http.StatusOK)
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

	// If still empty, use placeholder
	posterURL := suggestion.PosterURL
	if posterURL == "" {
		posterURL = "/web/images/placeholder-movie.jpg"
	}

	// Return ONLY the updated poster image part to be swapped in
	fmt.Fprintf(w, `<img src="%s" alt="%s" class="poster-image" style="width: 100%%; height: 100%%; object-fit: cover;">`,
		posterURL, suggestion.Title)
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
func (s *HTTPServer) handleSources(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *HTTPServer) handleRules(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *HTTPServer) handleTasks(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
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

	// Return the updated task list
	s.handleDownloadsListPartial(w, r)
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

	// Return the updated task list
	s.handleDownloadsListPartial(w, r)
}
func (s *HTTPServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskID uint

	// Try to parse as form data first (from hx-vals)
	if err := r.ParseForm(); err == nil {
		if idStr := r.FormValue("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				taskID = uint(id)
			}
		}
	}

	// If form parsing didn't work, try JSON
	if taskID == 0 {
		var req struct {
			Method string `json:"_method,omitempty"`
			ID     uint   `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[HTTP] Failed to parse delete request: %v", err)
			http.Error(w, "Invalid request - missing task ID", http.StatusBadRequest)
			return
		}
		taskID = req.ID
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
	w.WriteHeader(http.StatusOK)
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

	// Return the updated downloads list
	s.handleDownloadsListPartial(w, r)
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

	// Return the updated downloads list
	s.handleDownloadsListPartial(w, r)
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
func (s *HTTPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *HTTPServer) handleLibraryMovies(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
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

	// Return success message
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `✅ Fetched %d new posters (%d already cached, %d failed)`, fetchedCount, cachedCount, failedCount)
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

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div style="color: var(--accent-green); padding: 12px; background: rgba(46, 160, 67, 0.1); border-radius: 6px; margin-top: 12px;">
		✅ Downloaded %d posters, %d failed, %d skipped<br>
		<small>Jellyfin will update automatically</small>
	</div>`, successCount, failedCount, skippedCount)
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
func (s *HTTPServer) handleLibraryPartial(w http.ResponseWriter, r *http.Request) {
	// Try to use Jellyfin first
	jellyfinClient := s.downloadManager.GetJellyfinClient()

	type MediaItem struct {
		Title     string
		Year      int
		PosterURL string
		Size      string
		Overview  string
		Rating    string
		Path      string
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
				posterURL := "/web/images/placeholder-movie.jpg"
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

				// Add year to display if available
				if year > 0 {
					displayTitle = fmt.Sprintf("%s (%d)", displayTitle, year)
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

						posterURL := "/web/images/placeholder-movie.jpg"
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

	// Render the page
	sourceIndicator := ""
	if dataSource == "jellyfin" {
		sourceIndicator = `<span style="color: var(--accent-green); font-size: 14px; margin-left: 10px;">📡 Powered by Jellyfin</span>`
	} else if dataSource == "filesystem" {
		sourceIndicator = `<span style="color: var(--text-muted); font-size: 14px; margin-left: 10px;">📁 Filesystem</span>`
	} else {
		sourceIndicator = `<span style="color: var(--accent-red); font-size: 14px; margin-left: 10px;">⚠️ Jellyfin Error - Using Filesystem</span>`
	}

	fmt.Fprintf(w, `
		<div class="content-header">
			<div style="display: flex; justify-content: space-between; align-items: center;">
				<div>
					<h1>📚 Library</h1>
					<p class="subtitle">Browse your media collection (%d items)%s</p>
				</div>
				<div style="display: flex; gap: 8px;">
					<button class="btn btn-secondary" 
						hx-post="/api/library/reprocess" 
						hx-target="#reprocess-result"
						hx-confirm="Reprocess all library files with enhanced filename cleaning? This will rename folders and files to remove quality tags and special characters.">
						🔄 Clean Filenames
					</button>
					<button class="btn btn-secondary" 
						hx-post="/api/library/refresh-jellyfin-posters" 
						hx-target="#reprocess-result"
						hx-confirm="Download correct posters from TMDb for all movies? This will save posters to movie folders and trigger Jellyfin to refresh metadata.">
						🎨 Fix Posters
					</button>
				</div>
			</div>
			<div id="reprocess-result" style="margin-top: 12px;"></div>
		</div>
		
		<div class="media-grid">
	`, len(mediaItems), sourceIndicator)

	if len(mediaItems) == 0 {
		fmt.Fprintf(w, `
			<div class="empty-state">
				<div class="empty-state-icon">📚</div>
				<h2>No Media Files Found</h2>
				<p>Your library will appear here once you download some media</p>
			</div>
		`)
	} else {
		for _, media := range mediaItems {
			// Build display title with year (if not already in title)
			displayTitle := media.Title

			// Check if title already contains the year in parentheses
			hasYearInTitle := false
			if media.Year > 0 {
				yearPattern := fmt.Sprintf("(%d)", media.Year)
				hasYearInTitle = strings.Contains(media.Title, yearPattern)
			}

			// Only append year if it's not already in the title
			if media.Year > 0 && !hasYearInTitle {
				displayTitle = fmt.Sprintf("%s (%d)", media.Title, media.Year)
			}

			fmt.Fprintf(w, `
				<div class="media-card">
					<div class="poster-container">
						<img src="%s" alt="%s" class="poster-image" loading="lazy">
					</div>
					<div class="media-content">
						<h3 class="media-title">%s</h3>
						<div class="media-meta">
							<span class="media-size">%s</span>
							%s
						</div>
					</div>
				</div>
			`, media.PosterURL, media.Title, displayTitle, media.Size,
				func() string {
					if media.Rating != "" {
						return "<span>•</span><span>" + media.Rating + "</span>"
					}
					return ""
				}())
		}
	}

	fmt.Fprintf(w, `</div>`)
}
func (s *HTTPServer) handleSuggestionsPartial(w http.ResponseWriter, r *http.Request) {
	// Get stats
	stats, err := s.suggestionService.GetStats()
	if err != nil {
		log.Printf("[HTTP] Failed to get suggestion stats: %v", err)
		stats = map[string]int64{"pending": 0, "approved": 0, "rejected": 0, "total": 0}
	}

	// Get suggestions list (pending only by default)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	suggestions, _, err := s.suggestionService.ListSuggestions(status, 50, 0)
	if err != nil {
		log.Printf("[HTTP] Failed to list suggestions: %v", err)
		suggestions = []models.DownloadSuggestion{}
	}

	fmt.Fprintf(w, `
		<div class="content-header">
			<h1>💡 Suggestions</h1>
			<p class="subtitle">Review and approve suggested downloads</p>
		</div>
		
		<div class="controls" style="display: flex; gap: 10px; margin-bottom: 20px; flex-wrap: wrap;">
			<button class="btn btn-primary" hx-post="/api/suggestions/generate" hx-indicator="#generate-spinner">
				<span id="generate-spinner" class="htmx-indicator">⏳</span>
				🔍 Generate Suggestions
			</button>
			<button class="btn btn-secondary" hx-post="/api/suggestions/clear-rejected" hx-confirm="Clear all rejected suggestions?">
				🗑️ Clear Rejected
			</button>
			<div style="flex: 1;"></div>
			<div style="display: flex; gap: 10px; align-items: center;">
				<span style="color: var(--text-muted);">Filter:</span>
				<button class="btn-filter %s" hx-get="/suggestions?status=pending" hx-target="#content" hx-push-url="true">
					Pending (%d)
				</button>
				<button class="btn-filter %s" hx-get="/suggestions?status=approved" hx-target="#content" hx-push-url="true">
					Approved (%d)
				</button>
				<button class="btn-filter %s" hx-get="/suggestions?status=rejected" hx-target="#content" hx-push-url="true">
					Rejected (%d)
				</button>
				<button class="btn-filter %s" hx-get="/suggestions?status=" hx-target="#content" hx-push-url="true">
					All (%d)
				</button>
			</div>
		</div>
		
		<div id="suggestions-content">
	`,
		ternary(status == "pending", "active", ""),
		stats["pending"],
		ternary(status == "approved", "active", ""),
		stats["approved"],
		ternary(status == "rejected", "active", ""),
		stats["rejected"],
		ternary(status == "", "active", ""),
		stats["total"],
	)

	if len(suggestions) == 0 {
		fmt.Fprintf(w, `
			<div class="empty-state">
				<div class="empty-state-icon">💡</div>
				<h2>No %s Suggestions</h2>
				<p>Click "Generate Suggestions" to find new content to download</p>
			</div>
		`, status)
	} else {
		fmt.Fprintf(w, `<div class="media-grid">`)
		for _, suggestion := range suggestions {
			sizeGB := float64(suggestion.Size) / (1024 * 1024 * 1024)
			posterURL := suggestion.PosterURL
			if posterURL == "" {
				posterURL = "/web/images/placeholder-movie.jpg"
			}

			statusBadge := ""
			if suggestion.Status == "approved" {
				statusBadge = `<span style="background: var(--accent-green); color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px;">✓ Approved</span>`
			} else if suggestion.Status == "rejected" {
				statusBadge = `<span style="background: var(--accent-red); color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px;">✗ Rejected</span>`
			}

			actions := ""
			if suggestion.Status == "pending" {
				actions = fmt.Sprintf(`
					<div style="display: flex; gap: 8px; margin-top: 12px;">
						<button class="btn-success" style="flex: 1;" 
							hx-post="/api/suggestions/approve?id=%d" 
							hx-target="closest .media-card" 
							hx-swap="outerHTML">
							✓ Approve
						</button>
						<button class="btn-danger" style="flex: 1;" 
							hx-post="/api/suggestions/reject?id=%d" 
							hx-target="closest .media-card" 
							hx-swap="outerHTML">
							✗ Reject
						</button>
					</div>
				`, suggestion.ID, suggestion.ID)
			}

			fmt.Fprintf(w, `
				<div class="media-card">
					<div class="poster-container">
						<img src="%s" alt="%s" class="poster-image" loading="lazy">
					</div>
					<div class="media-content">
						<h3 class="media-title">%s</h3>
						<div class="media-meta">
							<span class="media-size">%.2f GB</span>
							<span>•</span>
							<span>🌱 %d</span>
							%s
						</div>
						%s
					</div>
				</div>
			`, posterURL, suggestion.Title, suggestion.Title, sizeGB, suggestion.Seeders, statusBadge, actions)
		}
		fmt.Fprintf(w, `</div>`)
	}

	fmt.Fprintf(w, `</div>`) // Close suggestions-content
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

func (s *HTTPServer) handleSearchPartial(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	showResults := r.URL.Query().Get("results") == "true"
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "seeders" // Default sort
	}

	// Get stats for filter buttons
	stats, err := s.suggestionService.GetStats()
	if err != nil {
		log.Printf("[SEARCH] Failed to get stats: %v", err)
		stats = map[string]int64{"pending": 0, "approved": 0, "rejected": 0, "total": 0}
	}

	fmt.Fprintf(w, `
		<div class="content-header">
			<h1>🔍 Search</h1>
			<p class="subtitle">Search for movies and TV shows to download</p>
		</div>
		
		<div class="controls" style="margin-bottom: 20px;">
			<!-- Search Form -->
			<form hx-get="/api/search" hx-target="#search-results" hx-indicator="#search-indicator" style="display: flex; gap: 10px; margin-bottom: 16px; position: relative;">
				<input type="text" name="q" value="%s" placeholder="Search for movies (e.g., Interstellar, Matrix 1999)..." 
					style="flex: 1; padding: 12px 16px; background: rgba(22, 27, 34, 0.6); border: 1px solid rgba(48, 54, 61, 0.3); border-radius: 8px; color: var(--text-primary); font-size: 14px;"
					required>
				<div id="search-indicator" class="htmx-indicator">
					<div class="spinner"></div>
				</div>
				<button type="submit" class="btn btn-primary" style="min-width: 120px;">
					🔍 Search
				</button>
			</form>
			
			<!-- Status Filters -->
			<div style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap; margin-bottom: 12px;">
				<span style="color: var(--text-muted); font-weight: 500;">Status:</span>
				<button class="btn-filter %s" hx-get="/search?status=pending&sort=%s" hx-target="#content" hx-push-url="true">
					⏳ Pending (%d)
				</button>
				<button class="btn-filter %s" hx-get="/search?status=approved&sort=%s" hx-target="#content" hx-push-url="true">
					✅ Approved (%d)
				</button>
				<button class="btn-filter %s" hx-get="/search?status=rejected&sort=%s" hx-target="#content" hx-push-url="true">
					❌ Rejected (%d)
				</button>
				<button class="btn-filter %s" hx-get="/search?status=&sort=%s" hx-target="#content" hx-push-url="true">
					📋 All (%d)
				</button>
			</div>
			
			<!-- Sort & Bulk Actions Row -->
			<div style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap;">
				<span style="color: var(--text-muted); font-weight: 500;">Sort:</span>
				<select id="sort-select" 
					style="padding: 8px 12px; background: rgba(22, 27, 34, 0.6); border: 1px solid rgba(48, 54, 61, 0.3); border-radius: 6px; color: var(--text-primary); font-size: 14px;"
					hx-get="/search?status=%s" 
					hx-target="#content" 
					hx-include="this"
					hx-trigger="change"
					hx-push-url="true"
					name="sort">
					<option value="seeders" %s>🌱 Most Seeders</option>
					<option value="size-desc" %s>📦 Largest First</option>
					<option value="size-asc" %s>📦 Smallest First</option>
					<option value="date-desc" %s>🕒 Newest First</option>
					<option value="date-asc" %s>🕒 Oldest First</option>
				</select>
				
				<div style="flex: 1;"></div>
				
				<!-- Bulk Actions (only show for pending) -->
				<div id="bulk-actions" style="display: %s; gap: 8px;">
					<button class="btn btn-success" onclick="selectAll(true)" style="min-width: 100px;">
						☑️ Select All
					</button>
					<button class="btn btn-secondary" onclick="selectAll(false)" style="min-width: 100px;">
						⬜ Deselect All
					</button>
					<button class="btn btn-primary" 
						hx-post="/api/search/bulk-approve" 
						hx-include="[name='selected']:checked"
						hx-target="#search-results"
						hx-confirm="Approve selected items?">
						✅ Approve Selected
					</button>
					<button class="btn btn-danger" 
						hx-post="/api/search/bulk-reject" 
						hx-include="[name='selected']:checked"
						hx-target="#search-results"
						hx-confirm="Reject selected items?">
						❌ Reject Selected
					</button>
				</div>
				
				<!-- Clear Buttons -->
				<button class="btn btn-secondary" 
					hx-post="/api/search/clear?status=%s" 
					hx-target="#search-results"
					hx-confirm="Clear all %s items?">
					🗑️ Clear %s
				</button>
				<button class="btn btn-secondary" 
					hx-post="/api/suggestions/refresh-posters" 
					hx-target="#poster-refresh-result"
					hx-confirm="Search TMDb for missing posters? This will update all search results with missing cover art.">
					🎨 Fix Posters
				</button>
			</div>
			
			<!-- Result area for poster refresh -->
			<div id="poster-refresh-result" style="margin-top: 12px;"></div>
		</div>
		
		<script>
		function selectAll(checked) {
			document.querySelectorAll('[name="selected"]').forEach(cb => cb.checked = checked);
		}
		</script>
		
		<div id="search-results">
	`,
		query,
		// Status filter buttons with sort preservation
		ternary(status == "pending" || (query != "" && status == ""), "active", ""), sortBy,
		stats["pending"],
		ternary(status == "approved", "active", ""), sortBy,
		stats["approved"],
		ternary(status == "rejected", "active", ""), sortBy,
		stats["rejected"],
		ternary(status == "" && query == "", "active", ""), sortBy,
		stats["total"],
		// Sort dropdown
		status,
		ternary(sortBy == "seeders", "selected", ""),
		ternary(sortBy == "size-desc", "selected", ""),
		ternary(sortBy == "size-asc", "selected", ""),
		ternary(sortBy == "date-desc", "selected", ""),
		ternary(sortBy == "date-asc", "selected", ""),
		// Bulk actions (only show for pending)
		ternary(status == "pending", "flex", "none"),
		// Clear button
		status,
		ternary(status == "", "all", status),
		ternary(status == "", "All", capitalize(status)),
	)

	// Show search results or filtered saved results
	if showResults || status != "" || query != "" {
		// Determine which status to show
		filterStatus := status
		if (showResults || query != "") && filterStatus == "" {
			filterStatus = "pending"
		}

		// Get results from database - use SearchSuggestions if query is present, otherwise ListSuggestions
		var results []models.DownloadSuggestion
		var err error
		if query != "" {
			results, _, err = s.suggestionService.SearchSuggestions(query, filterStatus, sortBy, 0, 1000, 0)
		} else {
			results, _, err = s.suggestionService.ListSuggestions(filterStatus, 1000, 0)

			// Custom manual sorting for ListSuggestions since it only does seeders by default
			sort.Slice(results, func(i, j int) bool {
				switch sortBy {
				case "size-desc":
					return results[i].Size > results[j].Size
				case "size-asc":
					return results[i].Size < results[j].Size
				case "date-desc":
					return results[i].CreatedAt.After(results[j].CreatedAt)
				case "date-asc":
					return results[i].CreatedAt.Before(results[j].CreatedAt)
				case "seeders":
					fallthrough
				default:
					return results[i].Seeders > results[j].Seeders
				}
			})
		}

		if err != nil {
			log.Printf("[SEARCH] Failed to list results: %v", err)
			results = []models.DownloadSuggestion{}
		}

		if len(results) == 0 {
			emptyMsg := "Try searching for movies or TV shows above"
			if query != "" {
				emptyMsg = fmt.Sprintf("No results found for '%s'. Try a different search term.", query)
			}
			fmt.Fprintf(w, `
				<div class="empty-state">
					<div class="empty-state-icon">🔍</div>
					<h2>No Results Found</h2>
					<p>%s</p>
				</div>
			`, emptyMsg)
		} else {
			fmt.Fprintf(w, `<div class="media-grid">`)
			for _, result := range results {
				sizeGB := float64(result.Size) / (1024 * 1024 * 1024)

				// Poster logic: if empty, show skeleton and lazy load
				posterHtml := ""
				if result.PosterURL != "" {
					posterHtml = fmt.Sprintf(`<img src="%s" alt="%s" class="poster-image" style="width: 100%%; height: 100%%; object-fit: cover;">`,
						result.PosterURL, result.Title)
				} else {
					posterHtml = fmt.Sprintf(`
						<div class="skeleton-card" 
							 hx-get="/api/search/poster?id=%d" 
							 hx-trigger="load" 
							 hx-swap="outerHTML">
							<div class="shimmer"></div>
						</div>`, result.ID)
				}

				statusBadge := ""
				if result.Status == "approved" {
					statusBadge = `<span style="background: var(--accent-green); color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px;">✓ Approved</span>`
				} else if result.Status == "rejected" {
					statusBadge = `<span style="background: var(--accent-red); color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px;">✗ Rejected</span>`
				}

				// Checkbox for bulk selection (only for pending items)
				checkbox := ""
				if result.Status == "pending" {
					checkbox = fmt.Sprintf(`
						<div style="position: absolute; top: 8px; left: 8px; z-index: 10;">
							<input type="checkbox" name="selected" value="%d" 
								style="width: 20px; height: 20px; cursor: pointer;">
						</div>
					`, result.ID)
				}

				actions := ""
				if result.Status == "pending" {
					actions = fmt.Sprintf(`
						<div style="display: flex; gap: 8px; margin-top: 12px;">
							<button class="btn-success" style="flex: 1;" 
								hx-post="/api/search/approve?id=%d" 
								hx-target="closest .media-card" 
								hx-swap="outerHTML">
								✓ Approve
							</button>
							<button class="btn-danger" style="flex: 1;" 
								hx-post="/api/search/reject?id=%d" 
								hx-target="closest .media-card" 
								hx-swap="outerHTML">
								✗ Reject
							</button>
						</div>
					`, result.ID, result.ID)
				}

				// Add info icon if TMDb ID exists
				infoIcon := ""
				if result.TMDbID > 0 {
					infoIcon = fmt.Sprintf(`
						<button class="info-icon" 
							onclick="openModal(%d, '%s', '%s')"
							title="More info">
							ℹ️
						</button>
					`, result.TMDbID, result.ContentType, result.Title)
				}

				fmt.Fprintf(w, `
					<div class="media-card" style="position: relative;">
						%s
						<div class="poster-container" style="position: relative; padding-bottom: 150%%; overflow: hidden; border-bottom: 1px solid rgba(48, 54, 61, 0.3);">
							%s
							%s
						</div>
						<div class="media-content">
							<h3 class="media-title">%s</h3>
							<div class="media-meta">
								<span class="media-size">%.2f GB</span>
								<span>•</span>
								<span>🌱 %d</span>
								%s
							</div>
							%s
						</div>
					</div>
				`, checkbox, posterHtml, infoIcon, result.Title, sizeGB, result.Seeders, statusBadge, actions)
			}
			fmt.Fprintf(w, `</div>`)
		}
	} else {
		// Show empty state with search prompt
		fmt.Fprintf(w, `
			<div class="empty-state">
				<div class="empty-state-icon">🎬</div>
				<h2>Start Searching</h2>
				<p>Enter a movie or TV show name above to find downloads</p>
				<p style="margin-top: 16px; color: var(--text-muted); font-size: 14px;">
					Or browse your <a href="/search?status=pending" style="color: var(--accent-blue);">pending</a>, 
					<a href="/search?status=approved" style="color: var(--accent-green);">approved</a>, or 
					<a href="/search?status=rejected" style="color: var(--accent-red);">rejected</a> items
				</p>
			</div>
		`)
	}

	fmt.Fprintf(w, `</div>`) // Close search-results
}

func (s *HTTPServer) handleDownloadsPartial(w http.ResponseWriter, r *http.Request) {
	// Calculate actual stats
	stats := s.getDownloadStats()

	fmt.Fprintf(w, `
		<div class="content-header">
			<h1>⬇️ Downloads</h1>
			<p class="subtitle">Manage your active downloads</p>
		</div>
		
		<div class="controls">
			<select id="status-filter" hx-get="/partials/downloads/list" hx-target="#downloads-list" hx-include="[id='status-filter']" hx-trigger="change">
				<option value="">All</option>
				<option value="pending">Pending</option>
				<option value="downloading">Downloading</option>
				<option value="completed">Completed</option>
				<option value="failed">Failed</option>
			</select>
			<button hx-get="/partials/downloads/list" hx-target="#downloads-list" hx-include="[id='status-filter']" class="btn-secondary">
				🔄 Refresh
			</button>
			<button hx-post="/api/tasks/clear-completed" hx-confirm="Are you sure you want to delete all completed downloads?" hx-target="#downloads-list" hx-swap="outerHTML" class="btn-secondary">
				🗑️ Clear Completed
			</button>
			<button hx-post="/api/tasks/clear-failed" hx-confirm="Are you sure you want to delete all failed downloads?" hx-target="#downloads-list" hx-swap="outerHTML" class="btn-secondary">
				🗑️ Clear Failed
			</button>
		</div>

		<div class="stats-row">
			<div class="stat-card">
				<div class="stat-label">Pending</div>
				<div class="stat-value" id="stat-pending">%d</div>
			</div>
			<div class="stat-card">
				<div class="stat-label">Downloading</div>
				<div class="stat-value" id="stat-downloading">%d</div>
			</div>
			<div class="stat-card">
				<div class="stat-label">Completed</div>
				<div class="stat-value" id="stat-completed">%d</div>
			</div>
			<div class="stat-card">
				<div class="stat-label">Failed</div>
				<div class="stat-value" id="stat-failed">%d</div>
			</div>
		</div>

		<div id="downloads-list" 
			 hx-get="/partials/downloads/list" 
			 hx-trigger="load, every 2s" 
			 hx-include="[id='status-filter']">
			<div class="loading">Loading downloads...</div>
		</div>
	`, stats["pending"], stats["downloading"], stats["completed"], stats["failed"])
}
func (s *HTTPServer) handleDownloadsListPartial(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status-filter")

	tasks := s.downloadManager.GetAllTasks()

	// Filter tasks if status filter is provided
	var filteredTasks []models.DownloadTask
	if statusFilter != "" {
		for _, task := range tasks {
			if task.Status == statusFilter {
				filteredTasks = append(filteredTasks, task)
			}
		}
	} else {
		filteredTasks = tasks
	}

	if len(filteredTasks) == 0 {
		fmt.Fprintf(w, `
			<div class="empty-state">
				<h2>No download tasks found</h2>
				<p>Approve some suggestions to start downloading</p>
			</div>
		`)
		return
	}

	fmt.Fprintf(w, `<div class="tasks-list">`)

	for _, task := range filteredTasks {
		sizeGB := float64(task.Size) / (1024 * 1024 * 1024)

		statusClass := task.Status
		progressHTML := ""
		errorHTML := ""

		if task.Status == "downloading" {
			progressHTML = fmt.Sprintf(`
				<div class="progress-bar">
					<div class="progress-fill" style="width: %.1f%%"></div>
				</div>
				<div style="font-size: 12px; color: #999; margin-top: 4px;">
					%.1f%% complete
				</div>
			`, task.Progress, task.Progress)
		}

		if task.Error != "" {
			errorHTML = fmt.Sprintf(`
				<div style="color: #e74c3c; margin-top: 8px; font-size: 13px;">
					Error: %s
				</div>
			`, task.Error)
		}

		actionsHTML := ""
		if task.Status == "pending" || task.Status == "downloading" {
			actionsHTML = fmt.Sprintf(`
				<button class="btn-danger" hx-post="/api/tasks/cancel" hx-vals='{"id": %d}' hx-target="#downloads-list" hx-swap="outerHTML">
					Cancel
				</button>
				<button class="btn-secondary" hx-post="/api/tasks/delete" hx-vals='{"id": %d}' hx-target="closest .task-card" hx-swap="outerHTML swap:0.3s">
					🗑️ Delete
				</button>
			`, task.ID, task.ID)
		} else if task.Status == "failed" {
			actionsHTML = fmt.Sprintf(`
				<button class="btn-primary" hx-post="/api/tasks/restart" hx-vals='{"id": %d}' hx-target="#downloads-list" hx-swap="outerHTML">
					Retry
				</button>
				<button class="btn-secondary" hx-post="/api/tasks/delete" hx-vals='{"id": %d}' hx-target="closest .task-card" hx-swap="outerHTML swap:0.3s">
					🗑️ Delete
				</button>
			`, task.ID, task.ID)
		} else {
			// For completed tasks, just show delete
			actionsHTML = fmt.Sprintf(`
				<button class="btn-secondary" hx-post="/api/tasks/delete" hx-vals='{"id": %d}' hx-target="closest .task-card" hx-swap="outerHTML swap:0.3s">
					🗑️ Delete
				</button>
			`, task.ID)
		}

		fmt.Fprintf(w, `
			<div class="task-card %s">
				<div class="task-info">
					<div class="task-title">%s</div>
					<div class="task-meta">
						<div class="meta-item">
							<span class="meta-label">Status:</span>
							<span class="status-badge status-%s">%s</span>
						</div>
						<div class="meta-item">
							<span class="meta-label">Seeders:</span>
							<span class="meta-value seeders">%d</span>
						</div>
						<div class="meta-item">
							<span class="meta-label">Size:</span>
							<span class="meta-value size">%.2f GB</span>
						</div>
					</div>
					%s
					%s
				</div>
				<div class="task-actions">
					%s
				</div>
			</div>
		`, statusClass, task.Title, statusClass, task.Status, task.Seeders, sizeGB, progressHTML, errorHTML, actionsHTML)
	}

	fmt.Fprintf(w, `</div>`)
}
func (s *HTTPServer) handleSettingsPartial(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<div class="content-header">
			<h1>⚙️ Settings</h1>
			<p class="subtitle">Configure your media manager</p>
		</div>
		
		<div class="empty-state">
			<div class="empty-state-icon">⚙️</div>
			<h2>Settings Coming Soon</h2>
			<p>This section will allow you to configure download sources, quality preferences, and more</p>
		</div>
	`)
}
func (s *HTTPServer) servePageOrPartial(w http.ResponseWriter, r *http.Request, partialHandler func(http.ResponseWriter, *http.Request), activePage string) {
	// Check if HTMX request for partial content
	if r.Header.Get("HX-Request") == "true" {
		// Return just the partial
		partialHandler(w, r)
		return
	}

	// Return full page - read and return the index.html file
	html, err := os.ReadFile("/home/josh/media-manager/web/index.html")
	if err != nil {
		http.Error(w, "Failed to read HTML file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
}
