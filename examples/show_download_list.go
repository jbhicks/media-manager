package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/user/media-manager/internal/service"
	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

func main() {
	jackettURL := os.Getenv("JACKETT_URL")
	if jackettURL == "" {
		jackettURL = "http://localhost:9117"
	}

	jackettAPIKey := os.Getenv("JACKETT_API_KEY")
	if jackettAPIKey == "" {
		fmt.Println("Error: JACKETT_API_KEY environment variable is required")
		fmt.Println("Usage: JACKETT_API_KEY=your-key go run examples/show_download_list.go")
		os.Exit(1)
	}

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

	dm := &service.DownloadManager{}

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

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  MULTI-QUERY TORRENT SEARCH - TARGETED DOWNLOAD LIST")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Search Configuration:")
	fmt.Printf("  • Media Type: %s\n", rule.MediaType)
	fmt.Printf("  • Resolution: %s\n", rule.Resolution)
	fmt.Printf("  • Min Seeders: %d\n", rule.MinSeeders)
	fmt.Printf("  • Size Range: %s - %s\n", formatSize(rule.MinSize), formatSize(rule.MaxSize))
	fmt.Printf("  • Max Age: %d days (%.1f months)\n", rule.MaxUploadAge, float64(rule.MaxUploadAge)/30)
	fmt.Printf("  • Max Results: %d\n", rule.MaxResults)
	fmt.Printf("  • Max Per Title: %d\n", rule.MaxResultsPerTitle)
	fmt.Printf("  • Sort By: %s\n", rule.SortBy)
	fmt.Println()

	queries := generateQueries(rule.SearchQuery, rule.MediaType)
	fmt.Printf("Generated %d search queries:\n", len(queries))
	for i, q := range queries {
		if q == "" {
			fmt.Printf("  %2d. <empty> (popular torrents)\n", i+1)
		} else {
			fmt.Printf("  %2d. %s\n", i+1, q)
		}
	}
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  EXECUTING SEARCHES...")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	allResults := make([]models.SearchResult, 0)
	seenHashes := make(map[string]bool)

	for i, query := range queries {
		displayQuery := query
		if displayQuery == "" {
			displayQuery = "<popular>"
		}
		fmt.Printf("[%2d/%2d] Searching: %-15s ... ", i+1, len(queries), displayQuery)

		results, err := searcher.Search(query, rule.MediaType, 100)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			continue
		}

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
		fmt.Printf("✓ Found %d, Added %d unique (total: %d)\n", len(results), addedCount, len(allResults))
	}

	fmt.Println()
	fmt.Printf("Total unique results collected: %d\n", len(allResults))
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  FILTERING & PROCESSING...")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	filtered := filterResults(dm, allResults, rule)
	fmt.Printf("After filtering (seeders, size, age, resolution): %d results\n", len(filtered))

	if len(filtered) == 0 {
		fmt.Println("\n❌ No results match criteria!")
		return
	}

	sortResults(dm, filtered, rule.SortBy)
	fmt.Printf("After sorting by %s: %d results\n", rule.SortBy, len(filtered))

	deduplicated := deduplicateResults(dm, filtered, rule.MaxResultsPerTitle)
	fmt.Printf("After deduplication (max %d per title): %d results\n", rule.MaxResultsPerTitle, len(deduplicated))

	count := min(len(deduplicated), rule.MaxResults)
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  TOP %d MOVIES READY FOR DOWNLOAD\n", count)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	for i := 0; i < count; i++ {
		result := deduplicated[i]

		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("[%d] %s\n", i+1, result.Title)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		fmt.Printf("  📊 Seeders: %-5d | Leechers: %-5d | Ratio: %.2f\n",
			result.Seeders, result.Leechers, calculateRatio(result.Seeders, result.Leechers))

		fmt.Printf("  💾 Size: %s\n", formatSize(result.Size))

		if !result.UploadDate.IsZero() {
			age := int(time.Since(result.UploadDate).Hours() / 24)
			fmt.Printf("  📅 Upload Date: %s (age: %d days / %.1f months)\n",
				result.UploadDate.Format("2006-01-02"),
				age,
				float64(age)/30)
		}

		if result.InfoHash != "" {
			fmt.Printf("  🔑 InfoHash: %s\n", result.InfoHash)
		}

		if result.MagnetLink != "" {
			fmt.Printf("  🧲 Magnet: %s...\n", truncate(result.MagnetLink, 80))
		}

		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  SUMMARY STATISTICS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Search queries executed: %d\n", len(queries))
	fmt.Printf("  Raw results aggregated: %d\n", len(allResults))
	fmt.Printf("  After filtering: %d (%.1f%% passed)\n",
		len(filtered), float64(len(filtered))/float64(max(len(allResults), 1))*100)
	fmt.Printf("  After deduplication: %d unique movies\n", len(deduplicated))
	fmt.Printf("  Final selection: %d\n", count)
	fmt.Println()

	if len(deduplicated) > 0 {
		avgSeeders := 0
		minSeeders := deduplicated[0].Seeders
		maxSeeders := deduplicated[0].Seeders

		for i := 0; i < count; i++ {
			avgSeeders += deduplicated[i].Seeders
			if deduplicated[i].Seeders < minSeeders {
				minSeeders = deduplicated[i].Seeders
			}
			if deduplicated[i].Seeders > maxSeeders {
				maxSeeders = deduplicated[i].Seeders
			}
		}
		avgSeeders /= count

		fmt.Printf("  Seeder stats (top %d):\n", count)
		fmt.Printf("    • Average: %d seeders\n", avgSeeders)
		fmt.Printf("    • Range: %d - %d seeders\n", minSeeders, maxSeeders)
		fmt.Println()
	}

	fmt.Println("✅ Download list generation complete!")
	fmt.Println()
}

