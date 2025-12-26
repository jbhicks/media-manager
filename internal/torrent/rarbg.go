package torrent

import (
	"fmt"
	"regexp"
	"time"

	"github.com/qopher/go-torrentapi"
	"github.com/user/media-manager/pkg/models"
)

type RARBGProvider struct {
	name     string
	enabled  bool
	sourceID uint
	api      *torrentapi.API
}

func NewRARBGProvider(source *models.DownloadSource) (*RARBGProvider, error) {
	api, err := torrentapi.New("media-manager")
	if err != nil {
		return nil, fmt.Errorf("failed to create RARBG API client: %w", err)
	}

	return &RARBGProvider{
		name:     source.Name,
		enabled:  source.Enabled,
		sourceID: source.ID,
		api:      api,
	}, nil
}

func (r *RARBGProvider) GetName() string {
	return r.name
}

func (r *RARBGProvider) IsEnabled() bool {
	return r.enabled
}

func (r *RARBGProvider) Search(query string, category string) ([]models.SearchResult, error) {
	r.api.SearchString(query)

	if category == "movies" {
		r.api.Category(14)
		r.api.Category(17)
		r.api.Category(42)
		r.api.Category(44)
		r.api.Category(45)
		r.api.Category(46)
		r.api.Category(47)
		r.api.Category(48)
		r.api.Category(50)
		r.api.Category(51)
		r.api.Category(52)
		r.api.Category(54)
	} else if category == "tv" {
		r.api.Category(18)
		r.api.Category(41)
		r.api.Category(49)
	}

	r.api.Limit(100)
	r.api.Sort("seeders")
	r.api.Ranked(true)

	rarbgResults, err := r.api.Search()
	if err != nil {
		return nil, fmt.Errorf("RARBG search failed: %w", err)
	}

	results := make([]models.SearchResult, 0, len(rarbgResults))
	expiresAt := time.Now().Add(24 * time.Hour)

	for _, torrent := range rarbgResults {
		uploadDate, _ := time.Parse(time.RFC3339, torrent.PubDate)

		infoHash := extractInfoHash(torrent.Download)

		result := models.SearchResult{
			SourceID:   r.sourceID,
			Title:      torrent.Title,
			InfoHash:   infoHash,
			MagnetLink: torrent.Download,
			TorrentURL: torrent.InfoPage,
			Size:       int64(torrent.Size),
			Seeders:    torrent.Seeders,
			Leechers:   torrent.Leechers,
			UploadDate: uploadDate,
			ExpiresAt:  expiresAt,
		}
		results = append(results, result)
	}

	return results, nil
}

func extractInfoHash(magnetLink string) string {
	re := regexp.MustCompile(`btih:([a-fA-F0-9]{40})`)
	matches := re.FindStringSubmatch(magnetLink)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
