package service

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
	"gorm.io/gorm/clause"
)

// jsonError sends a JSON error response with proper Content-Type
func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    code,
	})
}

type SearchActivity struct {
	Query           string                   `json:"query"`
	Timestamp       time.Time                `json:"timestamp"`
	Duration        float64                  `json:"duration_seconds"`
	ResultCount     int                      `json:"result_count"`
	SavedCount      int                      `json:"saved_count"`
	Providers       []string                 `json:"providers"`
	ProviderTimings []torrent.ProviderTiming `json:"provider_timings"`
	Error           string                   `json:"error,omitempty"`
}

type HTTPServer struct {
	addr                string
	suggestionService   *SuggestionService
	downloadManager     *DownloadManager
	tmdbService         *TMDbService
	db                  *db.Database
	server              *http.Server
	sseClients          map[chan string]bool
	sseClientsMutex     sync.RWMutex
	searchActivities    []SearchActivity
	searchActivityMutex sync.RWMutex
	streamHandler       *StreamHandler
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
		streamHandler:     NewStreamHandler("/mnt/media"),
	}

	dm.SetUpdateCallback(func() {
		server.broadcastTaskUpdate()
	})

	return server
}

func (s *HTTPServer) logSearchActivity(activity SearchActivity) {
	s.searchActivityMutex.Lock()
	defer s.searchActivityMutex.Unlock()

	// Keep only last 50 activities
	s.searchActivities = append(s.searchActivities, activity)
	if len(s.searchActivities) > 50 {
		s.searchActivities = s.searchActivities[len(s.searchActivities)-50:]
	}
}