func generateQueries(baseQuery string, mediaType string) []string {
	if baseQuery != "" {
		return []string{baseQuery}
	}

	if mediaType != "movie" {
		return []string{""}
	}

	return []string{
		"",
		"action",
		"adventure",
		"comedy",
		"drama",
		"thriller",
		"horror",
		"sci-fi",
		"fantasy",
		"romance",
		"crime",
		"mystery",
		"animation",
		"documentary",
		"family",
		"superhero",
		"war",
		"western",
		"musical",
		"biography",
		"sport",
	}
}

func filterResults(dm *service.DownloadManager, results []models.SearchResult, rule *models.DownloadRule) []models.SearchResult {
	filtered := make([]models.SearchResult, 0)
	now := time.Now()

	for _, result := range results {
		if rule.MinSeeders > 0 && result.Seeders < rule.MinSeeders {
			continue
		}

		if rule.MaxSeeders > 0 && result.Seeders > rule.MaxSeeders {
			continue
		}

		if rule.MinSize > 0 && result.Size < rule.MinSize {
			continue
		}

		if rule.MaxSize > 0 && result.Size > rule.MaxSize {
			continue
		}

		if rule.MinUploadAge > 0 && !result.UploadDate.IsZero() {
			age := int(now.Sub(result.UploadDate).Hours() / 24)
			if age < rule.MinUploadAge {
				continue
			}
		}

		if rule.MaxUploadAge > 0 && !result.UploadDate.IsZero() {
			age := int(now.Sub(result.UploadDate).Hours() / 24)
			if age > rule.MaxUploadAge {
				continue
			}
		}

		if rule.Resolution != "" {
			if !matchesResolution(result.Title, rule.Resolution) {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

func matchesResolution(title string, resolution string) bool {
	titleLower := strings.ToLower(title)

	resolutionPatterns := map[string][]string{
		"2160p": {"2160p", "4k", "uhd"},
		"1080p": {"1080p"},
		"720p":  {"720p"},
		"480p":  {"480p"},
		"4k":    {"2160p", "4k", "uhd"},
		"uhd":   {"2160p", "4k", "uhd"},
	}

	patterns := resolutionPatterns[resolution]
	if patterns == nil {
		patterns = []string{resolution}
	}

	for _, pattern := range patterns {
		if strings.Contains(titleLower, pattern) {
			return true
		}
	}

	return false
}

func sortResults(dm *service.DownloadManager, results []models.SearchResult, sortBy string) {
	// Simple bubble sort by seeders (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Seeders < results[j].Seeders {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func deduplicateResults(dm *service.DownloadManager, results []models.SearchResult, maxPerTitle int) []models.SearchResult {
	if maxPerTitle <= 0 {
		maxPerTitle = 1
	}

	titleCounts := make(map[string]int)
	deduplicated := make([]models.SearchResult, 0)

	for _, result := range results {
		normalizedTitle := normalizeTitle(result.Title)
		count := titleCounts[normalizedTitle]

		if count < maxPerTitle {
			deduplicated = append(deduplicated, result)
			titleCounts[normalizedTitle] = count + 1
		}
	}

	return deduplicated
}

func normalizeTitle(title string) string {
	lower := strings.ToLower(title)

	replacements := map[string]string{
		".":  " ",
		"-":  " ",
		"_":  " ",
		"  ": " ",
	}
	for k, v := range replacements {
		lower = strings.ReplaceAll(lower, k, v)
	}

	qualityTerms := []string{
		"1080p", "720p", "2160p", "4k", "uhd", "hdr", "bluray", "brrip", "webrip", "web-dl",
		"x264", "x265", "h264", "h265", "hevc", "10bit", "dts", "ac3", "aac", "opus",
		"yify", "rarbg", "galaxyrg", "yts", "proper", "repack", "extended", "unrated",
		"(", ")", "[", "]", "2010", "2011", "2012", "2013", "2014", "2015", "2016",
		"2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025",
	}

	for _, term := range qualityTerms {
		lower = strings.ReplaceAll(lower, term, "")
	}

	return strings.TrimSpace(lower)
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

func calculateRatio(seeders, leechers int) float64 {
	if leechers == 0 {
		if seeders == 0 {
			return 0.0
		}
		return 999.9
	}
	return float64(seeders) / float64(leechers)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
