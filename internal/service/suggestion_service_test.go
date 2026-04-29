package service

import (
	"testing"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

func TestSuggestionService_CreateAndList(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestion := &models.DownloadSuggestion{
		SourceID:   source.ID,
		Title:      "Test Movie 1080p",
		InfoHash:   "abc123",
		MagnetLink: "magnet:?xt=urn:btih:abc123",
		Size:       5000000000,
		Seeders:    100,
		Leechers:   10,
		Status:     "pending",
	}

	if err := database.GetDB().Create(suggestion).Error; err != nil {
		t.Fatalf("Failed to create suggestion: %v", err)
	}

	suggestions, total, err := suggestionService.ListSuggestions("pending", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list suggestions: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 suggestion, got %d", total)
	}

	if len(suggestions) != 1 {
		t.Errorf("Expected 1 suggestion in list, got %d", len(suggestions))
	}

	if suggestions[0].Title != "Test Movie 1080p" {
		t.Errorf("Expected title 'Test Movie 1080p', got '%s'", suggestions[0].Title)
	}
}

func TestSuggestionService_ApproveSuggestion(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestion := &models.DownloadSuggestion{
		SourceID:   source.ID,
		Title:      "Test Movie 1080p",
		InfoHash:   "def456",
		MagnetLink: "magnet:?xt=urn:btih:def456",
		Size:       5000000000,
		Seeders:    100,
		Leechers:   10,
		Status:     "pending",
	}

	if err := database.GetDB().Create(suggestion).Error; err != nil {
		t.Fatalf("Failed to create suggestion: %v", err)
	}

	if err := suggestionService.ApproveSuggestion(suggestion.ID, "Looks good", false); err != nil {
		t.Fatalf("Failed to approve suggestion: %v", err)
	}

	var updated models.DownloadSuggestion
	if err := database.GetDB().First(&updated, suggestion.ID).Error; err != nil {
		t.Fatalf("Failed to get updated suggestion: %v", err)
	}

	if updated.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", updated.Status)
	}

	if updated.ApprovedAt == nil {
		t.Error("Expected ApprovedAt to be set")
	}

	if updated.Notes != "Looks good" {
		t.Errorf("Expected notes 'Looks good', got '%s'", updated.Notes)
	}

	var task models.DownloadTask
	if err := database.GetDB().Where("info_hash = ?", "def456").First(&task).Error; err != nil {
		t.Fatalf("Failed to find download task: %v", err)
	}

	if task.Title != "Test Movie 1080p" {
		t.Errorf("Expected task title 'Test Movie 1080p', got '%s'", task.Title)
	}
}

func TestSuggestionService_RejectSuggestion(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestion := &models.DownloadSuggestion{
		SourceID:   source.ID,
		Title:      "Bad Movie 1080p",
		InfoHash:   "ghi789",
		MagnetLink: "magnet:?xt=urn:btih:ghi789",
		Size:       5000000000,
		Seeders:    100,
		Leechers:   10,
		Status:     "pending",
	}

	if err := database.GetDB().Create(suggestion).Error; err != nil {
		t.Fatalf("Failed to create suggestion: %v", err)
	}

	if err := suggestionService.RejectSuggestion(suggestion.ID, "Not interested"); err != nil {
		t.Fatalf("Failed to reject suggestion: %v", err)
	}

	var updated models.DownloadSuggestion
	if err := database.GetDB().First(&updated, suggestion.ID).Error; err != nil {
		t.Fatalf("Failed to get updated suggestion: %v", err)
	}

	if updated.Status != "rejected" {
		t.Errorf("Expected status 'rejected', got '%s'", updated.Status)
	}

	if updated.RejectedAt == nil {
		t.Error("Expected RejectedAt to be set")
	}
}