func (s *HTTPServer) getSearchActivities() []SearchActivity {
	s.searchActivityMutex.RLock()
	defer s.searchActivityMutex.RUnlock()

	// Return a copy in reverse order (newest first)
	result := make([]SearchActivity, len(s.searchActivities))
	for i := range s.searchActivities {
		result[i] = s.searchActivities[len(s.searchActivities)-1-i]
	}
	return result
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// Force IPv4 binding - always bind to all interfaces explicitly
	if strings.HasPrefix(s.addr, ":") {
		s.addr = "0.0.0.0" + s.addr
	}

	// Get web root from environment or use default
	webRoot := os.Getenv("WEB_ROOT")
	if webRoot == "" {
		webRoot = "/home/josh/media-manager/web/dist"
	}

	// Serve static files from React build
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(webRoot, "assets")))))

	// Serve images from web/images/
	imagesRoot := os.Getenv("IMAGES_ROOT")
	if imagesRoot == "" {
		imagesRoot = "/home/josh/media-manager/web/images"
	}
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(imagesRoot))))

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
	mux.HandleFunc("/api/search/extract-images", s.handleExtractTorrentImages)
	mux.HandleFunc("/api/search/activity", s.handleSearchActivity)
	mux.HandleFunc("/api/logs", s.handleLogs)

	// Suggestions API endpoints
	mux.HandleFunc("/api/suggestions", s.handleSuggestionsAPI)
	mux.HandleFunc("/api/suggestions/stats", s.handleSuggestionsStatsAPI)
	mux.HandleFunc("/api/suggestions/search", s.handleSearchSuggestionsAPI)
	mux.HandleFunc("/api/suggestions/grouped", s.handleGroupedSuggestionsAPI)
	mux.HandleFunc("/api/suggestions/grouped/search", s.handleSearchGroupedSuggestionsAPI)
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
	mux.HandleFunc("/api/sources/jackett-indexers", s.handleAllJackettIndexers)
	mux.HandleFunc("/api/sources/", s.handleSourceDetail)
	mux.HandleFunc("/api/sources/create", s.handleCreateSource)
	mux.HandleFunc("/api/sources/update", s.handleUpdateSource)
	mux.HandleFunc("/api/sources/delete", s.handleDeleteSource)
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

	// RSS feed endpoints
	mux.HandleFunc("/api/rss", s.handleRSSFeeds)
	mux.HandleFunc("/api/rss/create", s.handleCreateRSSFeed)
	mux.HandleFunc("/api/rss/update", s.handleUpdateRSSFeed)
	mux.HandleFunc("/api/rss/delete", s.handleDeleteRSSFeed)
	mux.HandleFunc("/api/rss/check", s.handleCheckRSSFeed)

	// VPN endpoints
	mux.HandleFunc("/api/vpn/status", s.handleVPNStatus)
	mux.HandleFunc("/api/vpn/connect", s.handleVPNConnect)
	mux.HandleFunc("/api/vpn/disconnect", s.handleVPNDisconnect)

	// Media library endpoints
	mux.HandleFunc("/api/library/movies", s.handleLibraryMovies)
	mux.HandleFunc("/api/library/poster", s.handleFetchPoster)
	mux.HandleFunc("/api/library/poster-by-title", s.handleFetchPosterByTitle)
	mux.HandleFunc("/api/library/poster-file", s.handlePosterFile)
	mux.HandleFunc("/api/library/fetch-all-posters", s.handleFetchAllPosters)
	mux.HandleFunc("/api/library/refresh-jellyfin-posters", s.handleRefreshJellyfinPosters)
	mux.HandleFunc("/api/library/reprocess", s.handleReprocessLibrary)
	mux.HandleFunc("/api/library/delete", s.handleDeleteMovie)
	mux.HandleFunc("/api/library/tags", s.handleTags)
	mux.HandleFunc("/api/library/tag/create", s.handleCreateTag)
	mux.HandleFunc("/api/library/tag/assign", s.handleAssignTag)
	mux.HandleFunc("/api/library/tag/remove", s.handleRemoveTag)

	// Auth endpoints
	authHandler := NewAuthHandler(s.db)
	authHandler.RegisterRoutes(mux)

	// Discover endpoints
	discoverEndpoints := NewDiscoverEndpoints(s.tmdbService)
	discoverEndpoints.RegisterRoutes(mux)

	// Watchlist endpoints
	watchlistEndpoints := NewWatchlistEndpoints(s.db)
	watchlistEndpoints.RegisterRoutes(mux)

	// Watch history endpoints
	historyEndpoints := NewWatchHistoryEndpoints(s.db)
	historyEndpoints.RegisterRoutes(mux)

	// Streaming endpoints
	mux.HandleFunc("/api/stream/init", s.streamHandler.HandleStreamInit)
	mux.HandleFunc("/api/stream/playlist", s.streamHandler.HandleStreamPlaylist)
	mux.HandleFunc("/api/stream/segment", s.streamHandler.HandleStreamSegment)
	mux.HandleFunc("/api/stream/status", s.streamHandler.HandleStreamStatus)
	mux.HandleFunc("/api/stream/stop", s.streamHandler.HandleStreamStop)
	mux.HandleFunc("/api/stream/direct", s.streamHandler.HandleDirectStream)

	// Health check endpoint
	mux.HandleFunc("/api/health", s.handleHealth)

	// Debug endpoint to check TMDB API key status
	mux.HandleFunc("/api/debug/tmdb", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tmdb_api_key_configured": s.tmdbService.apiKey != "",
			"tmdb_api_key_length":     len(s.tmdbService.apiKey),
		})
	})

	// Catch-all handler for React SPA routing
	mux.HandleFunc("/", s.ServeReactSPA)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.recoveryMiddleware(s.corsMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
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
	webRoot := os.Getenv("WEB_ROOT")
	if webRoot == "" {
		webRoot = "/home/josh/media-manager/web/dist"
	}
	indexPath := filepath.Join(webRoot, "index.html")
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

	limit := 0
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
		"data":   suggestions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
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

	limit := 0
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
		"data":   suggestions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleGroupedSuggestionsAPI(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
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

	groups, total, err := s.suggestionService.ListGroupedSuggestions(status, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list grouped suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":   groups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) handleSearchGroupedSuggestionsAPI(w http.ResponseWriter, r *http.Request) {
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

	groups, total, err := s.suggestionService.SearchGroupedSuggestions(query, status, sortBy, minSeeders, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search grouped suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":   groups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
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

	// Check for auto-start parameter
	autoStart := r.URL.Query().Get("auto_start") == "true"

	if err := s.suggestionService.ApproveSuggestion(uint(id), "", autoStart); err != nil {
		log.Printf("[HTTP] Failed to approve suggestion %d: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to approve suggestion: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Approved suggestion %d (auto_start: %v)", id, autoStart)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"id":         id,
		"message":    "Suggestion approved",
		"auto_start": autoStart,
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
		if err := s.suggestionService.ApproveSuggestion(id, "Bulk approved", false); err != nil {
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

	// Check if download manager is available
	if s.downloadManager == nil {
		jsonError(w, "Search service not available", http.StatusServiceUnavailable)
		return
	}

	searchStart := time.Now()

	// Get enabled sources for activity logging and response
	var enabledSources []models.DownloadSource
	s.db.GetDB().Where("enabled = ?", true).Find(&enabledSources)
	sourceMap := make(map[uint]*models.DownloadSource)
	providerNames := make([]string, len(enabledSources))
	for i, src := range enabledSources {
		providerNames[i] = src.Name
		sourceMap[src.ID] = &enabledSources[i]
	}

	// Parse selected indexers from query
	indexersParam := r.URL.Query().Get("indexers")
	var selectedIndexers []string
	if indexersParam != "" {
		selectedIndexers = strings.Split(indexersParam, ",")
		log.Printf("[SEARCH] Using specific indexers: %v", selectedIndexers)
	}

	// Perform real-time torrent search
	log.Printf("[SEARCH] Searching for: %s", query)

	rule := &models.DownloadRule{
		Name:        "Search: " + query,
		SearchQuery: query,
		Enabled:     true,
		Indexers:    indexersParam,
	}

	searchResult, err := s.downloadManager.SearchWithoutDownload(rule)
	if err != nil {
		log.Printf("[SEARCH] Search failed: %v", err)
		s.logSearchActivity(SearchActivity{
			Query:           query,
			Timestamp:       time.Now(),
			Duration:        time.Since(searchStart).Seconds(),
			Providers:       providerNames,
			ProviderTimings: searchResult.Timings,
			Error:           err.Error(),
		})
		jsonError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	results := searchResult.Results
	totalResults := len(results)
	log.Printf("[SEARCH] Found %d total results for '%s'", totalResults, query)

	// Apply pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	if limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		offset, _ := strconv.Atoi(offsetStr)
		if limit > 0 {
			if offset < len(results) {
				end := offset + limit
				if end > len(results) {
					end = len(results)
				}
				results = results[offset:end]
			} else {
				results = []models.SearchResult{}
			}
			log.Printf("[SEARCH] Paginated to %d results (offset: %d, limit: %d)", len(results), offset, limit)
		}
	}

	// Build suggestions from search results directly (no redundant DB query)
	suggestions := make([]models.DownloadSuggestion, 0, len(results))
	for _, result := range results {
		suggestion := models.DownloadSuggestion{
			SourceID:   result.SourceID,
			Indexer:    result.Indexer,
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
			PosterURL:  result.PosterURL,
			TMDbID:     0,
		}
		// Populate source relation from memory
		if src, ok := sourceMap[result.SourceID]; ok {
			suggestion.Source = src
		}
		suggestions = append(suggestions, suggestion)
	}

	// Batch insert with ON CONFLICT DO NOTHING for performance
	created := 0
	if len(suggestions) > 0 {
		// Use Clauses to ignore duplicates on unique index (info_hash)
		result := s.db.GetDB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "info_hash"}},
			DoNothing: true,
		}).CreateInBatches(suggestions, 100)

		if result.Error != nil {
			log.Printf("[WARNING] Failed to batch save search results: %v", result.Error)
		} else {
			created = int(result.RowsAffected)
			log.Printf("[SEARCH] Batch saved %d search results to database (inserted %d new)", len(suggestions), created)
		}
	}

	// Fetch posters synchronously for first batch (visible results)
	// Then async for all remaining results
	if s.tmdbService != nil && len(suggestions) > 0 {
		posterStart := time.Now()
		syncFetchCount := 20
		if len(suggestions) < syncFetchCount {
			syncFetchCount = len(suggestions)
		}

		// Synchronous fetch for visible results (first page)
		fetched := 0
		for i := 0; i < syncFetchCount; i++ {
			if suggestions[i].PosterURL != "" {
				continue
			}

			posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(suggestions[i].Title)
			if err == nil && posterURL != "" {
				// Update in-memory so response includes it
				suggestions[i].PosterURL = posterURL
				suggestions[i].TMDbID = tmdbID
				// Update database for caching
				s.db.GetDB().Model(&models.DownloadSuggestion{}).Where("id = ?", suggestions[i].ID).Updates(map[string]interface{}{
					"poster_url": posterURL,
					"tmdb_id":    tmdbID,
				})
				fetched++
			}
		}

		// Async fetch for ALL remaining results (including pagination)
		go func(suggestionsToFetch []models.DownloadSuggestion) {
			asyncFetched := 0
			asyncFailed := 0

			for i := range suggestionsToFetch {
				if suggestionsToFetch[i].PosterURL != "" {
					continue
				}

				posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(suggestionsToFetch[i].Title)
				if err == nil && posterURL != "" {
					// Update database for caching (fire and forget)
					s.db.GetDB().Model(&models.DownloadSuggestion{}).Where("id = ?", suggestionsToFetch[i].ID).Updates(map[string]interface{}{
						"poster_url": posterURL,
						"tmdb_id":    tmdbID,
					})
					asyncFetched++
				} else {
					asyncFailed++
				}
			}
			log.Printf("[SEARCH] Background poster fetch complete: %d sync + %d async fetched, %d failed, %d total (%.2fs)",
				fetched, asyncFetched, asyncFailed, len(suggestionsToFetch), time.Since(posterStart).Seconds())
		}(suggestions)
	}

	searchDuration := time.Since(searchStart)
	s.logSearchActivity(SearchActivity{
		Query:           query,
		Timestamp:       time.Now(),
		Duration:        searchDuration.Seconds(),
		ResultCount:     len(results),
		SavedCount:      created,
		Providers:       providerNames,
		ProviderTimings: searchResult.Timings,
	})

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results":          suggestions,
		"total":            totalResults,
		"query":            query,
		"saved":            created,
		"providers":        providerNames,
		"provider_timings": searchResult.Timings,
		"cached":           len(results) > 0 && searchDuration.Seconds() < 0.05,
	})
}

