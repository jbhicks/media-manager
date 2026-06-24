package service

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

// RSSItem represents an item in an RSS feed
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Enclosure   struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

// RSSChannel represents the channel in an RSS feed
type RSSChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

// RSSDocument represents the root of an RSS feed
type RSSDocument struct {
	Channel RSSChannel `xml:"channel"`
}

// RSSMonitor monitors RSS feeds for new torrents
type RSSMonitor struct {
	db              *db.Database
	downloadManager *DownloadManager
	stopChan        chan bool
	isRunning       bool
}

// NewRSSMonitor creates a new RSS monitor
func NewRSSMonitor(database *db.Database, dm *DownloadManager) *RSSMonitor {
	return &RSSMonitor{
		db:              database,
		downloadManager: dm,
		stopChan:        make(chan bool),
	}
}

// Start begins monitoring RSS feeds
func (r *RSSMonitor) Start() {
	if r.isRunning {
		return
	}
	r.isRunning = true

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-r.stopChan:
				log.Println("[RSS] Monitor stopped")
				return
			case <-ticker.C:
				r.checkFeeds()
			}
		}
	}()

	log.Println("[RSS] Monitor started")
}

// Stop halts the RSS monitor
func (r *RSSMonitor) Stop() {
	if !r.isRunning {
		return
	}
	r.isRunning = false
	close(r.stopChan)
}

// checkFeeds checks all enabled RSS feeds
func (r *RSSMonitor) checkFeeds() {
	var feeds []models.RSSFeed
	if err := r.db.GetDB().Where("enabled = ?", true).Find(&feeds).Error; err != nil {
		log.Printf("[RSS] Failed to fetch feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		// Check if it's time to check this feed
		if time.Since(feed.LastCheck) < time.Duration(feed.Interval)*time.Minute {
			continue
		}

		log.Printf("[RSS] Checking feed: %s (%s)", feed.Name, feed.URL)

		if err := r.checkFeed(&feed); err != nil {
			log.Printf("[RSS] Feed check failed for %s: %v", feed.Name, err)
			r.db.GetDB().Model(&feed).Updates(map[string]interface{}{
				"last_error": err.Error(),
				"last_check": time.Now(),
			})
		} else {
			r.db.GetDB().Model(&feed).Updates(map[string]interface{}{
				"last_error": "",
				"last_check": time.Now(),
			})
		}
	}
}

// checkFeed checks a single RSS feed for new torrents
func (r *RSSMonitor) checkFeed(feed *models.RSSFeed) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(feed.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read feed body: %w", err)
	}

	var doc RSSDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("failed to parse feed: %w", err)
	}

	log.Printf("[RSS] Feed %s has %d items", feed.Name, len(doc.Channel.Items))

	for _, item := range doc.Channel.Items {
		if err := r.processRSSItem(feed, &item); err != nil {
			log.Printf("[RSS] Failed to process item '%s': %v", item.Title, err)
		}
	}

	return nil
}

// processRSSItem processes a single RSS item
func (r *RSSMonitor) processRSSItem(feed *models.RSSFeed, item *RSSItem) error {
	// Extract magnet link from description or enclosure
	magnetLink := ""
	if strings.HasPrefix(item.Link, "magnet:") {
		magnetLink = item.Link
	} else if strings.HasPrefix(item.Enclosure.URL, "magnet:") {
		magnetLink = item.Enclosure.URL
	} else {
		// Try to extract magnet link from description
		magnetRegex := regexp.MustCompile(`magnet:\?xt=urn:btih:[a-zA-Z0-9]+`)
		matches := magnetRegex.FindString(item.Description)
		if matches != "" {
			magnetLink = matches
		}
	}

	if magnetLink == "" {
		return fmt.Errorf("no magnet link found")
	}

	// Parse torrent info from title
	title := item.Title
	infoHash := ""
	if hash := torrent.ExtractInfoHash(magnetLink); hash != "" {
		infoHash = hash
	}

	// Check quality filter
	if feed.Quality != "" {
		if !strings.Contains(strings.ToLower(title), strings.ToLower(feed.Quality)) {
			return fmt.Errorf("quality filter mismatch")
		}
	}

	// Check category filter
	if feed.Category != "" {
		if !strings.Contains(strings.ToLower(title), strings.ToLower(feed.Category)) {
			return fmt.Errorf("category filter mismatch")
		}
	}

	// Check if already downloaded
	exists, err := r.downloadManager.CheckIfAlreadyDownloaded(infoHash)
	if err != nil {
		return fmt.Errorf("failed to check if already downloaded: %w", err)
	}
	if exists {
		return fmt.Errorf("already downloaded")
	}

	// Create suggestion
	suggestion := &models.DownloadSuggestion{
		Title:      title,
		InfoHash:   infoHash,
		MagnetLink: magnetLink,
		Status:     "pending",
	}

	if err := r.db.GetDB().Where("info_hash = ?", infoHash).FirstOrCreate(suggestion).Error; err != nil {
		return fmt.Errorf("failed to create suggestion: %w", err)
	}

	log.Printf("[RSS] Created suggestion from feed '%s': %s", feed.Name, title)

	// Auto-approve if enabled and meets criteria
	if feed.AutoApprove {
		log.Printf("[RSS] Auto-approving suggestion from feed '%s': %s", feed.Name, title)

		// Create download task directly
		task := &models.DownloadTask{
			Title:        title,
			InfoHash:     infoHash,
			MagnetLink:   magnetLink,
			Status:       "pending",
			DownloadPath: "/mnt/media/Downloads",
		}

		if err := r.db.GetDB().Create(task).Error; err != nil {
			return fmt.Errorf("failed to create download task: %w", err)
		}

		// Update suggestion status
		r.db.GetDB().Model(suggestion).Update("status", "approved")
	}

	return nil
}
