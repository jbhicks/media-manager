package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

// WatchlistEndpoints handles watchlist-related API endpoints
// Users can save movies and TV shows to watch later
// POST   /api/watchlist              - Add item to watchlist
// GET    /api/watchlist              - Get user's watchlist
// DELETE /api/watchlist/:id          - Remove item from watchlist
// GET    /api/watchlist/check/:type/:tmdb_id - Check if item is in watchlist

type WatchlistEndpoints struct {
	db *db.Database
}

func NewWatchlistEndpoints(database *db.Database) *WatchlistEndpoints {
	return &WatchlistEndpoints{db: database}
}

// RegisterRoutes registers all watchlist endpoints with the HTTP mux
func (w *WatchlistEndpoints) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/watchlist", AuthMiddleware(w.HandleWatchlist))
	mux.HandleFunc("/api/watchlist/", AuthMiddleware(w.HandleWatchlistItem))
	mux.HandleFunc("/api/watchlist/check/", AuthMiddleware(w.CheckWatchlist))
}

// HandleWatchlist handles GET /api/watchlist (list) and POST /api/watchlist (add)
func (w *WatchlistEndpoints) HandleWatchlist(wr http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.GetWatchlist(wr, r)
	case http.MethodPost:
		w.AddToWatchlist(wr, r)
	default:
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
	}
}

// HandleWatchlistItem handles DELETE /api/watchlist/:id
func (w *WatchlistEndpoints) HandleWatchlistItem(wr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
		return
	}
	w.RemoveFromWatchlist(wr, r)
}

// AddToWatchlist adds a movie or TV show to the user's watchlist
func (w *WatchlistEndpoints) AddToWatchlist(wr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user, ok := GetUserFromContext(r)
	if !ok {
		wr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Authentication required"})
		return
	}

	var req struct {
		MediaType string `json:"media_type"`
		TMDbID    int    `json:"tmdb_id"`
		Title     string `json:"title"`
		PosterURL string `json:"poster_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Invalid request body"})
		return
	}

	if req.MediaType == "" || req.TMDbID == 0 || req.Title == "" {
		wr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "media_type, tmdb_id, and title are required"})
		return
	}

	// Check if already in watchlist
	var existing models.Watchlist
	result := w.db.GetDB().Where("user_id = ? AND media_type = ? AND tm_db_id = ?", user.ID, req.MediaType, req.TMDbID).First(&existing)
	if result.Error == nil {
		wr.WriteHeader(http.StatusConflict)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Item already in watchlist"})
		return
	}

	// Add to watchlist
	item := models.Watchlist{
		UserID:    user.ID,
		MediaType: req.MediaType,
		TMDbID:    req.TMDbID,
		Title:     req.Title,
		PosterURL: req.PosterURL,
		AddedAt:   time.Now(),
	}

	if err := w.db.GetDB().Create(&item).Error; err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Failed to add to watchlist"})
		return
	}

	wr.WriteHeader(http.StatusCreated)
	json.NewEncoder(wr).Encode(item)
}

// GetWatchlist returns the user's watchlist
func (w *WatchlistEndpoints) GetWatchlist(wr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user, ok := GetUserFromContext(r)
	if !ok {
		wr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Authentication required"})
		return
	}

	var items []models.Watchlist
	if err := w.db.GetDB().Where("user_id = ?", user.ID).Order("added_at DESC").Find(&items).Error; err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Failed to fetch watchlist"})
		return
	}

	json.NewEncoder(wr).Encode(items)
}

// RemoveFromWatchlist removes an item from the user's watchlist
func (w *WatchlistEndpoints) RemoveFromWatchlist(wr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user, ok := GetUserFromContext(r)
	if !ok {
		wr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Authentication required"})
		return
	}

	// Extract ID from URL path
	path := r.URL.Path
	// Expected path: /api/watchlist/:id
	// We need to extract the ID after /api/watchlist/
	idStr := path[len("/api/watchlist/"):]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Invalid watchlist ID"})
		return
	}

	// Delete the item (only if it belongs to the user)
	result := w.db.GetDB().Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.Watchlist{})
	if result.Error != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Failed to remove from watchlist"})
		return
	}

	if result.RowsAffected == 0 {
		wr.WriteHeader(http.StatusNotFound)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Item not found in watchlist"})
		return
	}

	json.NewEncoder(wr).Encode(map[string]interface{}{"success": true, "message": "Item removed from watchlist"})
}

// CheckWatchlist checks if a specific movie/TV show is in the user's watchlist
func (w *WatchlistEndpoints) CheckWatchlist(wr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		wr.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Method not allowed"})
		return
	}

	// Get user from context (set by auth middleware)
	user, ok := GetUserFromContext(r)
	if !ok {
		wr.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Authentication required"})
		return
	}

	// Extract media_type and tmdb_id from URL path
	// Expected path: /api/watchlist/check/:type/:tmdb_id
	path := r.URL.Path
	log.Printf("[Watchlist] CheckWatchlist called with path: %s", path)
	path = path[len("/api/watchlist/check/"):]
	log.Printf("[Watchlist] Path after prefix removal: %s", path)
	
	// Split remaining path
	parts := strings.Split(path, "/")
	log.Printf("[Watchlist] Path parts: %v", parts)

	if len(parts) != 2 {
		wr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Invalid URL format. Expected: /api/watchlist/check/:type/:tmdb_id"})
		return
	}

	mediaType := parts[0]
	tmdbID, err := strconv.Atoi(parts[1])
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(wr).Encode(map[string]interface{}{"error": true, "message": "Invalid TMDB ID"})
		return
	}

	var item models.Watchlist
	result := w.db.GetDB().Where("user_id = ? AND media_type = ? AND tm_db_id = ?", user.ID, mediaType, tmdbID).First(&item)
	
	isInWatchlist := result.Error == nil
	json.NewEncoder(wr).Encode(map[string]interface{}{
		"in_watchlist": isInWatchlist,
		"item":         item,
	})
}

// GetUserFromContext extracts the user from the request headers
// The AuthMiddleware sets X-User-ID, X-User-Name, and X-User-Role headers
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return nil, false
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, false
	}

	user := &models.User{
		ID:       uint(userID),
		Username: r.Header.Get("X-User-Name"),
		Role:     r.Header.Get("X-User-Role"),
	}
	return user, true
}