func TestSuggestionService_GetStats(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestions := []*models.DownloadSuggestion{
		{SourceID: source.ID, Title: "Movie 1", InfoHash: "hash1", MagnetLink: "magnet1", Size: 5000000000, Seeders: 100, Status: "pending"},
		{SourceID: source.ID, Title: "Movie 2", InfoHash: "hash2", MagnetLink: "magnet2", Size: 5000000000, Seeders: 100, Status: "pending"},
		{SourceID: source.ID, Title: "Movie 3", InfoHash: "hash3", MagnetLink: "magnet3", Size: 5000000000, Seeders: 100, Status: "approved"},
		{SourceID: source.ID, Title: "Movie 4", InfoHash: "hash4", MagnetLink: "magnet4", Size: 5000000000, Seeders: 100, Status: "rejected"},
	}

	for _, s := range suggestions {
		if err := database.GetDB().Create(s).Error; err != nil {
			t.Fatalf("Failed to create suggestion: %v", err)
		}
	}

	stats, err := suggestionService.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["pending"] != 2 {
		t.Errorf("Expected 2 pending, got %d", stats["pending"])
	}

	if stats["approved"] != 1 {
		t.Errorf("Expected 1 approved, got %d", stats["approved"])
	}

	if stats["rejected"] != 1 {
		t.Errorf("Expected 1 rejected, got %d", stats["rejected"])
	}

	if stats["total"] != 4 {
		t.Errorf("Expected 4 total, got %d", stats["total"])
	}
}

