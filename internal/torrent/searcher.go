package torrent

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/user/media-manager/pkg/models"
)

type SearchProvider interface {
	Search(query string, category string, indexers []string) ([]models.SearchResult, error)
	GetName() string
	IsEnabled() bool
}

type ProviderTiming struct {
	Name       string  `json:"name"`
	Duration   float64 `json:"duration_seconds"`
	ResultCount int    `json:"result_count"`
	Error      string  `json:"error,omitempty"`
}

type TorrentSearcher struct {
	providers []SearchProvider
}

func NewTorrentSearcher() *TorrentSearcher {
	return &TorrentSearcher{
		providers: make([]SearchProvider, 0),
	}
}

func (ts *TorrentSearcher) AddProvider(provider SearchProvider) {
	ts.providers = append(ts.providers, provider)
	log.Printf("[TORRENT] Added search provider: %s", provider.GetName())
}

func (ts *TorrentSearcher) Search(query string, category string, maxResults int, indexers []string) ([]models.SearchResult, []ProviderTiming, error) {
	allResults := make([]models.SearchResult, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	timings := make([]ProviderTiming, 0)
	var timingMu sync.Mutex

	for _, provider := range ts.providers {
		if !provider.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(p SearchProvider) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] Provider %s recovered from panic: %v\n%s", p.GetName(), r, debug.Stack())
				}
				wg.Done()
			}()

			log.Printf("[TORRENT] Searching %s for: %s", p.GetName(), query)
			start := time.Now()
			results, err := p.Search(query, category, indexers)
			duration := time.Since(start)

			timing := ProviderTiming{
				Name:     p.GetName(),
				Duration: duration.Seconds(),
			}

			if err != nil {
				log.Printf("[WARN] Search failed on %s: %v", p.GetName(), err)
				timing.Error = err.Error()
				timingMu.Lock()
				timings = append(timings, timing)
				timingMu.Unlock()
				return
			}

			log.Printf("[TORRENT] Found %d results from %s (%.2fs)", len(results), p.GetName(), duration.Seconds())
			timing.ResultCount = len(results)
			timingMu.Lock()
			timings = append(timings, timing)
			timingMu.Unlock()

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(provider)
	}

	wg.Wait()

	return allResults, timings, nil
}

func (ts *TorrentSearcher) SearchMovies(title string, year int, quality string) ([]models.SearchResult, []ProviderTiming, error) {
	query := title
	if year > 0 {
		query = fmt.Sprintf("%s %d", title, year)
	}
	if quality != "" {
		query = fmt.Sprintf("%s %s", query, quality)
	}

	return ts.Search(query, "movies", 50, nil)
}

func (ts *TorrentSearcher) SearchTV(title string, season int, episode int, quality string) ([]models.SearchResult, []ProviderTiming, error) {
	query := title
	if season > 0 && episode > 0 {
		query = fmt.Sprintf("%s S%02dE%02d", title, season, episode)
	} else if season > 0 {
		query = fmt.Sprintf("%s S%02d", title, season)
	}
	if quality != "" {
		query = fmt.Sprintf("%s %s", query, quality)
	}

	return ts.Search(query, "tv", 50, nil)
}

// ExtractInfoHash extracts the info hash from a magnet link
func ExtractInfoHash(magnet string) string {
	// magnet:?xt=urn:btih:HASH&...
	if idx := strings.Index(magnet, "btih:"); idx != -1 {
		start := idx + 5
		end := strings.Index(magnet[start:], "&")
		if end == -1 {
			return magnet[start:]
		}
		return magnet[start : start+end]
	}
	return ""
}