func (s *HTTPServer) handleSearchActivity(w http.ResponseWriter, r *http.Request) {
	activities := s.getSearchActivities()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"activities": activities,
	})
}

func (s *HTTPServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	// Read recent log lines from service log
	logPath := "tmp/service.log"
	if envPath := os.Getenv("SERVICE_LOG_PATH"); envPath != "" {
		logPath = envPath
	}

	lines := 50
	if n := r.URL.Query().Get("lines"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			lines = parsed
		}
	}

	filter := r.URL.Query().Get("filter")

	data, err := os.ReadFile(logPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs": []string{},
		})
		return
	}

	// Split into lines and get last N
	allLines := strings.Split(string(data), "\n")
	var filtered []string
	for _, line := range allLines {
		if line == "" {
			continue
		}
		if filter != "" {
			if strings.Contains(line, filter) || strings.Contains(line, "[SEARCH]") || strings.Contains(line, "[TMDB]") || strings.Contains(line, "[TORRENT]") || strings.Contains(line, "[Jackett]") {
				filtered = append(filtered, line)
			}
		} else {
			filtered = append(filtered, line)
		}
	}

	// Get last N lines
	start := len(filtered) - lines
	if start < 0 {
		start = 0
	}
	result := filtered[start:]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  result,
		"total": len(filtered),
	})
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// extractTorrentPoster attempts to extract cover art from a torrent magnet link.
// It saves the first found image to the web images directory and returns the public URL.
func (s *HTTPServer) extractTorrentPoster(magnetLink string) (string, error) {
	if s.downloadManager == nil || magnetLink == "" {
		return "", fmt.Errorf("no download manager or magnet link")
	}

	log.Printf("[TORRENT] Attempting to extract cover art from torrent")

	images, err := s.downloadManager.ExtractImagesFromTorrent(magnetLink)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}
	if len(images) == 0 {
		return "", fmt.Errorf("no images found in torrent")
	}

	// Use the first image found as the poster
	sourcePath := images[0]

	// Generate a unique filename based on magnet link hash
	hash := sha1.Sum([]byte(magnetLink))
	filename := hex.EncodeToString(hash[:]) + filepath.Ext(sourcePath)

	// Determine web images directory
	imagesRoot := os.Getenv("IMAGES_ROOT")
	if imagesRoot == "" {
		imagesRoot = "/home/josh/media-manager/web/images"
	}

	// Ensure directory exists
	if err := os.MkdirAll(imagesRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	destPath := filepath.Join(imagesRoot, filename)

	// Copy the extracted image to the web-accessible directory
	if err := copyFile(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}

	posterURL := "/images/" + filename
	log.Printf("[TORRENT] Saved torrent cover art to %s", posterURL)

	return posterURL, nil
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

	// Fallback: try extracting cover art from torrent
	if suggestion.PosterURL == "" && suggestion.MagnetLink != "" {
		posterURL, err := s.extractTorrentPoster(suggestion.MagnetLink)
		if err == nil && posterURL != "" {
			suggestion.PosterURL = posterURL
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

func (s *HTTPServer) handleExtractTorrentImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MagnetLink string `json:"magnet_link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MagnetLink == "" {
		http.Error(w, "Magnet link required", http.StatusBadRequest)
		return
	}

	log.Printf("[HTTP] Extracting images from torrent: %s", req.MagnetLink)

	// Check if we have a native torrent client that supports image extraction
	images, err := s.downloadManager.ExtractImagesFromTorrent(req.MagnetLink)
	if err != nil {
		log.Printf("[HTTP] Failed to extract images: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return 200 with error info
		json.NewEncoder(w).Encode(map[string]interface{}{
			"images":    []string{},
			"count":     0,
			"error":     err.Error(),
			"supported": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images":    images,
		"count":     len(images),
		"supported": true,
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

// handleCreateSource creates a new download source
func (s *HTTPServer) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var source models.DownloadSource
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if source.Name == "" || source.Type == "" {
		http.Error(w, "Name and type are required", http.StatusBadRequest)
		return
	}

	if err := s.db.GetDB().Create(&source).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to create source: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(source)
}

// handleUpdateSource updates an existing download source
func (s *HTTPServer) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var source models.DownloadSource
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if source.ID == 0 {
		http.Error(w, "Source ID is required", http.StatusBadRequest)
		return
	}

	if err := s.db.GetDB().Model(&models.DownloadSource{}).Where("id = ?", source.ID).Updates(map[string]interface{}{
		"name":     source.Name,
		"type":     source.Type,
		"url":      source.URL,
		"api_key":  source.APIKey,
		"enabled":  source.Enabled,
		"priority": source.Priority,
	}).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to update source: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDeleteSource deletes a download source
func (s *HTTPServer) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == 0 {
		http.Error(w, "Source ID is required", http.StatusBadRequest)
		return
	}

	if err := s.db.GetDB().Delete(&models.DownloadSource{}, req.ID).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete source: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSourceDetail handles requests to /api/sources/{id}/*
func (s *HTTPServer) handleSourceDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sources/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	sourceID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	action := parts[1]

	switch action {
	case "jackett-indexers":
		s.handleJackettIndexers(w, r, uint(sourceID))
	default:
		http.Error(w, "Unknown action", http.StatusNotFound)
	}
}

// handleJackettIndexers returns the list of configured indexers from a Jackett instance
// Uses a dummy search query since the /indexers endpoint requires admin cookies
func (s *HTTPServer) handleJackettIndexers(w http.ResponseWriter, r *http.Request, sourceID uint) {
	// Get the source
	var source models.DownloadSource
	if err := s.db.GetDB().First(&source, sourceID).Error; err != nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	if source.Type != "jackett" {
		http.Error(w, "Source is not a Jackett instance", http.StatusBadRequest)
		return
	}

	if source.URL == "" || source.APIKey == "" {
		http.Error(w, "Jackett URL or API key not configured", http.StatusBadRequest)
		return
	}

	// Build the Jackett search API URL with a dummy query
	// The search response includes an Indexers array with status info
	jackettURL := strings.TrimSuffix(source.URL, "/")
	searchURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results?apikey=%s&Query=test", jackettURL, source.APIKey)

	// Make request to Jackett with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		log.Printf("[HTTP] Failed to fetch Jackett indexers: %v", err)
		http.Error(w, fmt.Sprintf("Failed to fetch indexers from Jackett: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[HTTP] Jackett returned status %d: %s", resp.StatusCode, string(body))
		http.Error(w, fmt.Sprintf("Jackett returned status %d", resp.StatusCode), http.StatusBadGateway)
		return
	}

	// Parse Jackett search response - it includes an Indexers array
	var jackettResponse struct {
		Indexers []struct {
			ID          string `json:"ID"`
			Name        string `json:"Name"`
			Status      int    `json:"Status"`
			Results     int    `json:"Results"`
			Error       string `json:"Error"`
			ElapsedTime int    `json:"ElapsedTime"`
		} `json:"Indexers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jackettResponse); err != nil {
		log.Printf("[HTTP] Failed to parse Jackett response: %v", err)
		http.Error(w, "Failed to parse Jackett response", http.StatusInternalServerError)
		return
	}

	// Format response
	type IndexerInfo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		Results     int    `json:"results"`
		Error       string `json:"error,omitempty"`
		ElapsedTime int    `json:"elapsed_ms"`
	}

	indexers := make([]IndexerInfo, 0, len(jackettResponse.Indexers))
	for _, idx := range jackettResponse.Indexers {
		status := "unknown"
		switch idx.Status {
		case 0:
			status = "disabled"
		case 1:
			status = "enabled"
		case 2:
			status = "error"
		}
		indexers = append(indexers, IndexerInfo{
			ID:          idx.ID,
			Name:        idx.Name,
			Status:      status,
			Results:     idx.Results,
			Error:       idx.Error,
			ElapsedTime: idx.ElapsedTime,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source":   source.Name,
		"indexers": indexers,
	})
}

// handleAllJackettIndexers returns indexers from all configured Jackett sources
// Reads indexers from Jackett's config directory to avoid slow API calls
func (s *HTTPServer) handleAllJackettIndexers(w http.ResponseWriter, r *http.Request) {
	// Get all Jackett sources
	var sources []models.DownloadSource
	if err := s.db.GetDB().Where("type = ? AND enabled = ?", "jackett", true).Find(&sources).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch Jackett sources: %v", err), http.StatusInternalServerError)
		return
	}

	type IndexerInfo struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	type SourceIndexers struct {
		Source   string        `json:"source"`
		URL      string        `json:"url"`
		Indexers []IndexerInfo `json:"indexers"`
	}

	allSources := make([]SourceIndexers, 0, len(sources))

	for _, source := range sources {
		if source.URL == "" || source.APIKey == "" {
			continue
		}

		// Read configured indexers from Jackett's Indexers directory
		// The indexer ID is the filename without .json extension
		indexersDir := filepath.Join(os.Getenv("HOME"), ".config", "Jackett", "Indexers")
		entries, err := os.ReadDir(indexersDir)
		if err != nil {
			log.Printf("[HTTP] Failed to read Jackett indexers directory: %v", err)
			continue
		}

		indexers := make([]IndexerInfo, 0)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			// Skip backup files
			if strings.HasSuffix(entry.Name(), ".json.bak") {
				continue
			}
			id := strings.TrimSuffix(entry.Name(), ".json")
			// Convert kebab-case to readable name
			name := strings.ReplaceAll(id, "-", " ")
			name = strings.Title(name)
			indexers = append(indexers, IndexerInfo{
				ID:     id,
				Name:   name,
				Status: "enabled",
			})
		}

		allSources = append(allSources, SourceIndexers{
			Source:   source.Name,
			URL:      source.URL,
			Indexers: indexers,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sources": allSources,
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

// handleVPNConnect attempts to connect the VPN
func (s *HTTPServer) handleVPNConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("[VPN] Connect requested via API")

	if _, err := exec.LookPath("nordvpn"); err != nil {
		http.Error(w, "NordVPN not installed", http.StatusServiceUnavailable)
		return
	}

	exec.Command("/etc/init.d/nordvpn", "start").Start()
	time.Sleep(2 * time.Second)

	cmd := exec.Command("nordvpn", "connect")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[VPN] Connect failed: %v, output: %s", err, string(output))
		http.Error(w, fmt.Sprintf("Failed to connect: %s", string(output)), http.StatusInternalServerError)
		return
	}

	log.Printf("[VPN] Connect output: %s", string(output))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "connecting", "message": string(output)})
}

