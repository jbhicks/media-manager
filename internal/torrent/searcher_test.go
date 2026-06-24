package torrent

import (
	"fmt"
	"testing"

	"github.com/user/media-manager/pkg/models"
)

type mockSearchProvider struct {
	name     string
	enabled  bool
	results  []models.SearchResult
	searchFn func(query string, category string, indexers []string) ([]models.SearchResult, error)
}

func (m *mockSearchProvider) Search(query string, category string, indexers []string) ([]models.SearchResult, error) {
	if m.searchFn != nil {
		return m.searchFn(query, category, indexers)
	}
	return m.results, nil
}

func (m *mockSearchProvider) GetName() string {
	return m.name
}

func (m *mockSearchProvider) IsEnabled() bool {
	return m.enabled
}

func TestTorrentSearcher_AddProvider(t *testing.T) {
	ts := NewTorrentSearcher()

	provider := &mockSearchProvider{name: "test-provider", enabled: true}
	ts.AddProvider(provider)

	if len(ts.providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(ts.providers))
	}
}

func TestTorrentSearcher_Search(t *testing.T) {
	tests := []struct {
		name          string
		providers     []SearchProvider
		query         string
		category      string
		maxResults    int
		expectedCount int
	}{
		{
			name: "single provider with results",
			providers: []SearchProvider{
				&mockSearchProvider{
					name:    "provider1",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 1"},
						{Title: "Result 2"},
					},
				},
			},
			query:         "test",
			category:      "movies",
			maxResults:    10,
			expectedCount: 2,
		},
		{
			name: "multiple providers",
			providers: []SearchProvider{
				&mockSearchProvider{
					name:    "provider1",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 1"},
						{Title: "Result 2"},
					},
				},
				&mockSearchProvider{
					name:    "provider2",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 3"},
						{Title: "Result 4"},
					},
				},
			},
			query:         "test",
			category:      "movies",
			maxResults:    10,
			expectedCount: 4,
		},
		{
			name: "disabled provider ignored",
			providers: []SearchProvider{
				&mockSearchProvider{
					name:    "provider1",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 1"},
					},
				},
				&mockSearchProvider{
					name:    "provider2",
					enabled: false,
					results: []models.SearchResult{
						{Title: "Result 2"},
					},
				},
			},
			query:         "test",
			category:      "movies",
			maxResults:    10,
			expectedCount: 1,
		},
		{
			name: "max results limit",
			providers: []SearchProvider{
				&mockSearchProvider{
					name:    "provider1",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 1"},
						{Title: "Result 2"},
						{Title: "Result 3"},
						{Title: "Result 4"},
						{Title: "Result 5"},
					},
				},
			},
			query:         "test",
			category:      "movies",
			maxResults:    3,
			expectedCount: 3,
		},
		{
			name: "provider error continues to next",
			providers: []SearchProvider{
				&mockSearchProvider{
					name:    "provider1",
					enabled: true,
					searchFn: func(query, category string, indexers []string) ([]models.SearchResult, error) {
						return nil, fmt.Errorf("search failed")
					},
				},
				&mockSearchProvider{
					name:    "provider2",
					enabled: true,
					results: []models.SearchResult{
						{Title: "Result 1"},
					},
				},
			},
			query:         "test",
			category:      "movies",
			maxResults:    10,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTorrentSearcher()
			for _, provider := range tt.providers {
				ts.AddProvider(provider)
			}

			results, _, err := ts.Search(tt.query, tt.category, tt.maxResults, nil)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

func TestTorrentSearcher_SearchMovies(t *testing.T) {
	ts := NewTorrentSearcher()
	provider := &mockSearchProvider{
		name:    "test-provider",
		enabled: true,
		searchFn: func(query, category string, indexers []string) ([]models.SearchResult, error) {
			if category != "movies" {
				t.Errorf("expected category 'movies', got '%s'", category)
			}
			return []models.SearchResult{{Title: query}}, nil
		},
	}
	ts.AddProvider(provider)

	tests := []struct {
		name          string
		title         string
		year          int
		quality       string
		expectedQuery string
	}{
		{
			name:          "title only",
			title:         "Inception",
			year:          0,
			quality:       "",
			expectedQuery: "Inception",
		},
		{
			name:          "title and year",
			title:         "Inception",
			year:          2010,
			quality:       "",
			expectedQuery: "Inception 2010",
		},
		{
			name:          "title, year, and quality",
			title:         "Inception",
			year:          2010,
			quality:       "1080p",
			expectedQuery: "Inception 2010 1080p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := ts.SearchMovies(tt.title, tt.year, tt.quality)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("expected 1 result, got %d", len(results))
			}

			if results[0].Title != tt.expectedQuery {
				t.Errorf("expected query '%s', got '%s'", tt.expectedQuery, results[0].Title)
			}
		})
	}
}

func TestTorrentSearcher_SearchTV(t *testing.T) {
	ts := NewTorrentSearcher()
	provider := &mockSearchProvider{
		name:    "test-provider",
		enabled: true,
		searchFn: func(query, category string, indexers []string) ([]models.SearchResult, error) {
			if category != "tv" {
				t.Errorf("expected category 'tv', got '%s'", category)
			}
			return []models.SearchResult{{Title: query}}, nil
		},
	}
	ts.AddProvider(provider)

	tests := []struct {
		name          string
		title         string
		season        int
		episode       int
		quality       string
		expectedQuery string
	}{
		{
			name:          "title only",
			title:         "Breaking Bad",
			season:        0,
			episode:       0,
			quality:       "",
			expectedQuery: "Breaking Bad",
		},
		{
			name:          "season only",
			title:         "Breaking Bad",
			season:        1,
			episode:       0,
			quality:       "",
			expectedQuery: "Breaking Bad S01",
		},
		{
			name:          "season and episode",
			title:         "Breaking Bad",
			season:        1,
			episode:       1,
			quality:       "",
			expectedQuery: "Breaking Bad S01E01",
		},
		{
			name:          "with quality",
			title:         "Breaking Bad",
			season:        1,
			episode:       1,
			quality:       "1080p",
			expectedQuery: "Breaking Bad S01E01 1080p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := ts.SearchTV(tt.title, tt.season, tt.episode, tt.quality)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("expected 1 result, got %d", len(results))
			}

			if results[0].Title != tt.expectedQuery {
				t.Errorf("expected query '%s', got '%s'", tt.expectedQuery, results[0].Title)
			}
		})
	}
}
