package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

func TestTorrentIDForTask(t *testing.T) {
	dm := &DownloadManager{}

	task := models.DownloadTask{
		InfoHash: "08ada5a7a6193aae36ec839d9eaa8132e13e7286",
	}
	expected := torrent.TorrentIDFromInfoHash(task.InfoHash)
	if dm.torrentIDForTask(&task) != expected {
		t.Fatalf("expected torrent ID from info hash")
	}

	task.TorrentID = 12345
	if dm.torrentIDForTask(&task) != 12345 {
		t.Fatal("stored torrent ID should take precedence")
	}
}

func TestMagnetForTask(t *testing.T) {
	dm := &DownloadManager{}

	task := models.DownloadTask{MagnetLink: "magnet:one"}
	if dm.magnetForTask(&task) != "magnet:one" {
		t.Fatal("expected magnet link")
	}

	task = models.DownloadTask{TorrentURL: "magnet:two"}
	if dm.magnetForTask(&task) != "magnet:two" {
		t.Fatal("expected torrent URL fallback")
	}
}

func TestFilterResults(t *testing.T) {
	dm := &DownloadManager{}
	now := time.Now()

	tests := []struct {
		name          string
		results       []models.SearchResult
		rule          *models.DownloadRule
		expectedCount int
	}{
		{
			name: "filter by min seeders",
			results: []models.SearchResult{
				{Title: "high_seeders", Seeders: 100},
				{Title: "low_seeders", Seeders: 5},
			},
			rule: &models.DownloadRule{
				MinSeeders: 10,
			},
			expectedCount: 1,
		},
		{
			name: "filter by max seeders",
			results: []models.SearchResult{
				{Title: "high_seeders", Seeders: 1000},
				{Title: "medium_seeders", Seeders: 50},
			},
			rule: &models.DownloadRule{
				MaxSeeders: 100,
			},
			expectedCount: 1,
		},
		{
			name: "filter by size range",
			results: []models.SearchResult{
				{Title: "too_small", Size: 100 * 1024 * 1024},
				{Title: "just_right", Size: 2 * 1024 * 1024 * 1024},
				{Title: "too_large", Size: 20 * 1024 * 1024 * 1024},
			},
			rule: &models.DownloadRule{
				MinSize: 1024 * 1024 * 1024,
				MaxSize: 10 * 1024 * 1024 * 1024,
			},
			expectedCount: 1,
		},
		{
			name: "filter by upload age - max",
			results: []models.SearchResult{
				{Title: "recent", UploadDate: now.AddDate(0, 0, -30)},
				{Title: "old", UploadDate: now.AddDate(0, 0, -180)},
			},
			rule: &models.DownloadRule{
				MaxUploadAge: 90,
			},
			expectedCount: 1,
		},
		{
			name: "filter by upload age - min",
			results: []models.SearchResult{
				{Title: "too_new", UploadDate: now.AddDate(0, 0, -5)},
				{Title: "old_enough", UploadDate: now.AddDate(0, 0, -30)},
			},
			rule: &models.DownloadRule{
				MinUploadAge: 10,
			},
			expectedCount: 1,
		},
		{
			name: "combined filters",
			results: []models.SearchResult{
				{Title: "perfect", Seeders: 50, Size: 5 * 1024 * 1024 * 1024, UploadDate: now.AddDate(0, 0, -30)},
				{Title: "low_seeders", Seeders: 5, Size: 5 * 1024 * 1024 * 1024, UploadDate: now.AddDate(0, 0, -30)},
				{Title: "too_large", Seeders: 50, Size: 50 * 1024 * 1024 * 1024, UploadDate: now.AddDate(0, 0, -30)},
				{Title: "too_old", Seeders: 50, Size: 5 * 1024 * 1024 * 1024, UploadDate: now.AddDate(0, 0, -200)},
			},
			rule: &models.DownloadRule{
				MinSeeders:   10,
				MaxSeeders:   100,
				MinSize:      1024 * 1024 * 1024,
				MaxSize:      10 * 1024 * 1024 * 1024,
				MaxUploadAge: 90,
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := dm.filterResults(tt.results, tt.rule)
			if len(filtered) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(filtered))
			}
		})
	}
}

