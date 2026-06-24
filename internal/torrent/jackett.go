package torrent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/SkYNewZ/go-jackett"
	"github.com/user/media-manager/pkg/models"
)

type JackettProvider struct {
	name         string
	enabled      bool
	sourceID     uint
	config       jackett.Config
	httpClient   *http.Client
	flareSolverr *FlareSolverrClient
}

func NewJackettProvider(source *models.DownloadSource) *JackettProvider {
	flareSolverrURL := os.Getenv("FLARESOLVERR_URL")

	// Create HTTP client with FlareSolverr fallback support
	var httpClient *http.Client
	var flareClient *FlareSolverrClient

	if flareSolverrURL != "" {
		flareClient = NewFlareSolverrClient(flareSolverrURL, 60*time.Second)
		if err := flareClient.HealthCheck(); err != nil {
			log.Printf("[Jackett] WARNING: FlareSolverr health check failed: %v", err)
			log.Printf("[Jackett] Indexers behind Cloudflare will fail. Start FlareSolverr with: docker run -d -p 8191:8191 ghcr.io/flaresolverr/flaresolverr:latest")
			flareClient = nil
			httpClient = &http.Client{Timeout: 30 * time.Second}
		} else {
			log.Printf("[Jackett] FlareSolverr connected at %s", flareSolverrURL)
			// Use HTTP client with FlareSolverr fallback for direct requests
			httpClient = NewFlareSolverrHTTPClient(flareSolverrURL, 30*time.Second)
		}
	} else {
		log.Printf("[Jackett] FlareSolverr not configured (set FLARESOLVERR_URL env var)")
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	config := &jackett.Config{
		APIUrl:     source.URL,
		APIKey:     source.APIKey,
		HTTPClient: httpClient,
	}
	return &JackettProvider{
		name:         source.Name,
		enabled:      source.Enabled,
		sourceID:     source.ID,
		config:       *config,
		httpClient:   httpClient,
		flareSolverr: flareClient,
	}
}

func (j *JackettProvider) GetName() string {
	return j.name
}

func (j *JackettProvider) IsEnabled() bool {
	return j.enabled
}

func (j *JackettProvider) Search(query string, category string, indexers []string) ([]models.SearchResult, error) {
	jackettBaseURL := strings.TrimSuffix(j.config.APIUrl, "/")
	expiresAt := time.Now().Add(24 * time.Hour)

	allResults := make([]models.SearchResult, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrent indexer searches to prevent overwhelming Jackett
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	// Determine which indexers to search
	var indexersToSearch []string

	if len(indexers) > 0 {
		// Use specified indexers
		indexersToSearch = indexers
	} else {
		// Smart indexer selection based on category
		if category == "movie" || category == "movies" {
			// For movies: prioritize YTS (has posters)
			indexersToSearch = []string{"yts", "1337x", "thepiratebay"}
		} else if category == "tv" {
			// For TV: use TV-focused indexers
			indexersToSearch = []string{"eztv", "1337x", "thepiratebay"}
		} else {
			// Default: search popular indexers
			indexersToSearch = []string{"yts", "1337x", "thepiratebay"}
		}
	}

	// Search each indexer individually (preserves poster URLs)
	for _, indexer := range indexersToSearch {
		wg.Add(1)
		go func(idx string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] Jackett indexer %s recovered from panic: %v\n%s", idx, r, debug.Stack())
				}
				wg.Done()
			}()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results, err := j.searchSingleIndexer(jackettBaseURL, idx, query, category, expiresAt)
			if err != nil {
				log.Printf("[Jackett] %s search failed: %v", idx, err)
				return
			}

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(indexer)
	}

	wg.Wait()

	// If no results from individual indexers, fall back to /all
	if len(allResults) == 0 && len(indexers) == 0 {
		log.Printf("[Jackett] No results from individual indexers, falling back to /all")
		return j.searchAllIndexers(jackettBaseURL, query, category, expiresAt)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no results from any indexer")
	}

	return allResults, nil
}