func TestSearchSuggestions(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestions := []*models.DownloadSuggestion{
		{SourceID: source.ID, Title: "Inception 2010 1080p", InfoHash: "hash1", MagnetLink: "magnet1", Size: 10000000000, Seeders: 150, Status: "pending"},
		{SourceID: source.ID, Title: "Interstellar 2014 720p", InfoHash: "hash2", MagnetLink: "magnet2", Size: 8000000000, Seeders: 80, Status: "pending"},
		{SourceID: source.ID, Title: "The Matrix 1999 1080p", InfoHash: "hash3", MagnetLink: "magnet3", Size: 12000000000, Seeders: 200, Status: "pending"},
		{SourceID: source.ID, Title: "Inception 2010 4K", InfoHash: "hash4", MagnetLink: "magnet4", Size: 30000000000, Seeders: 50, Status: "approved"},
	}

	for _, s := range suggestions {
		if err := database.GetDB().Create(s).Error; err != nil {
			t.Fatalf("Failed to create suggestion: %v", err)
		}
	}

	t.Run("search by title query", func(t *testing.T) {
		results, total, err := suggestionService.SearchSuggestions("Inception", "", "seeders", 0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected 2 results, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results in list, got %d", len(results))
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		_, total, err := suggestionService.SearchSuggestions("", "pending", "seeders", 0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if total != 3 {
			t.Errorf("Expected 3 pending results, got %d", total)
		}
	})

	t.Run("filter by min seeders", func(t *testing.T) {
		_, total, err := suggestionService.SearchSuggestions("", "", "seeders", 100, 10, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected 2 results with >=100 seeders, got %d", total)
		}
	})

	t.Run("sort by seeders descending", func(t *testing.T) {
		results, _, err := suggestionService.SearchSuggestions("", "", "seeders", 0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if len(results) > 1 && results[0].Seeders < results[1].Seeders {
			t.Error("Results not sorted by seeders descending")
		}
	})

	t.Run("sort by size ascending", func(t *testing.T) {
		results, _, err := suggestionService.SearchSuggestions("", "", "size", 0, 10, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if len(results) > 1 && results[0].Size > results[1].Size {
			t.Error("Results not sorted by size ascending")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := suggestionService.SearchSuggestions("", "", "seeders", 0, 2, 0)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		if total != 4 {
			t.Errorf("Expected total 4, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results in page, got %d", len(results))
		}

		results, _, err = suggestionService.SearchSuggestions("", "", "seeders", 0, 2, 2)
		if err != nil {
			t.Fatalf("Failed to search page 2: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results in page 2, got %d", len(results))
		}
	})
}

func TestCalculateQualityScore(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	suggestionService := NewSuggestionService(database, nil, nil)

	t.Run("high seeders", func(t *testing.T) {
		suggestion := &models.DownloadSuggestion{
			Seeders: 1000,
			Size:    10 * 1024 * 1024 * 1024,
			Quality: "BluRay",
		}
		score := suggestionService.CalculateQualityScore(suggestion)
		if score < 6.0 {
			t.Errorf("Expected high score for high seeders, got %.2f", score)
		}
	})

	t.Run("low seeders", func(t *testing.T) {
		suggestion := &models.DownloadSuggestion{
			Seeders: 5,
			Size:    10 * 1024 * 1024 * 1024,
		}
		score := suggestionService.CalculateQualityScore(suggestion)
		if score > 4.0 {
			t.Errorf("Expected low score for low seeders, got %.2f", score)
		}
	})

	t.Run("optimal size", func(t *testing.T) {
		suggestion := &models.DownloadSuggestion{
			Seeders: 100,
			Size:    10 * 1024 * 1024 * 1024,
			Quality: "WEB-DL",
		}
		score := suggestionService.CalculateQualityScore(suggestion)
		if score < 3.0 || score > 5.0 {
			t.Errorf("Expected score between 3-5 for this configuration, got %.2f", score)
		}
	})

	t.Run("oversized", func(t *testing.T) {
		suggestion := &models.DownloadSuggestion{
			Seeders: 100,
			Size:    50 * 1024 * 1024 * 1024,
		}
		score := suggestionService.CalculateQualityScore(suggestion)
		if score > 6.0 {
			t.Errorf("Expected lower score for oversized, got %.2f", score)
		}
	})
}

func TestGetTopRecommendations(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	source := &models.DownloadSource{
		Name:    "Test Source",
		Type:    "jackett",
		URL:     "http://localhost",
		Enabled: true,
	}
	if err := database.GetDB().Create(source).Error; err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	suggestionService := NewSuggestionService(database, nil, nil)

	suggestions := []*models.DownloadSuggestion{
		{SourceID: source.ID, Title: "High Seeders Pending", InfoHash: "hash1", MagnetLink: "magnet1", Size: 10000000000, Seeders: 200, Status: "pending"},
		{SourceID: source.ID, Title: "Medium Seeders Pending", InfoHash: "hash2", MagnetLink: "magnet2", Size: 10000000000, Seeders: 50, Status: "pending"},
		{SourceID: source.ID, Title: "Low Seeders Pending", InfoHash: "hash3", MagnetLink: "magnet3", Size: 10000000000, Seeders: 5, Status: "pending"},
		{SourceID: source.ID, Title: "Approved", InfoHash: "hash4", MagnetLink: "magnet4", Size: 10000000000, Seeders: 300, Status: "approved"},
	}

	for _, s := range suggestions {
		if err := database.GetDB().Create(s).Error; err != nil {
			t.Fatalf("Failed to create suggestion: %v", err)
		}
	}

	t.Run("returns only pending with min seeders", func(t *testing.T) {
		results, err := suggestionService.GetTopRecommendations(5)
		if err != nil {
			t.Fatalf("Failed to get recommendations: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 recommendations (pending with >=10 seeders), got %d", len(results))
		}
		for _, r := range results {
			if r.Status != "pending" {
				t.Errorf("Expected pending status, got %s", r.Status)
			}
			if r.Seeders < 10 {
				t.Errorf("Expected seeders >= 10, got %d", r.Seeders)
			}
		}
	})

	t.Run("sorted by seeders descending", func(t *testing.T) {
		results, err := suggestionService.GetTopRecommendations(5)
		if err != nil {
			t.Fatalf("Failed to get recommendations: %v", err)
		}
		if len(results) > 1 && results[0].Seeders < results[1].Seeders {
			t.Error("Results not sorted by seeders descending")
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		results, err := suggestionService.GetTopRecommendations(1)
		if err != nil {
			t.Fatalf("Failed to get recommendations: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 recommendation, got %d", len(results))
		}
	})
}
