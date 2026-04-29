package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

func TestSearchAndDisplayResults(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	jackettURL := os.Getenv("JACKETT_URL")
	if jackettURL == "" {
		jackettURL = "http://localhost:9117"
	}

	jackettAPIKey := os.Getenv("JACKETT_API_KEY")

	source := &models.DownloadSource{
		ID:      1,
		Name:    "Jackett",
		Type:    "jackett",
		URL:     jackettURL,
		APIKey:  jackettAPIKey,
		Enabled: true,
	}

	provider := torrent.NewJackettProvider(source)
	searcher := torrent.NewTorrentSearcher()
	searcher.AddProvider(provider)

	dm := &DownloadManager{
		torrentSearcher: searcher,
	}

	tests := []struct {
		name     string
		query    string
		category string
	}{
		{
			name:     "current year movies",
			query:    dm.expandQueryTemplates("top movies {year}"),
			category: "movie",
		},
		{
			name:     "static query",
			query:    "inception",
			category: "movie",
		},
		{
			name:     "tv shows",
			query:    "breaking bad",
			category: "tv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("\n=== Search Query: %s (category: %s) ===\n", tt.query, tt.category)

		results, _, err := searcher.Search(tt.query, tt.category, 10, nil)
		if err != nil {
			t.Errorf("Search failed: %v", err)
			return
		}

		fmt.Printf("Found %d results:\n\n", len(results))

			for i, result := range results {
				fmt.Printf("[%d] %s\n", i+1, result.Title)
				fmt.Printf("    Seeders: %d | Leechers: %d | Size: %s\n",
					result.Seeders, result.Leechers, formatSize(result.Size))
				if !result.UploadDate.IsZero() {
					fmt.Printf("    Upload Date: %s\n", result.UploadDate.Format("2006-01-02"))
				}
				if result.InfoHash != "" {
					fmt.Printf("    InfoHash: %s\n", result.InfoHash[:20]+"...")
				}
				fmt.Println()
			}

			if len(results) == 0 {
				t.Log("No results found - this might be expected if Jackett has no configured indexers")
			}
		})
	}
}