// handleVPNDisconnect disconnects the VPN
func (s *HTTPServer) handleVPNDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("[VPN] Disconnect requested via API")

	if _, err := exec.LookPath("nordvpn"); err != nil {
		http.Error(w, "NordVPN not installed", http.StatusServiceUnavailable)
		return
	}

	cmd := exec.Command("nordvpn", "disconnect")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[VPN] Disconnect failed: %v, output: %s", err, string(output))
		http.Error(w, fmt.Sprintf("Failed to disconnect: %s", string(output)), http.StatusInternalServerError)
		return
	}

	log.Printf("[VPN] Disconnect output: %s", string(output))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected", "message": string(output)})
}

// commonLocalPosterNames lists filenames commonly used for local poster art
var commonLocalPosterNames = []string{
	"poster.jpg", "poster.png", "poster.jpeg", "poster.webp",
	"folder.jpg", "folder.png", "folder.jpeg", "folder.webp",
	"cover.jpg", "cover.png", "cover.jpeg", "cover.webp",
	"movie.jpg", "movie.png", "movie.jpeg", "movie.webp",
}

// findLocalPoster searches a movie directory for local poster image files (case-insensitive)
func findLocalPoster(moviePath string) string {
	dir := filepath.Dir(moviePath)

	// First try exact match
	for _, name := range commonLocalPosterNames {
		posterPath := filepath.Join(dir, name)
		if info, err := os.Stat(posterPath); err == nil && !info.IsDir() {
			return posterPath
		}
	}

	// Case-insensitive fallback: list directory and check case-insensitively
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		nameLower := strings.ToLower(entry.Name())
		for _, posterName := range commonLocalPosterNames {
			if nameLower == posterName {
				return filepath.Join(dir, entry.Name())
			}
		}
	}

	return ""
}

