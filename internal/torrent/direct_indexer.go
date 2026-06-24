package torrent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/SkYNewZ/go-flaresolverr"
	"github.com/google/uuid"
	"github.com/user/media-manager/pkg/models"
)

// DirectIndexer provides direct indexer scraping without Jackett
type DirectIndexer struct {
	name         string
	enabled      bool
	sourceID     uint
	flareSolverr flaresolverr.Client
	httpClient   *http.Client
}

// NewDirectIndexer creates an indexer that scrapes directly (no Jackett needed)
func NewDirectIndexer(source *models.DownloadSource) *DirectIndexer {
	flareSolverrURL := source.URL // Use the URL field for FlareSolverr URL

	var flareClient flaresolverr.Client
	if flareSolverrURL != "" {
		flareClient = flaresolverr.New(flareSolverrURL, 60*time.Second, nil)
	}

	return &DirectIndexer{
		name:         source.Name,
		enabled:      source.Enabled,
		sourceID:     source.ID,
		flareSolverr: flareClient,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *DirectIndexer) GetName() string {
	return d.name
}

func (d *DirectIndexer) IsEnabled() bool {
	return d.enabled
}

// Search searches indexers directly using FlareSolverr for Cloudflare bypass
func (d *DirectIndexer) Search(query string, category string, indexers []string) ([]models.SearchResult, error) {
	if d.flareSolverr == nil {
		return nil, fmt.Errorf("flaresolverr not configured for direct indexing")
	}

	results := make([]models.SearchResult, 0)

	// Search multiple indexers in parallel
	ctx := context.Background()

	// Try YTS (movies)
	if category == "movie" || category == "movies" {
		ytsResults, err := d.searchYTS(ctx, query)
		if err != nil {
			log.Printf("[DirectIndexer] YTS search failed: %v", err)
		} else {
			results = append(results, ytsResults...)
		}
	}

	// Try 1337x
	_1337xResults, err := d.search1337x(ctx, query)
	if err != nil {
		log.Printf("[DirectIndexer] 1337x search failed: %v", err)
	} else {
		results = append(results, _1337xResults...)
	}

	return results, nil
}

// searchYTS searches YTS API directly
func (d *DirectIndexer) searchYTS(ctx context.Context, query string) ([]models.SearchResult, error) {
	apiURL := fmt.Sprintf("https://yts.mx/api/v2/list_movies.json?query_term=%s&sort_by=seeders&order_by=desc", url.QueryEscape(query))

	resp, err := d.flareSolverr.Get(ctx, apiURL, uuid.Nil, nil)
	if err != nil {
		return nil, fmt.Errorf("yts flaresolverr request failed: %w", err)
	}

	if resp.Solution == nil {
		return nil, fmt.Errorf("no solution from flaresolverr")
	}

	// Parse JSON response
	var ytsResp struct {
		Data struct {
			Movies []struct {
				Title    string `json:"title"`
				Year     int    `json:"year"`
				Torrents []struct {
					URL     string `json:"url"`
					Hash    string `json:"hash"`
					Quality string `json:"quality"`
					Size    string `json:"size"`
					Seeds   int    `json:"seeds"`
					Peers   int    `json:"peers"`
				} `json:"torrents"`
			} `json:"movies"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(resp.Solution.Response), &ytsResp); err != nil {
		return nil, fmt.Errorf("failed to parse YTS response: %w", err)
	}

	results := make([]models.SearchResult, 0)
	expiresAt := time.Now().Add(24 * time.Hour)

	for _, movie := range ytsResp.Data.Movies {
		for _, torrent := range movie.Torrents {
			result := models.SearchResult{
				SourceID:   d.sourceID,
				Title:      fmt.Sprintf("%s (%d) [%s]", movie.Title, movie.Year, torrent.Quality),
				InfoHash:   torrent.Hash,
				MagnetLink: fmt.Sprintf("magnet:?xt=urn:btih:%s&dn=%s", torrent.Hash, url.QueryEscape(movie.Title)),
				Size:       parseSize(torrent.Size),
				Seeders:    torrent.Seeds,
				Leechers:   torrent.Peers,
				ExpiresAt:  expiresAt,
			}
			results = append(results, result)
		}
	}

	log.Printf("[DirectIndexer] YTS returned %d results", len(results))
	return results, nil
}

// search1337x searches 1337x via FlareSolverr bypass
func (d *DirectIndexer) search1337x(ctx context.Context, query string) ([]models.SearchResult, error) {
	searchURL := fmt.Sprintf("https://1337x.to/search/%s/1/", url.QueryEscape(query))

	resp, err := d.flareSolverr.Get(ctx, searchURL, uuid.Nil, nil)
	if err != nil {
		return nil, fmt.Errorf("1337x flaresolverr request failed: %w", err)
	}

	if resp.Solution == nil {
		return nil, fmt.Errorf("no solution from flaresolverr")
	}

	// In a real implementation, you'd parse the HTML here
	// For now, return empty results
	log.Printf("[DirectIndexer] 1337x page loaded successfully, HTML parsing not yet implemented")
	return []models.SearchResult{}, nil
}

func parseSize(sizeStr string) int64 {
	// Parse size strings like "1.5 GB", "800 MB"
	var size float64
	var unit string
	fmt.Sscanf(sizeStr, "%f %s", &size, &unit)

	switch unit {
	case "GB":
		return int64(size * 1024 * 1024 * 1024)
	case "MB":
		return int64(size * 1024 * 1024)
	case "KB":
		return int64(size * 1024)
	default:
		return 0
	}
}