func TestExpandQueryTemplates(t *testing.T) {
	dm := &DownloadManager{}

	now := time.Now()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "year template",
			input:    "top movies {year}",
			expected: fmt.Sprintf("top movies %d", now.Year()),
		},
		{
			name:     "month template",
			input:    "releases {year}-{month}",
			expected: fmt.Sprintf("releases %d-%02d", now.Year(), now.Month()),
		},
		{
			name:     "full date template",
			input:    "daily {year}-{month}-{day}",
			expected: fmt.Sprintf("daily %d-%02d-%02d", now.Year(), now.Month(), now.Day()),
		},
		{
			name:     "no templates",
			input:    "inception",
			expected: "inception",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dm.expandQueryTemplates(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSearchForMoviesToDownload(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	jackettURL := os.Getenv("JACKETT_URL")
	if jackettURL == "" {
		jackettURL = "http://localhost:9117"
	}

	jackettAPIKey := os.Getenv("JACKETT_API_KEY")

	source := &models.DownloadSource{
		ID:      1,
		Name:    "Jackett",
		Type:    "jackett",
		URL:     jackettURL,
		APIKey:  jackettAPIKey,
		Enabled: true,
	}

	provider := torrent.NewJackettProvider(source)
	searcher := torrent.NewTorrentSearcher()
	searcher.AddProvider(provider)

	dm := &DownloadManager{
		torrentSearcher: searcher,
	}

	rule := &models.DownloadRule{
		SearchQuery:        "",
		MediaType:          "movie",
		Resolution:         "1080p",
		MinSeeders:         10,
		MinSize:            1024 * 1024 * 1024,
		MaxSize:            20 * 1024 * 1024 * 1024,
		MaxUploadAge:       180,
		MaxResults:         100,
		MaxResultsPerTitle: 1,
		SortBy:             "seeders",
	}

	queries := dm.generateSearchQueries(rule.SearchQuery, rule.MediaType)
	fmt.Printf("\n=== Multi-Query Search Strategy for Top Movies ===\n")
	fmt.Printf("Generated %d search queries for category: %s\n", len(queries), rule.MediaType)
	fmt.Printf("Queries: %v\n", queries)
	fmt.Printf("Resolution: %s\n", rule.Resolution)
	fmt.Printf("Min Seeders: %d\n", rule.MinSeeders)
	fmt.Printf("Size Range: %s - %s\n", formatSize(rule.MinSize), formatSize(rule.MaxSize))
	fmt.Printf("Max Upload Age: %d days\n", rule.MaxUploadAge)
	fmt.Printf("Max Results: %d\n", rule.MaxResults)
	fmt.Printf("Max Per Title: %d\n", rule.MaxResultsPerTitle)
	fmt.Printf("Sort By: %s\n\n", rule.SortBy)

	allResults := make([]models.SearchResult, 0)
	seenHashes := make(map[string]bool)

	for i, query := range queries {
		fmt.Printf("[Query %d/%d] Searching: '%s'\n", i+1, len(queries), query)

		results, _, err := searcher.Search(query, rule.MediaType, 100, nil)
		if err != nil {
			fmt.Printf("  ⚠ Search failed: %v\n\n", err)
			continue
		}

		fmt.Printf("  Found %d results\n", len(results))

		addedCount := 0
		for _, result := range results {
			if result.InfoHash != "" && !seenHashes[result.InfoHash] {
				allResults = append(allResults, result)
				seenHashes[result.InfoHash] = true
				addedCount++
			} else if result.InfoHash == "" {
				allResults = append(allResults, result)
				addedCount++
			}
		}
		fmt.Printf("  Added %d unique results (total: %d)\n\n", addedCount, len(allResults))
	}

	fmt.Printf("=== Aggregation Complete ===\n")
	fmt.Printf("Total unique results across all queries: %d\n\n", len(allResults))

	filtered := dm.filterResults(allResults, rule)
	fmt.Printf("Results after filtering: %d\n\n", len(filtered))

	if len(filtered) == 0 {
		t.Fatal("No results match criteria")
	}

	dm.sortResults(filtered, rule.SortBy)

	deduplicated := dm.deduplicateResults(filtered, rule.MaxResultsPerTitle)
	fmt.Printf("Results after deduplication: %d\n\n", len(deduplicated))

	fmt.Printf("=== Top %d Movies Ready for Download ===\n\n", min(len(deduplicated), rule.MaxResults))

	count := min(len(deduplicated), rule.MaxResults)
	for i := 0; i < count; i++ {
		result := deduplicated[i]
		fmt.Printf("[%d] %s\n", i+1, result.Title)
		fmt.Printf("    Seeders: %d | Leechers: %d | Size: %s\n",
			result.Seeders, result.Leechers, formatSize(result.Size))
		if !result.UploadDate.IsZero() {
			fmt.Printf("    Upload Date: %s (age: %d days)\n",
				result.UploadDate.Format("2006-01-02"),
				int(time.Since(result.UploadDate).Hours()/24))
		}
		if result.InfoHash != "" {
			fmt.Printf("    InfoHash: %s\n", result.InfoHash)
		}
		if result.MagnetLink != "" {
			fmt.Printf("    Magnet: %s...\n", result.MagnetLink[:60])
		}
		fmt.Println()
	}

	fmt.Printf("=== Summary ===\n")
	fmt.Printf("Total queries executed: %d\n", len(queries))
	fmt.Printf("Total results aggregated: %d\n", len(allResults))
	fmt.Printf("After filtering: %d\n", len(filtered))
	fmt.Printf("After deduplication: %d\n", len(deduplicated))
	fmt.Printf("Final selection: %d\n", count)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
