package torrent

import (
	"fmt"
	"log"

	"github.com/user/media-manager/pkg/models"
)

type SearchProvider interface {
	Search(query string, category string) ([]models.SearchResult, error)
	GetName() string
	IsEnabled() bool
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

func (ts *TorrentSearcher) Search(query string, category string, maxResults int) ([]models.SearchResult, error) {
	allResults := make([]models.SearchResult, 0)

	for _, provider := range ts.providers {
		if !provider.IsEnabled() {
			continue
		}

		log.Printf("[TORRENT] Searching %s for: %s", provider.GetName(), query)
		results, err := provider.Search(query, category)
		if err != nil {
			log.Printf("[WARN] Search failed on %s: %v", provider.GetName(), err)
			continue
		}

		log.Printf("[TORRENT] Found %d results from %s", len(results), provider.GetName())
		allResults = append(allResults, results...)

		if len(allResults) >= maxResults {
			break
		}
	}

	if len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	return allResults, nil
}

func (ts *TorrentSearcher) SearchMovies(title string, year int, quality string) ([]models.SearchResult, error) {
	query := title
	if year > 0 {
		query = fmt.Sprintf("%s %d", title, year)
	}
	if quality != "" {
		query = fmt.Sprintf("%s %s", query, quality)
	}

	return ts.Search(query, "movies", 50)
}

func (ts *TorrentSearcher) SearchTV(title string, season int, episode int, quality string) ([]models.SearchResult, error) {
	query := title
	if season > 0 && episode > 0 {
		query = fmt.Sprintf("%s S%02dE%02d", title, season, episode)
	} else if season > 0 {
		query = fmt.Sprintf("%s S%02d", title, season)
	}
	if quality != "" {
		query = fmt.Sprintf("%s %s", query, quality)
	}

	return ts.Search(query, "tv", 50)
}