func TestSortResults(t *testing.T) {
	dm := &DownloadManager{}
	now := time.Now()

	tests := []struct {
		name          string
		results       []models.SearchResult
		sortBy        string
		expectedFirst string
	}{
		{
			name: "sort by seeders",
			results: []models.SearchResult{
				{Title: "low", Seeders: 10},
				{Title: "high", Seeders: 100},
				{Title: "medium", Seeders: 50},
			},
			sortBy:        "seeders",
			expectedFirst: "high",
		},
		{
			name: "sort by size",
			results: []models.SearchResult{
				{Title: "small", Size: 1024},
				{Title: "large", Size: 10 * 1024 * 1024 * 1024},
				{Title: "medium", Size: 1024 * 1024 * 1024},
			},
			sortBy:        "size",
			expectedFirst: "large",
		},
		{
			name: "sort by upload_date",
			results: []models.SearchResult{
				{Title: "old", UploadDate: now.AddDate(0, 0, -100)},
				{Title: "recent", UploadDate: now.AddDate(0, 0, -1)},
				{Title: "medium", UploadDate: now.AddDate(0, 0, -30)},
			},
			sortBy:        "upload_date",
			expectedFirst: "recent",
		},
		{
			name: "sort by balanced - high seeders and recent",
			results: []models.SearchResult{
				{Title: "old_low_seeders", Seeders: 30, UploadDate: now.AddDate(0, 0, -200)},
				{Title: "recent_low_seeders", Seeders: 10, UploadDate: now.AddDate(0, 0, -1)},
				{Title: "balanced", Seeders: 50, UploadDate: now.AddDate(0, 0, -5)},
			},
			sortBy:        "balanced",
			expectedFirst: "balanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultsCopy := make([]models.SearchResult, len(tt.results))
			copy(resultsCopy, tt.results)

			dm.sortResults(resultsCopy, tt.sortBy)

			if resultsCopy[0].Title != tt.expectedFirst {
				t.Errorf("expected first result to be '%s', got '%s'", tt.expectedFirst, resultsCopy[0].Title)
			}
		})
	}
}

func TestCalculateBalancedScore(t *testing.T) {
	dm := &DownloadManager{}
	now := time.Now()

	tests := []struct {
		name          string
		result        models.SearchResult
		expectedScore float64
	}{
		{
			name: "high seeders, very recent",
			result: models.SearchResult{
				Seeders:    100,
				UploadDate: now.AddDate(0, 0, -3),
			},
			expectedScore: 1100.0,
		},
		{
			name: "low seeders, very recent",
			result: models.SearchResult{
				Seeders:    10,
				UploadDate: now.AddDate(0, 0, -5),
			},
			expectedScore: 200.0,
		},
		{
			name: "high seeders, old",
			result: models.SearchResult{
				Seeders:    100,
				UploadDate: now.AddDate(0, 0, -180),
			},
			expectedScore: 1025.0,
		},
		{
			name: "no upload date",
			result: models.SearchResult{
				Seeders:    50,
				UploadDate: time.Time{},
			},
			expectedScore: 500.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := dm.calculateBalancedScore(&tt.result)
			if score != tt.expectedScore {
				t.Errorf("expected score %.2f, got %.2f", tt.expectedScore, score)
			}
		})
	}
}

