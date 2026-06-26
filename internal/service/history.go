package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
	"gorm.io/gorm"
)

// WatchHistoryEndpoints handles watch history API routes
type WatchHistoryEndpoints struct {
	db *db.Database
}

// NewWatchHistoryEndpoints creates a new watch history handler
func NewWatchHistoryEndpoints(database *db.Database) *WatchHistoryEndpoints {
	return &WatchHistoryEndpoints{db: database}
}

// RegisterRoutes registers watch history routes
func (wh *WatchHistoryEndpoints) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/history", wh.handleHistory)
	mux.HandleFunc("/api/history/progress", wh.handleUpdateProgress)
	mux.HandleFunc("/api/history/complete", wh.handleMarkComplete)
	mux.HandleFunc("/api/history/resume", wh.handleResumePoints)
	mux.HandleFunc("/api/history/stats", wh.handleStats)
}

// getUserID extracts user ID from request context (set by auth middleware)
func getUserID(r *http.Request) (uint, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf("unauthorized")
	}
	
	id, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID")
	}
	
	return uint(id), nil
}

// handleHistory handles GET (list) and DELETE (clear) for watch history
func (wh *WatchHistoryEndpoints) handleHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		wh.listHistory(w, r, userID)
	case http.MethodDelete:
		wh.clearHistory(w, r, userID)
	default:
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listHistory returns the user's watch history
func (wh *WatchHistoryEndpoints) listHistory(w http.ResponseWriter, r *http.Request, userID uint) {
	var history []models.WatchHistory
	
	result := wh.db.GetDB().Where("user_id = ?", userID).
		Order("watched_at DESC").
		Limit(100).
		Find(&history)
	
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to fetch history for user %d: %v", userID, result.Error)
		jsonError(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// clearHistory removes all watch history for a user
func (wh *WatchHistoryEndpoints) clearHistory(w http.ResponseWriter, r *http.Request, userID uint) {
	result := wh.db.GetDB().Where("user_id = ?", userID).Delete(&models.WatchHistory{})
	
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to clear history for user %d: %v", userID, result.Error)
		jsonError(w, "Failed to clear history", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "History cleared",
		"deleted": result.RowsAffected,
	})
}

// handleUpdateProgress updates or creates a watch progress entry
func (wh *WatchHistoryEndpoints) handleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID, err := getUserID(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req struct {
		MediaType string  `json:"media_type"`
		MediaID   uint    `json:"media_id"`
		Position  int     `json:"position"`
		Duration  int     `json:"duration"`
		Progress  float64 `json:"progress"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.MediaType == "" || req.MediaID == 0 {
		jsonError(w, "Media type and ID are required", http.StatusBadRequest)
		return
	}
	
	// Find or create history entry
	var history models.WatchHistory
	result := wh.db.GetDB().Where(
		"user_id = ? AND media_type = ? AND media_id = ?",
		userID, req.MediaType, req.MediaID,
	).First(&history)
	
	now := time.Now()
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new entry
		history = models.WatchHistory{
			UserID:    userID,
			MediaType: req.MediaType,
			MediaID:   req.MediaID,
			Position:  req.Position,
			Duration:  req.Duration,
			Progress:  req.Progress,
			WatchedAt: now,
		}
		result = wh.db.GetDB().Create(&history)
	} else if result.Error == nil {
		// Update existing entry
		history.Position = req.Position
		history.Duration = req.Duration
		history.Progress = req.Progress
		history.WatchedAt = now
		result = wh.db.GetDB().Save(&history)
	}
	
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to update progress for user %d: %v", userID, result.Error)
		jsonError(w, "Failed to update progress", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Progress updated",
		"progress": req.Progress,
		"position": req.Position,
	})
}

// handleMarkComplete marks a media item as fully watched
func (wh *WatchHistoryEndpoints) handleMarkComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID, err := getUserID(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req struct {
		MediaType string `json:"media_type"`
		MediaID   uint   `json:"media_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	result := wh.db.GetDB().Model(&models.WatchHistory{}).
		Where("user_id = ? AND media_type = ? AND media_id = ?", userID, req.MediaType, req.MediaID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"progress":     1.0,
			"watched_at":   time.Now(),
		})
	
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to mark complete for user %d: %v", userID, result.Error)
		jsonError(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Marked as complete",
	})
}

// handleResumePoints returns media items the user can resume watching
func (wh *WatchHistoryEndpoints) handleResumePoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID, err := getUserID(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var history []models.WatchHistory
	result := wh.db.GetDB().Where(
		"user_id = ? AND is_completed = ? AND progress > 0 AND progress < 1",
		userID, false,
	).Order("watched_at DESC").Limit(20).Find(&history)
	
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to fetch resume points for user %d: %v", userID, result.Error)
		jsonError(w, "Failed to fetch resume points", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resume_points": history,
		"count":         len(history),
	})
}

// handleStats returns watch statistics for the user
func (wh *WatchHistoryEndpoints) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID, err := getUserID(r)
	if err != nil {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var totalWatched int64
	var totalCompleted int64
	var totalTime int64
	
	wh.db.GetDB().Model(&models.WatchHistory{}).Where("user_id = ?", userID).Count(&totalWatched)
	wh.db.GetDB().Model(&models.WatchHistory{}).Where("user_id = ? AND is_completed = ?", userID, true).Count(&totalCompleted)
	wh.db.GetDB().Model(&models.WatchHistory{}).Where("user_id = ?", userID).Select("COALESCE(SUM(duration * progress), 0)").Scan(&totalTime)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_watched":   totalWatched,
		"total_completed": totalCompleted,
		"total_time_seconds": totalTime,
		"total_time_hours":   float64(totalTime) / 3600.0,
	})
}