// handlePosterFile serves a local poster image file
func (s *HTTPServer) handlePosterFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	posterPath := r.URL.Query().Get("path")
	if posterPath == "" {
		http.Error(w, "Path parameter required", http.StatusBadRequest)
		return
	}

	// Security check: ensure the path is within the media directory
	mediaDir := "/mnt/media"
	if envDir := os.Getenv("MEDIA_DIR"); envDir != "" {
		mediaDir = envDir
	}

	absPath, err := filepath.Abs(posterPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absMediaDir, err := filepath.Abs(mediaDir)
	if err != nil {
		http.Error(w, "Invalid media directory", http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(absPath, absMediaDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists
	if info, err := os.Stat(absPath); err != nil || info.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, absPath)
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
				} else {
					// Jellyfin has no poster - try local file, then DB cache, then TMDb
					if item.Path != "" {
						movieDir := filepath.Dir(item.Path)
						// Map Jellyfin's Docker path to actual path if needed
						movieDir = strings.Replace(movieDir, "/media/movies", "/mnt/media/Movies", 1)
						if localPoster := findLocalPoster(filepath.Join(movieDir, "movie.mkv")); localPoster != "" {
							posterURL = "/api/library/poster-file?path=" + url.QueryEscape(localPoster)
						}
					}
					// If still no poster, try DB cache
					if posterURL == "/images/placeholder-movie.jpg" && s.tmdbService != nil {
						cleanName, year := s.tmdbService.ExtractMovieInfo(item.Name)
						cleanTitle := s.tmdbService.CleanTitle(cleanName)
						if cleanTitle != "" {
							var metadata models.MovieMetadata
							if err := s.db.GetDB().Where("clean_title = ?", cleanTitle).First(&metadata).Error; err == nil && metadata.PosterURL != "" {
								posterURL = metadata.PosterURL
							} else {
								// Not in cache, try fetching from TMDb
								if fetchedURL, _, err := s.tmdbService.FetchPosterForTask(item.Name); err == nil && fetchedURL != "" {
									posterURL = fetchedURL
								}
							}
							_ = year
						}
					}
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

						// Check for local poster file first
						if localPoster := findLocalPoster(path); localPoster != "" {
							posterURL = "/api/library/poster-file?path=" + url.QueryEscape(localPoster)
						} else if s.tmdbService != nil {
							// Fall back to DB cache, then fetch from TMDb
							cleanMovieTitle := s.tmdbService.CleanTitle(movieName)

							if cleanMovieTitle != "" {
								var metadata models.MovieMetadata
								if err := s.db.GetDB().Where("clean_title = ?", cleanMovieTitle).
									First(&metadata).Error; err == nil && metadata.PosterURL != "" {
									posterURL = metadata.PosterURL
								} else {
									// Not in cache, fetch from TMDb
									if fetchedURL, _, err := s.tmdbService.FetchPosterForTask(title); err == nil && fetchedURL != "" {
										posterURL = fetchedURL
									}
								}
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.URL.Query().Get("title")
	moviePath := r.URL.Query().Get("path")

	if title == "" && moviePath == "" {
		http.Error(w, "Title or path parameter required", http.StatusBadRequest)
		return
	}

	// If path provided, check for local poster first
	if moviePath != "" {
		if localPoster := findLocalPoster(moviePath); localPoster != "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"poster_url": "/api/library/poster-file?path=" + url.QueryEscape(localPoster),
			})
			return
		}
		// Derive title from path if not provided
		if title == "" {
			title = filepath.Base(filepath.Dir(moviePath))
			if title == "." || title == "Movies" {
				title = strings.TrimSuffix(filepath.Base(moviePath), filepath.Ext(moviePath))
			}
		}
	}

	if s.tmdbService == nil {
		http.Error(w, "TMDb service not available", http.StatusServiceUnavailable)
		return
	}

	posterURL, _, err := s.tmdbService.FetchPosterForTask(title)
	if err != nil {
		log.Printf("[HTTP] Failed to fetch poster for '%s': %v", title, err)
		http.Error(w, "Failed to fetch poster", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"poster_url": posterURL,
	})
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
		log.Printf("[HTTP] Failed to fetch poster for '%s': %v", title, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"poster_url": "",
		})
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
	mediaDir := "/mnt/media/Movies"
	if envDir := os.Getenv("MEDIA_DIR"); envDir != "" {
		mediaDir = envDir
	}

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
		// Skip if local poster already exists
		if localPoster := findLocalPoster(media.Path); localPoster != "" {
			cachedCount++
			continue
		}

		// Extract clean movie name and year using the same logic as TMDb service
		movieName, _ := s.tmdbService.ExtractMovieInfo(media.Title)
		cleanTitle := s.tmdbService.CleanTitle(movieName)

		// Check if already cached in DB
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
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID   uint   `json:"id"`
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var moviePath string

	if req.Path != "" {
		moviePath = req.Path
	} else if req.ID != 0 {
		// Look up media file by ID
		var mediaFile models.MediaFile
		if err := s.db.GetDB().First(&mediaFile, req.ID).Error; err != nil {
			log.Printf("[HTTP] Movie not found with ID %d: %v", req.ID, err)
			http.Error(w, "Movie not found", http.StatusNotFound)
			return
		}
		moviePath = mediaFile.Path
	} else {
		http.Error(w, "Missing path or id parameter", http.StatusBadRequest)
		return
	}

	// Validate the path is within the media directory to prevent directory traversal
	mediaDir := "/mnt/media"
	if !strings.HasPrefix(moviePath, mediaDir) {
		log.Printf("[HTTP] Invalid movie path (outside media dir): %s", moviePath)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Check if path exists
	info, err := os.Stat(moviePath)
	if err != nil {
		log.Printf("[HTTP] Movie path not found: %s", moviePath)
		http.Error(w, "Movie not found on disk", http.StatusNotFound)
		return
	}

	// Delete the file or directory
	if info.IsDir() {
		err = os.RemoveAll(moviePath)
	} else {
		err = os.Remove(moviePath)
		// Also try to remove empty parent directory
		parentDir := filepath.Dir(moviePath)
		if remaining, _ := os.ReadDir(parentDir); len(remaining) == 0 {
			os.Remove(parentDir)
		}
	}

	if err != nil {
		log.Printf("[HTTP] Failed to delete movie at %s: %v", moviePath, err)
		http.Error(w, fmt.Sprintf("Failed to delete movie: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove from database if it exists
	if err := s.db.GetDB().Where("path = ?", moviePath).Delete(&models.MediaFile{}).Error; err != nil {
		log.Printf("[HTTP] Failed to delete media file from database: %v", err)
	}

	log.Printf("[HTTP] Deleted movie: %s", moviePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    moviePath,
		"message": "Movie deleted successfully",
	})
}

func (s *HTTPServer) handleTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.db.GetTags()
	if err != nil {
		log.Printf("[HTTP] Failed to fetch tags: %v", err)
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags": tags,
	})
}

func (s *HTTPServer) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	tag := &models.Tag{
		Name:  req.Name,
		Color: req.Color,
	}

	if err := s.db.CreateTag(tag); err != nil {
		log.Printf("[HTTP] Failed to create tag: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create tag: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Created tag: %s", req.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tag":     tag,
		"message": "Tag created successfully",
	})
}