// searchAllIndexers uses Jackett's /all endpoint to search all configured indexers in parallel
func (j *JackettProvider) searchAllIndexers(baseURL, query, category string, expiresAt time.Time) ([]models.SearchResult, error) {
	searchURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results", baseURL)

	// Use a longer timeout since we're searching all indexers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	q := req.URL.Query()
	q.Set("apikey", j.config.APIKey)
	q.Set("Query", query)

	if category == "movie" || category == "movies" {
		q.Add("Category[]", "2000")
	} else if category == "tv" {
		q.Add("Category[]", "5000")
	}

	req.URL.RawQuery = q.Encode()

	log.Printf("[Jackett] Searching all indexers for: %s", query)
	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	// Use a custom response struct to handle date parsing gracefully
	var result struct {
		Results []struct {
			Title       string      `json:"Title"`
			Tracker     string      `json:"Tracker"`
			Link        string      `json:"Link"`
			Size        int64       `json:"Size"`
			Seeders     int         `json:"Seeders"`
			Peers       int         `json:"Peers"`
			PublishDate string      `json:"PublishDate"`
			MagnetURI   interface{} `json:"MagnetUri"`
			InfoHash    interface{} `json:"InfoHash"`
			Poster      string      `json:"Poster"`
		} `json:"Results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("response decode failed: %w", err)
	}

	log.Printf("[Jackett] All indexers returned %d results", len(result.Results))

	return j.parseRawResults(result.Results, expiresAt), nil
}

// searchSingleIndexer searches a single indexer
func (j *JackettProvider) searchSingleIndexer(baseURL, indexer, query, category string, expiresAt time.Time) ([]models.SearchResult, error) {
	searchURL := fmt.Sprintf("%s/api/v2.0/indexers/%s/results", baseURL, indexer)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("apikey", j.config.APIKey)
	q.Set("Query", query)

	if category == "movie" || category == "movies" {
		q.Add("Category[]", "2000")
	} else if category == "tv" {
		q.Add("Category[]", "5000")
	}

	req.URL.RawQuery = q.Encode()

	log.Printf("[Jackett] Searching %s for: %s", indexer, query)
	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Title       string      `json:"Title"`
			Tracker     string      `json:"Tracker"`
			Link        string      `json:"Link"`
			Size        int64       `json:"Size"`
			Seeders     int         `json:"Seeders"`
			Peers       int         `json:"Peers"`
			PublishDate string      `json:"PublishDate"`
			MagnetURI   interface{} `json:"MagnetUri"`
			InfoHash    interface{} `json:"InfoHash"`
			Poster      string      `json:"Poster"`
		} `json:"Results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	log.Printf("[Jackett] %s returned %d results", indexer, len(result.Results))

	return j.parseRawResults(result.Results, expiresAt), nil
}

// parseResults converts Jackett results to our SearchResult model
func (j *JackettProvider) parseResults(jackettResults []jackett.Result, expiresAt time.Time) []models.SearchResult {
	results := make([]models.SearchResult, 0, len(jackettResults))

	for _, r := range jackettResults {
		leechers := r.Peers - r.Seeders
		if leechers < 0 {
			leechers = 0
		}

		magnetLink := ""
		if r.MagnetURI != nil {
			if uri, ok := r.MagnetURI.(string); ok {
				magnetLink = uri
			}
		}

		infoHash := ""
		if r.InfoHash != nil {
			if hash, ok := r.InfoHash.(string); ok {
				infoHash = hash
			}
		}

		results = append(results, models.SearchResult{
			SourceID:   j.sourceID,
			Indexer:    r.Tracker,
			Title:      r.Title,
			InfoHash:   infoHash,
			MagnetLink: magnetLink,
			TorrentURL: r.Link,
			Size:       r.Size,
			Seeders:    r.Seeders,
			Leechers:   leechers,
			UploadDate: r.PublishDate,
			ExpiresAt:  expiresAt,
		})
	}

	return filterCyrillicResults(results)
}

// filterCyrillicResults removes results with Cyrillic text (mostly Russian trackers)
func filterCyrillicResults(results []models.SearchResult) []models.SearchResult {
	filtered := make([]models.SearchResult, 0, len(results))
	for _, r := range results {
		if !containsCyrillic(r.Title) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// containsCyrillic checks if a string contains Cyrillic characters
func containsCyrillic(s string) bool {
	for _, r := range s {
		if r >= '\u0400' && r <= '\u04FF' || r >= '\u0500' && r <= '\u052F' || r >= '\u2DE0' && r <= '\u2DFF' || r >= '\uA640' && r <= '\uA69F' {
			return true
		}
	}
	return false
}

// parseRawResults converts raw Jackett results (with string dates) to our SearchResult model
func (j *JackettProvider) parseRawResults(jackettResults []struct {
	Title       string      `json:"Title"`
	Tracker     string      `json:"Tracker"`
	Link        string      `json:"Link"`
	Size        int64       `json:"Size"`
	Seeders     int         `json:"Seeders"`
	Peers       int         `json:"Peers"`
	PublishDate string      `json:"PublishDate"`
	MagnetURI   interface{} `json:"MagnetUri"`
	InfoHash    interface{} `json:"InfoHash"`
	Poster      string      `json:"Poster"`
}, expiresAt time.Time) []models.SearchResult {
	results := make([]models.SearchResult, 0, len(jackettResults))

	for _, r := range jackettResults {
		leechers := r.Peers - r.Seeders
		if leechers < 0 {
			leechers = 0
		}

		magnetLink := ""
		if r.MagnetURI != nil {
			if uri, ok := r.MagnetURI.(string); ok {
				magnetLink = uri
			}
		}

		infoHash := ""
		if r.InfoHash != nil {
			if hash, ok := r.InfoHash.(string); ok {
				infoHash = hash
			}
		}

		// Parse date - handle multiple formats
		var uploadDate time.Time
		if r.PublishDate != "" {
			// Try RFC3339 first
			uploadDate, _ = time.Parse(time.RFC3339, r.PublishDate)
			if uploadDate.IsZero() {
				// Try without timezone
				uploadDate, _ = time.Parse("2006-01-02T15:04:05", r.PublishDate)
			}
		}

		results = append(results, models.SearchResult{
			SourceID:   j.sourceID,
			Indexer:    r.Tracker,
			Title:      r.Title,
			InfoHash:   infoHash,
			MagnetLink: magnetLink,
			TorrentURL: r.Link,
			Size:       r.Size,
			Seeders:    r.Seeders,
			Leechers:   leechers,
			UploadDate: uploadDate,
			ExpiresAt:  expiresAt,
			PosterURL:  r.Poster,
		})
	}

	return filterCyrillicResults(results)
}

// searchWithFlareSolverr attempts to search indexers directly using FlareSolverr bypass
// This is the integration point between go-jackett and go-flaresolverr
func (j *JackettProvider) searchWithFlareSolverr(ctx context.Context, query string, category string) ([]models.SearchResult, error) {
	if j.flareSolverr == nil || !j.flareSolverr.IsEnabled() {
		return nil, fmt.Errorf("flaresolverr not available")
	}

	results := make([]models.SearchResult, 0)

	// Try YTS directly (doesn't always need FlareSolverr but benefits from it)
	if category == "movie" || category == "movies" {
		ytsResults, err := j.searchYTS(ctx, query)
		if err == nil {
			results = append(results, ytsResults...)
		}
	}

	// Try 1337x via FlareSolverr
	torrentResults, err := j.search1337x(ctx, query)
	if err == nil {
		results = append(results, torrentResults...)
	}

	return results, nil
}

// searchYTS searches YTS directly via their API (no Cloudflare usually)
func (j *JackettProvider) searchYTS(ctx context.Context, query string) ([]models.SearchResult, error) {
	apiURL := fmt.Sprintf("https://yts.mx/api/v2/list_movies.json?query_term=%s", url.QueryEscape(query))

	resp, err := j.flareSolverr.Get(ctx, apiURL)
	if err != nil {
		return nil, err
	}

	// Parse YTS response
	if resp.Solution == nil {
		return nil, fmt.Errorf("no solution from flaresolverr")
	}

	// Simple parsing - in production you'd want proper JSON parsing
	results := make([]models.SearchResult, 0)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Note: This is a simplified example. Full implementation would parse the JSON
	_ = expiresAt

	return results, nil
}

// search1337x searches 1337x via FlareSolverr bypass
func (j *JackettProvider) search1337x(ctx context.Context, query string) ([]models.SearchResult, error) {
	searchURL := fmt.Sprintf("https://1337x.to/search/%s/1/", url.QueryEscape(query))

	resp, err := j.flareSolverr.Get(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	if resp.Solution == nil {
		return nil, fmt.Errorf("no solution from flaresolverr")
	}

	// Parse HTML response - in production you'd want goquery or similar
	results := make([]models.SearchResult, 0)

	return results, nil
}