func TestGenerateSearchQueries(t *testing.T) {
	dm := &DownloadManager{}

	tests := []struct {
		name          string
		baseQuery     string
		mediaType     string
		expectedCount int
		checkContains []string
	}{
		{
			name:          "custom query - returns single query",
			baseQuery:     "inception",
			mediaType:     "movie",
			expectedCount: 1,
			checkContains: []string{"inception"},
		},
		{
			name:          "empty query for movies - returns genre queries",
			baseQuery:     "",
			mediaType:     "movie",
			expectedCount: 21,
			checkContains: []string{"", "action", "comedy", "thriller", "sci-fi", "horror"},
		},
		{
			name:          "empty query for tv - returns empty query only",
			baseQuery:     "",
			mediaType:     "tv",
			expectedCount: 1,
			checkContains: []string{""},
		},
		{
			name:          "template query - returns single expanded query",
			baseQuery:     "top {year}",
			mediaType:     "movie",
			expectedCount: 1,
			checkContains: []string{fmt.Sprintf("top %d", time.Now().Year())},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := dm.generateSearchQueries(tt.baseQuery, tt.mediaType)

			if len(queries) != tt.expectedCount {
				t.Errorf("expected %d queries, got %d", tt.expectedCount, len(queries))
			}

			for _, check := range tt.checkContains {
				found := false
				for _, query := range queries {
					if query == check {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected queries to contain '%s', got %v", check, queries)
				}
			}
		})
	}
}

func TestDeduplicateResults(t *testing.T) {
	dm := &DownloadManager{}

	tests := []struct {
		name          string
		results       []models.SearchResult
		maxPerTitle   int
		expectedCount int
	}{
		{
			name: "different movies - all pass through",
			results: []models.SearchResult{
				{Title: "Inception 2010 1080p BluRay x264"},
				{Title: "Interstellar 2014 1080p BluRay x264"},
				{Title: "The Matrix 1999 1080p BluRay x264"},
			},
			maxPerTitle:   1,
			expectedCount: 3,
		},
		{
			name: "same movie different quality - deduplicate to 1",
			results: []models.SearchResult{
				{Title: "Inception 2010 1080p BluRay x264"},
				{Title: "Inception 2010 720p BluRay x264"},
				{Title: "Inception.2010.2160p.UHD.BluRay.x265"},
			},
			maxPerTitle:   1,
			expectedCount: 1,
		},
		{
			name: "same movie different quality - allow 2 per title",
			results: []models.SearchResult{
				{Title: "Inception 2010 1080p BluRay x264"},
				{Title: "Inception 2010 720p BluRay x264"},
				{Title: "Inception.2010.2160p.UHD.BluRay.x265"},
			},
			maxPerTitle:   2,
			expectedCount: 2,
		},
		{
			name: "mixed - some duplicates some unique",
			results: []models.SearchResult{
				{Title: "Inception 2010 1080p BluRay x264"},
				{Title: "Inception 2010 720p BluRay x264"},
				{Title: "The Matrix 1999 1080p BluRay x264"},
				{Title: "The Matrix 1999 720p BluRay x264"},
				{Title: "Interstellar 2014 1080p BluRay x264"},
			},
			maxPerTitle:   1,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deduplicated := dm.deduplicateResults(tt.results, tt.maxPerTitle)
			if len(deduplicated) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(deduplicated))
			}
		})
	}
}

func TestNormalizeTitle(t *testing.T) {
	dm := &DownloadManager{}

	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "remove quality tags",
			title:    "Inception 2010 1080p BluRay x264",
			expected: "inception",
		},
		{
			name:     "remove dots and underscores",
			title:    "The.Matrix.1999.1080p.BluRay.x264",
			expected: "the matrix",
		},
		{
			name:     "remove release group",
			title:    "Interstellar.2014.1080p.BluRay.x264-YIFY",
			expected: "interstellar",
		},
		{
			name:     "remove multiple quality indicators",
			title:    "Dune 2021 2160p UHD BluRay HDR x265 10bit",
			expected: "dune",
		},
		{
			name:     "complex title",
			title:    "The.Dark.Knight.2008.1080p.BluRay.x264.DTS-YIFY",
			expected: "the dark knight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := dm.normalizeTitle(tt.title)
			if normalized != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, normalized)
			}
		})
	}
}