func (s *HTTPServer) handleAssignTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path  string `json:"path"`
		TagID uint   `json:"tag_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Path == "" || req.TagID == 0 {
		http.Error(w, "Path and tag_id are required", http.StatusBadRequest)
		return
	}

	if err := s.db.AssignTagToMediaFile(req.Path, req.TagID); err != nil {
		log.Printf("[HTTP] Failed to assign tag %d to %s: %v", req.TagID, req.Path, err)
		http.Error(w, fmt.Sprintf("Failed to assign tag: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Assigned tag %d to %s", req.TagID, req.Path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tag assigned successfully",
	})
}

func (s *HTTPServer) handleRemoveTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path  string `json:"path"`
		TagID uint   `json:"tag_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Path == "" || req.TagID == 0 {
		http.Error(w, "Path and tag_id are required", http.StatusBadRequest)
		return
	}

	if err := s.db.RemoveTagFromMediaFile(req.Path, req.TagID); err != nil {
		log.Printf("[HTTP] Failed to remove tag %d from %s: %v", req.TagID, req.Path, err)
		http.Error(w, fmt.Sprintf("Failed to remove tag: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[HTTP] Removed tag %d from %s", req.TagID, req.Path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tag removed successfully",
	})
}

// handleHealth returns the health status of the server
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// RSS Feed handlers
func (s *HTTPServer) handleRSSFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var feeds []models.RSSFeed
	if err := s.db.GetDB().Find(&feeds).Error; err != nil {
		log.Printf("[HTTP] Failed to fetch RSS feeds: %v", err)
		jsonError(w, "Failed to fetch RSS feeds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feeds": feeds,
	})
}

