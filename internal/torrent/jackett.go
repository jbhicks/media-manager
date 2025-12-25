package torrent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SkYNewZ/go-jackett"
	"github.com/user/media-manager/pkg/models"
)

type JackettProvider struct {
	name     string
	enabled  bool
	sourceID uint
	client   jackett.Client
}

func NewJackettProvider(source *models.DownloadSource) *JackettProvider {
	config := &jackett.Config{
		APIUrl:     source.URL,
		APIKey:     source.APIKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	client := jackett.New(config)

	return &JackettProvider{
		name:     source.Name,
		enabled:  source.Enabled,
		sourceID: source.ID,
		client:   client,
	}
}

func (j *JackettProvider) GetName() string {
	return j.name
}

func (j *JackettProvider) IsEnabled() bool {
	return j.enabled
}

func (j *JackettProvider) Search(query string, category string) ([]models.SearchResult, error) {
	opts := []jackett.Option{
		jackett.WithQuery(query),
	}

	if category == "movie" || category == "movies" {
		opts = append(opts, jackett.WithCategory(2000))
	} else if category == "tv" {
		opts = append(opts, jackett.WithCategory(5000))
	}

	ctx := context.Background()
	resp, err := j.client.Fetch(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("jackett search failed: %w", err)
	}

	results := make([]models.SearchResult, 0, len(resp.Results))
	expiresAt := time.Now().Add(24 * time.Hour)

	for _, r := range resp.Results {
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

		result := models.SearchResult{
			SourceID:   j.sourceID,
			Title:      r.Title,
			InfoHash:   infoHash,
			MagnetLink: magnetLink,
			TorrentURL: r.Link,
			Size:       r.Size,
			Seeders:    r.Seeders,
			Leechers:   leechers,
			UploadDate: r.PublishDate,
			ExpiresAt:  expiresAt,
		}
		results = append(results, result)
	}

	return results, nil
}