func (s *HTTPServer) handleCreateRSSFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var feed models.RSSFeed
	if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if feed.Name == "" || feed.URL == "" {
		http.Error(w, "Name and URL are required", http.StatusBadRequest)
		return
	}

	if err := s.db.GetDB().Create(&feed).Error; err != nil {
		log.Printf("[HTTP] Failed to create RSS feed: %v", err)
		jsonError(w, "Failed to create RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"feed":    feed,
	})
}

func (s *HTTPServer) handleUpdateRSSFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var feed models.RSSFeed
	if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.db.GetDB().Save(&feed).Error; err != nil {
		log.Printf("[HTTP] Failed to update RSS feed: %v", err)
		jsonError(w, "Failed to update RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"feed":    feed,
	})
}

func (s *HTTPServer) handleDeleteRSSFeed(w http.ResponseWriter, r *http.Request) {
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

	if err := s.db.GetDB().Delete(&models.RSSFeed{}, req.ID).Error; err != nil {
		log.Printf("[HTTP] Failed to delete RSS feed: %v", err)
		jsonError(w, "Failed to delete RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "RSS feed deleted",
	})
}

func (s *HTTPServer) handleCheckRSSFeed(w http.ResponseWriter, r *http.Request) {
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

	var feed models.RSSFeed
	if err := s.db.GetDB().First(&feed, req.ID).Error; err != nil {
		jsonError(w, "Feed not found", http.StatusNotFound)
		return
	}

	// TODO: Trigger immediate check
	feed.LastCheck = time.Now()
	s.db.GetDB().Save(&feed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Feed check triggered",
	})
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
