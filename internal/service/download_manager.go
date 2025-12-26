package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/jellyfin"
	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

type TorrentClient interface {
	AddTorrent(magnetLink string, downloadDir string) (int, error)
	GetTorrentStatus(torrentID int) (map[string]interface{}, error)
	RemoveTorrent(torrentID int, deleteData bool) error
}

type DownloadManager struct {
	db              *db.Database
	torrentClient   TorrentClient
	torrentSearcher *torrent.TorrentSearcher
	updateCallback  func()
	serviceConfig   *models.ServiceConfig
	jellyfinClient  *jellyfin.Client
}

func NewDownloadManager(database *db.Database, cfg *models.ServiceConfig) (*DownloadManager, error) {
	var torrentClient TorrentClient

	switch cfg.TorrentClientType {
	case "transmission":
		client, err := torrent.NewTransmissionClient(
			cfg.TorrentClientHost,
			cfg.TorrentClientUser,
			cfg.TorrentClientPass,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create transmission client: %w", err)
		}
		torrentClient = client
	case "native", "":
		downloadDir := "/mnt/media/Downloads"
		client, err := torrent.NewNativeClient(downloadDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create native torrent client: %w", err)
		}
		torrentClient = client
	default:
		return nil, fmt.Errorf("unsupported torrent client type: %s", cfg.TorrentClientType)
	}

	torrentSearcher := torrent.NewTorrentSearcher()

	var sources []models.DownloadSource
	database.GetDB().Where("enabled = ?", true).Find(&sources)

	for _, source := range sources {
		var provider torrent.SearchProvider
		var err error

		switch source.Type {
		case "rarbg":
			provider, err = torrent.NewRARBGProvider(&source)
			if err != nil {
				log.Printf("[WARN] Failed to create RARBG provider: %v", err)
				continue
			}
		case "jackett":
			provider = torrent.NewJackettProvider(&source)
		default:
			log.Printf("[WARN] Unknown source type: %s", source.Type)
			continue
		}

		torrentSearcher.AddProvider(provider)
	}

	// Initialize Jellyfin client if configured
	var jellyfinClient *jellyfin.Client
	if cfg.JellyfinURL != "" {
		jellyfinClient = jellyfin.NewClient(cfg.JellyfinURL, cfg.JellyfinAPIKey)
		// Test connection
		if err := jellyfinClient.TestConnection(); err != nil {
			log.Printf("[JELLYFIN] Warning: Failed to connect to Jellyfin: %v", err)
		}
	}

	dm := &DownloadManager{
		db:              database,
		torrentClient:   torrentClient,
		torrentSearcher: torrentSearcher,
		serviceConfig:   cfg,
		jellyfinClient:  jellyfinClient,
	}

	// Recover orphaned downloads on startup
	dm.recoverOrphanedDownloads()

	return dm, nil
}

// recoverOrphanedDownloads resets downloads that were marked as "downloading" but have no active torrent
// This happens when the service restarts since the native torrent client keeps state in memory only
func (dm *DownloadManager) recoverOrphanedDownloads() {
	var tasks []models.DownloadTask
	if err := dm.db.GetDB().Where("status = ?", "downloading").Find(&tasks).Error; err != nil {
		log.Printf("[RECOVERY] Failed to query downloading tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		log.Println("[RECOVERY] No orphaned downloads to recover")
		return
	}

	log.Printf("[RECOVERY] Found %d downloads marked as 'downloading', checking for orphaned torrents", len(tasks))

	orphanedCount := 0
	for _, task := range tasks {
		// Try to get torrent status from client
		if task.TorrentID > 0 {
			_, err := dm.torrentClient.GetTorrentStatus(task.TorrentID)
			if err != nil {
				// Torrent doesn't exist in client - reset to pending
				log.Printf("[RECOVERY] Torrent not found for task %d (%s), resetting to pending", task.ID, task.Title)
				task.Status = "pending"
				task.Progress = 0.0
				task.TorrentID = 0
				task.StartedAt = nil
				if err := dm.db.GetDB().Save(&task).Error; err != nil {
					log.Printf("[RECOVERY] Failed to reset task %d: %v", task.ID, err)
				} else {
					orphanedCount++
				}
			} else {
				log.Printf("[RECOVERY] Task %d (%s) has active torrent, keeping status", task.ID, task.Title)
			}
		} else {
			// No torrent ID means it's definitely orphaned
			log.Printf("[RECOVERY] Task %d (%s) has no torrent ID, resetting to pending", task.ID, task.Title)
			task.Status = "pending"
			task.Progress = 0.0
			if err := dm.db.GetDB().Save(&task).Error; err != nil {
				log.Printf("[RECOVERY] Failed to reset task %d: %v", task.ID, err)
			} else {
				orphanedCount++
			}
		}
	}

	if orphanedCount > 0 {
		log.Printf("[RECOVERY] Reset %d orphaned downloads to pending status", orphanedCount)
	} else {
		log.Println("[RECOVERY] All downloads have active torrents, no recovery needed")
	}
}

func (dm *DownloadManager) SearchAndDownload(rule *models.DownloadRule) error {
	queries := dm.generateSearchQueries(rule.SearchQuery, rule.MediaType)
	log.Printf("[DOWNLOAD] Generated %d search queries", len(queries))

	maxResults := rule.MaxResults
	if maxResults == 0 {
		maxResults = 1
	}

	var currentDownloads int64
	if err := dm.db.GetDB().Model(&models.DownloadTask{}).
		Where("status IN (?)", []string{"pending", "downloading"}).
		Count(&currentDownloads).Error; err != nil {
		return fmt.Errorf("failed to count current downloads: %w", err)
	}

	maxConcurrent := dm.serviceConfig.MaxConcurrentDownloads
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}

	availableSlots := maxConcurrent - int(currentDownloads)
	if availableSlots <= 0 {
		log.Printf("[DOWNLOAD] Max concurrent downloads reached (%d/%d), skipping new downloads", currentDownloads, maxConcurrent)
		return nil
	}

	if maxResults > availableSlots {
		maxResults = availableSlots
		log.Printf("[DOWNLOAD] Limiting to %d downloads due to concurrent limit (%d/%d active)", maxResults, currentDownloads, maxConcurrent)
	}

	allResults := make([]models.SearchResult, 0)
	seenHashes := make(map[string]bool)

	for i, query := range queries {
		log.Printf("[DOWNLOAD] Search %d/%d: %s", i+1, len(queries), query)

		results, err := dm.torrentSearcher.Search(query, rule.MediaType, 100)
		if err != nil {
			log.Printf("[WARN] Search failed for '%s': %v", query, err)
			continue
		}

		log.Printf("[DOWNLOAD] Found %d results for query '%s'", len(results), query)

		for _, result := range results {
			if result.InfoHash != "" && !seenHashes[result.InfoHash] {
				allResults = append(allResults, result)
				seenHashes[result.InfoHash] = true
			} else if result.InfoHash == "" {
				allResults = append(allResults, result)
			}
		}
	}

	log.Printf("[DOWNLOAD] Total unique results across all queries: %d", len(allResults))

	filtered := dm.filterResults(allResults, rule)
	log.Printf("[DOWNLOAD] %d results after filtering", len(filtered))

	if len(filtered) == 0 {
		log.Printf("[DOWNLOAD] No results match criteria")
		return nil
	}

	dm.sortResults(filtered, rule.SortBy)

	uniqueFiltered := dm.deduplicateResults(filtered, rule.MaxResultsPerTitle)
	log.Printf("[DOWNLOAD] %d results after deduplication", len(uniqueFiltered))

	if !dm.isVPNActive() {
		log.Printf("[DOWNLOAD] VPN is not active, aborting downloads for security")
		return fmt.Errorf("VPN is not active, refusing to download")
	}

	downloaded := 0
	for _, result := range uniqueFiltered {
		if downloaded >= maxResults {
			break
		}

		exists, err := dm.CheckIfAlreadyDownloaded(result.InfoHash)
		if err != nil {
			log.Printf("[WARN] Failed to check if torrent exists: %v", err)
			continue
		}

		if exists {
			log.Printf("[DOWNLOAD] Skipping duplicate: %s", result.Title)
			continue
		}

		if err := dm.startDownload(rule, &result); err != nil {
			log.Printf("[ERROR] Failed to start download: %v", err)
			continue
		}

		downloaded++
		log.Printf("[DOWNLOAD] Downloaded %d/%d", downloaded, maxResults)
	}

	if downloaded == 0 {
		log.Printf("[DOWNLOAD] No new torrents to download")
	}

	return nil
}

func (dm *DownloadManager) generateSearchQueries(baseQuery string, mediaType string) []string {
	baseQuery = dm.expandQueryTemplates(baseQuery)

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

func (dm *DownloadManager) filterResults(results []models.SearchResult, rule *models.DownloadRule) []models.SearchResult {
	filtered := make([]models.SearchResult, 0)
	now := time.Now()

	for _, result := range results {
		if rule.MinSeeders > 0 && result.Seeders < rule.MinSeeders {
			log.Printf("[FILTER] %s: seeders too low (%d < %d)", result.Title, result.Seeders, rule.MinSeeders)
			continue
		}

		if rule.MaxSeeders > 0 && result.Seeders > rule.MaxSeeders {
			log.Printf("[FILTER] %s: seeders too high (%d > %d)", result.Title, result.Seeders, rule.MaxSeeders)
			continue
		}

		if rule.MinSize > 0 && result.Size < rule.MinSize {
			log.Printf("[FILTER] %s: size too small", result.Title)
			continue
		}

		if rule.MaxSize > 0 && result.Size > rule.MaxSize {
			log.Printf("[FILTER] %s: size too large", result.Title)
			continue
		}

		if rule.MinUploadAge > 0 && !result.UploadDate.IsZero() {
			age := int(now.Sub(result.UploadDate).Hours() / 24)
			if age < rule.MinUploadAge {
				log.Printf("[FILTER] %s: too new (%d days < %d days)", result.Title, age, rule.MinUploadAge)
				continue
			}
		}

		if rule.MaxUploadAge > 0 && !result.UploadDate.IsZero() {
			age := int(now.Sub(result.UploadDate).Hours() / 24)
			if age > rule.MaxUploadAge {
				log.Printf("[FILTER] %s: too old (%d days > %d days)", result.Title, age, rule.MaxUploadAge)
				continue
			}
		}

		if rule.Resolution != "" {
			if !dm.matchesResolution(result.Title, rule.Resolution) {
				log.Printf("[FILTER] %s: wrong resolution (want %s)", result.Title, rule.Resolution)
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

func (dm *DownloadManager) matchesResolution(title string, resolution string) bool {
	titleLower := strings.ToLower(title)
	resolutionLower := strings.ToLower(resolution)

	resolutionPatterns := map[string][]string{
		"2160p": {"2160p", "4k", "uhd"},
		"1080p": {"1080p"},
		"720p":  {"720p"},
		"480p":  {"480p"},
		"4k":    {"2160p", "4k", "uhd"},
		"uhd":   {"2160p", "4k", "uhd"},
	}

	patterns, ok := resolutionPatterns[resolutionLower]
	if !ok {
		patterns = []string{resolutionLower}
	}

	for _, pattern := range patterns {
		if strings.Contains(titleLower, pattern) {
			return true
		}
	}

	return false
}

func (dm *DownloadManager) sortResults(results []models.SearchResult, sortBy string) {
	switch sortBy {
	case "seeders":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Seeders > results[j].Seeders
		})
	case "size":
		sort.Slice(results, func(i, j int) bool {
			return results[i].Size > results[j].Size
		})
	case "upload_date":
		sort.Slice(results, func(i, j int) bool {
			return results[i].UploadDate.After(results[j].UploadDate)
		})
	case "balanced":
		sort.Slice(results, func(i, j int) bool {
			scoreI := dm.calculateBalancedScore(&results[i])
			scoreJ := dm.calculateBalancedScore(&results[j])
			return scoreI > scoreJ
		})
	default:
		sort.Slice(results, func(i, j int) bool {
			return results[i].Seeders > results[j].Seeders
		})
	}
}

func (dm *DownloadManager) calculateBalancedScore(result *models.SearchResult) float64 {
	seederScore := float64(result.Seeders) * 10.0

	ageScore := 0.0
	if !result.UploadDate.IsZero() {
		age := time.Since(result.UploadDate).Hours() / 24
		if age <= 7 {
			ageScore = 100.0
		} else if age <= 30 {
			ageScore = 75.0
		} else if age <= 90 {
			ageScore = 50.0
		} else {
			ageScore = 25.0
		}
	}

	return seederScore + ageScore
}

func (dm *DownloadManager) startDownload(rule *models.DownloadRule, result *models.SearchResult) error {
	if dm.torrentClient == nil {
		return fmt.Errorf("no torrent client configured")
	}

	log.Printf("[DOWNLOAD] Starting download: %s", result.Title)

	torrentID, err := dm.torrentClient.AddTorrent(result.MagnetLink, rule.DestinationPath)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %w", err)
	}

	task := &models.DownloadTask{
		RuleID:       rule.ID,
		SourceID:     result.SourceID,
		Title:        result.Title,
		InfoHash:     result.InfoHash,
		MagnetLink:   result.MagnetLink,
		TorrentURL:   result.TorrentURL,
		Size:         result.Size,
		Seeders:      result.Seeders,
		Leechers:     result.Leechers,
		Status:       "downloading",
		Progress:     0,
		DownloadPath: filepath.Join(rule.DestinationPath, result.Title),
	}

	startTime := time.Now()
	task.StartedAt = &startTime

	if err := dm.db.GetDB().Create(task).Error; err != nil {
		return fmt.Errorf("failed to create download task: %w", err)
	}

	dm.notifyUpdate()

	log.Printf("[DOWNLOAD] Download started with ID: %d (Transmission ID: %d)", task.ID, torrentID)
	return nil
}

func (dm *DownloadManager) CheckIfAlreadyDownloaded(infoHash string) (bool, error) {
	if infoHash == "" {
		return false, nil
	}

	// Check if already in download_tasks table (any status)
	var count int64
	err := dm.db.GetDB().Model(&models.DownloadTask{}).
		Where("info_hash = ?", infoHash).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check if in download_history table (previously downloaded and deleted)
	err = dm.db.GetDB().Model(&models.DownloadHistory{}).
		Where("info_hash = ?", infoHash).
		Count(&count).Error

	return count > 0, err
}

func (dm *DownloadManager) expandQueryTemplates(query string) string {
	now := time.Now()
	query = strings.ReplaceAll(query, "{year}", fmt.Sprintf("%d", now.Year()))
	query = strings.ReplaceAll(query, "{month}", fmt.Sprintf("%02d", now.Month()))
	query = strings.ReplaceAll(query, "{day}", fmt.Sprintf("%02d", now.Day()))
	return query
}

func (dm *DownloadManager) isVPNActive() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		log.Printf("[VPN] Failed to check IP: %v", err)
		return false
	}
	defer resp.Body.Close()

	var ip strings.Builder
	_, err = io.Copy(&ip, resp.Body)
	if err != nil {
		log.Printf("[VPN] Failed to read IP response: %v", err)
		return false
	}

	publicIP := strings.TrimSpace(ip.String())
	log.Printf("[VPN] Current public IP: %s", publicIP)

	if strings.HasPrefix(publicIP, "192.168.") || strings.HasPrefix(publicIP, "10.") || strings.HasPrefix(publicIP, "172.") {
		log.Printf("[VPN] WARNING: Using local IP address, VPN is NOT active!")
		return false
	}

	log.Printf("[VPN] VPN appears to be active")
	return true
}

func (dm *DownloadManager) UpdateTaskProgress() error {
	var tasks []models.DownloadTask
	err := dm.db.GetDB().Where("status = ?", "downloading").Find(&tasks).Error
	if err != nil {
		return err
	}

	for _, task := range tasks {
		log.Printf("[DOWNLOAD] Updating progress for: %s", task.Title)
	}

	return nil
}

func (dm *DownloadManager) deduplicateResults(results []models.SearchResult, maxPerTitle int) []models.SearchResult {
	if maxPerTitle <= 0 {
		maxPerTitle = 1
	}

	titleCounts := make(map[string]int)
	deduplicated := make([]models.SearchResult, 0)

	for _, result := range results {
		normalizedTitle := dm.normalizeTitle(result.Title)
		count := titleCounts[normalizedTitle]

		if count < maxPerTitle {
			deduplicated = append(deduplicated, result)
			titleCounts[normalizedTitle] = count + 1
		} else {
			log.Printf("[DEDUP] Skipping duplicate title: %s (normalized: %s)", result.Title, normalizedTitle)
		}
	}

	return deduplicated
}

func (dm *DownloadManager) normalizeTitle(title string) string {
	title = strings.ToLower(title)

	yearPattern := `\b(19|20)\d{2}\b`
	title = strings.TrimSpace(strings.Split(title, " (")[0])

	replacements := []string{
		".", " ",
		"-", " ",
		"_", " ",
		"  ", " ",
	}
	for i := 0; i < len(replacements); i += 2 {
		title = strings.ReplaceAll(title, replacements[i], replacements[i+1])
	}

	qualityTerms := []string{
		"1080p", "720p", "2160p", "4k", "uhd", "hdr", "bluray", "brrip", "webrip", "web-dl",
		"x264", "x265", "h264", "h265", "hevc", "10bit", "dts", "ac3", "aac", "opus",
		"yify", "rarbg", "galaxyrg", "yts", "proper", "repack", "extended", "unrated",
	}

	for _, term := range qualityTerms {
		title = strings.ReplaceAll(title, term, "")
	}

	fields := strings.Fields(title)
	normalized := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) > 0 && !strings.ContainsAny(field, "[](){}") {
			if matched, _ := regexp.MatchString(yearPattern, field); !matched {
				normalized = append(normalized, field)
			}
		}
	}

	return strings.TrimSpace(strings.Join(normalized, " "))
}

func (dm *DownloadManager) CancelTask(taskID uint) error {
	var task models.DownloadTask
	if err := dm.db.GetDB().First(&task, taskID).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.Status == "completed" {
		return fmt.Errorf("cannot cancel completed task")
	}

	if task.Status == "cancelled" {
		return fmt.Errorf("task already cancelled")
	}

	if task.TorrentID > 0 && dm.torrentClient != nil {
		log.Printf("[DOWNLOAD] Removing torrent from client: %s (ID: %d)", task.Title, task.TorrentID)
		if err := dm.torrentClient.RemoveTorrent(task.TorrentID, false); err != nil {
			log.Printf("[WARN] Failed to remove torrent from client: %v", err)
		}
	}

	task.Status = "cancelled"
	if err := dm.db.GetDB().Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	dm.notifyUpdate()

	log.Printf("[DOWNLOAD] Task cancelled: %s", task.Title)
	return nil
}

func (dm *DownloadManager) RestartTask(taskID uint) error {
	var task models.DownloadTask
	if err := dm.db.GetDB().First(&task, taskID).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.Status == "pending" {
		return fmt.Errorf("task is already pending")
	}

	if task.Status == "completed" {
		return fmt.Errorf("cannot restart completed task")
	}

	// If there's an active torrent, remove it first
	if task.TorrentID > 0 && dm.torrentClient != nil {
		log.Printf("[DOWNLOAD] Removing torrent from client before restart: %s (ID: %d)", task.Title, task.TorrentID)
		if err := dm.torrentClient.RemoveTorrent(task.TorrentID, false); err != nil {
			log.Printf("[WARN] Failed to remove torrent from client: %v", err)
		}
	}

	// Reset task to pending state
	task.Status = "pending"
	task.Progress = 0.0
	task.TorrentID = 0
	if err := dm.db.GetDB().Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	dm.notifyUpdate()

	log.Printf("[DOWNLOAD] Task restarted and reset to pending: %s", task.Title)
	return nil
}

func (dm *DownloadManager) SearchWithoutDownload(rule *models.DownloadRule) ([]models.SearchResult, error) {
	queries := dm.generateSearchQueries(rule.SearchQuery, rule.MediaType)
	log.Printf("[SEARCH] Generated %d search queries", len(queries))

	allResults := make([]models.SearchResult, 0)
	seenHashes := make(map[string]bool)

	for i, query := range queries {
		log.Printf("[SEARCH] Search %d/%d: %s", i+1, len(queries), query)

		results, err := dm.torrentSearcher.Search(query, rule.MediaType, 100)
		if err != nil {
			log.Printf("[WARN] Search failed for '%s': %v", query, err)
			continue
		}

		log.Printf("[SEARCH] Found %d results for query '%s'", len(results), query)

		for _, result := range results {
			if result.InfoHash != "" && !seenHashes[result.InfoHash] {
				allResults = append(allResults, result)
				seenHashes[result.InfoHash] = true
			} else if result.InfoHash == "" {
				allResults = append(allResults, result)
			}
		}
	}

	log.Printf("[SEARCH] Total unique results across all queries: %d", len(allResults))

	filtered := dm.filterResults(allResults, rule)
	log.Printf("[SEARCH] %d results after filtering", len(filtered))

	dm.sortResults(filtered, rule.SortBy)

	maxResultsPerTitle := rule.MaxResultsPerTitle
	if maxResultsPerTitle == 0 {
		maxResultsPerTitle = 1
	}
	uniqueFiltered := dm.deduplicateResults(filtered, maxResultsPerTitle)
	log.Printf("[SEARCH] %d results after deduplication", len(uniqueFiltered))

	maxResults := rule.MaxResults
	if maxResults > 0 && len(uniqueFiltered) > maxResults {
		uniqueFiltered = uniqueFiltered[:maxResults]
	}

	return uniqueFiltered, nil
}

func (dm *DownloadManager) SetUpdateCallback(callback func()) {
	dm.updateCallback = callback
}

func (dm *DownloadManager) notifyUpdate() {
	if dm.updateCallback != nil {
		dm.updateCallback()
	}
}

func (dm *DownloadManager) UpdateAllProgress() {
	var tasks []models.DownloadTask
	dm.db.GetDB().Where("status = ?", "downloading").Find(&tasks)

	if len(tasks) == 0 {
		return
	}

	log.Printf("[MONITOR] Updating progress for %d active downloads", len(tasks))

	hasUpdates := false

	for _, task := range tasks {
		status, err := dm.torrentClient.GetTorrentStatus(task.TorrentID)
		if err != nil {
			// If torrent not found, it's orphaned (service restart, manual removal, etc.)
			// Reset to pending so it will be restarted automatically
			log.Printf("[MONITOR] Torrent not found for task %d (%s), resetting to pending: %v", task.ID, task.Title, err)
			task.Status = "pending"
			task.Progress = 0.0
			task.TorrentID = 0
			task.StartedAt = nil
			if err := dm.db.GetDB().Save(&task).Error; err != nil {
				log.Printf("[MONITOR] Failed to reset orphaned task %d: %v", task.ID, err)
			} else {
				hasUpdates = true
			}
			continue
		}

		percentDone := status["percentDone"].(float64) * 100

		task.Progress = percentDone
		if percentDone >= 100 {
			task.Status = "completed"
			now := time.Now()
			task.CompletedAt = &now
			log.Printf("[MONITOR] Download completed: %s", task.Title)

			// Record in download history
			go dm.recordDownloadHistory(&task)

			// Post-process for Plex
			go dm.postProcessForPlex(&task)
		}

		if err := dm.db.GetDB().Save(&task).Error; err != nil {
			log.Printf("[MONITOR] Failed to update task %d: %v", task.ID, err)
		} else {
			if percentDone > 0 {
				log.Printf("[MONITOR] %s: %.2f%%", task.Title, percentDone)
			}
			hasUpdates = true
		}
	}

	// Notify SSE clients of progress updates
	if hasUpdates {
		dm.notifyUpdate()
	}
}

// ProcessPendingDownloads checks for pending download tasks and starts them
func (dm *DownloadManager) ProcessPendingDownloads() {
	var tasks []models.DownloadTask
	if err := dm.db.GetDB().Where("status = ?", "pending").Order("created_at ASC").Find(&tasks).Error; err != nil {
		log.Printf("[DOWNLOAD] Failed to query pending tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	// Check concurrent download limit
	var activeCount int64
	if err := dm.db.GetDB().Model(&models.DownloadTask{}).
		Where("status = ?", "downloading").
		Count(&activeCount).Error; err != nil {
		log.Printf("[DOWNLOAD] Failed to count active downloads: %v", err)
		return
	}

	maxConcurrent := dm.serviceConfig.MaxConcurrentDownloads
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}

	availableSlots := maxConcurrent - int(activeCount)
	if availableSlots <= 0 {
		log.Printf("[DOWNLOAD] Max concurrent downloads reached (%d/%d), pending downloads will wait", activeCount, maxConcurrent)
		return
	}

	log.Printf("[DOWNLOAD] Processing %d pending downloads (%d slots available)", len(tasks), availableSlots)

	started := 0
	for _, task := range tasks {
		if started >= availableSlots {
			break
		}

		log.Printf("[DOWNLOAD] Starting pending download: %s", task.Title)

		// Add torrent to client
		torrentID, err := dm.torrentClient.AddTorrent(task.MagnetLink, "/mnt/media/Downloads")
		if err != nil {
			log.Printf("[DOWNLOAD] Failed to add torrent for task %d: %v", task.ID, err)
			// Mark as failed
			task.Status = "failed"
			task.Error = err.Error()
			dm.db.GetDB().Save(&task)
			continue
		}

		// Update task status
		now := time.Now()
		task.Status = "downloading"
		task.TorrentID = torrentID
		task.StartedAt = &now

		if err := dm.db.GetDB().Save(&task).Error; err != nil {
			log.Printf("[DOWNLOAD] Failed to update task %d: %v", task.ID, err)
			continue
		}

		log.Printf("[DOWNLOAD] Started download task %d (Torrent ID: %d)", task.ID, torrentID)
		started++
	}

	if started > 0 {
		log.Printf("[DOWNLOAD] Successfully started %d pending downloads", started)
		dm.notifyUpdate()
	}
}

func (dm *DownloadManager) postProcessForPlex(task *models.DownloadTask) {
	log.Printf("[POSTPROCESS] Starting post-processing for: %s", task.Title)

	// Extract movie name and year from title
	movieName, year := dm.extractMovieNameAndYear(task.Title)
	if movieName == "" {
		log.Printf("[POSTPROCESS] Could not extract movie name from: %s", task.Title)
		return
	}

	// Create Plex-compatible folder name
	var folderName string
	if year != "" {
		folderName = fmt.Sprintf("%s (%s)", movieName, year)
	} else {
		folderName = movieName
	}

	// Get source directory (where torrent downloaded to)
	sourceDir := filepath.Join("/mnt/media/Downloads", task.Title)

	// Create destination directory
	destDir := filepath.Join("/mnt/media/Movies", folderName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Printf("[POSTPROCESS] Failed to create directory %s: %v", destDir, err)
		return
	}

	// Find video file in source directory
	videoFiles, err := dm.findVideoFiles(sourceDir)
	if err != nil {
		log.Printf("[POSTPROCESS] Failed to find video files in %s: %v", sourceDir, err)
		return
	}

	if len(videoFiles) == 0 {
		log.Printf("[POSTPROCESS] No video files found in %s", sourceDir)
		return
	}

	// Move/copy main video file and related files
	for _, videoFile := range videoFiles {
		ext := filepath.Ext(videoFile)
		var destFile string
		if year != "" {
			destFile = filepath.Join(destDir, fmt.Sprintf("%s (%s)%s", movieName, year, ext))
		} else {
			destFile = filepath.Join(destDir, fmt.Sprintf("%s%s", movieName, ext))
		}

		// Move the file
		if err := os.Rename(videoFile, destFile); err != nil {
			log.Printf("[POSTPROCESS] Failed to move %s to %s: %v", videoFile, destFile, err)
			continue
		}

		log.Printf("[POSTPROCESS] Moved: %s -> %s", videoFile, destFile)

		// Move subtitle files if they exist
		dm.moveSubtitles(videoFile, destFile)
	}

	// Update task with final location
	task.DownloadPath = destDir
	dm.db.GetDB().Save(task)

	log.Printf("[POSTPROCESS] Completed post-processing for: %s", task.Title)

	// Trigger Jellyfin library refresh (in background to avoid blocking)
	if dm.jellyfinClient != nil {
		go dm.refreshJellyfinMetadata(destDir)
	}
}

// refreshJellyfinMetadata triggers metadata refresh for a newly added movie
func (dm *DownloadManager) refreshJellyfinMetadata(movieDir string) {
	// Wait for Jellyfin's realtime monitor to detect the new file (2-5 seconds)
	time.Sleep(5 * time.Second)

	log.Printf("[JELLYFIN] Triggering full library refresh for new movie in: %s", movieDir)
	if err := dm.jellyfinClient.RefreshLibrary(); err != nil {
		log.Printf("[JELLYFIN] Failed to refresh library: %v", err)
		return
	}

	// Wait for scan to complete and find the new item
	time.Sleep(10 * time.Second)

	// Try to find the item by path and refresh its metadata to fetch missing posters
	videoFiles, _ := dm.findVideoFiles(movieDir)
	if len(videoFiles) > 0 {
		videoPath := videoFiles[0]
		log.Printf("[JELLYFIN] Searching for item with path: %s", videoPath)

		item, err := dm.jellyfinClient.SearchByPath(videoPath)
		if err != nil {
			log.Printf("[JELLYFIN] Could not find item by path: %v", err)
			return
		}

		// Check if item is missing poster
		if _, hasPoster := item.ImageTags["Primary"]; !hasPoster {
			log.Printf("[JELLYFIN] Item '%s' missing poster, triggering metadata refresh", item.Name)
			if err := dm.jellyfinClient.RefreshItemMetadata(item.Id, true); err != nil {
				log.Printf("[JELLYFIN] Failed to refresh item metadata: %v", err)
			} else {
				log.Printf("[JELLYFIN] Successfully triggered poster download for '%s'", item.Name)
			}
		} else {
			log.Printf("[JELLYFIN] Item '%s' already has poster", item.Name)
		}
	}
}

func (dm *DownloadManager) extractMovieNameAndYear(title string) (string, string) {
	// Extract year first - look for (YYYY) or [YYYY]
	yearRegex := regexp.MustCompile(`[\[\(](\d{4})[\]\)]`)
	yearMatches := yearRegex.FindStringSubmatch(title)
	var year string
	if len(yearMatches) > 1 {
		year = yearMatches[1]
		// Remove ALL occurrences of year with brackets/parens from title
		title = yearRegex.ReplaceAllString(title, "")
	}

	// Remove file extensions that might be in the title
	title = regexp.MustCompile(`(?i)\.(mkv|mp4|avi|mov|m4v|wmv|flv|webm)$`).ReplaceAllString(title, "")

	// Remove resolution indicators (with various separators)
	title = regexp.MustCompile(`(?i)[\[\(\s]*(2160|1080|720|480|360)p?[\]\)\s]*`).ReplaceAllString(title, " ")

	// Remove common quality/release indicators
	title = regexp.MustCompile(`(?i)\[YTS\.(MX|LT|AG)\]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)\[TGx\]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)\[?(BONE|YIFY|RARBG|PSA|ETRG|SPARKS|EVO|ION10|CMRG)\]?`).ReplaceAllString(title, "")

	// Remove source/quality indicators
	title = regexp.MustCompile(`(?i)\[?(BluRay|BRRip|BDRip|WEBRip|WEB-DL|HDTV|HDRip|BrRip|TS|CAM|HC|HDRIP|DVDRip|DVDSCR)\]?`).ReplaceAllString(title, "")

	// Remove codec and audio format info
	title = regexp.MustCompile(`(?i)(x264|x265|h\.?264|h\.?265|HEVC|AVC|10bit|8bit)`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)(AAC|AC3|DD|DDP|EAC3|TrueHD|DTS|FLAC|MP3|Atmos)`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`(?i)[\[\(]?(5\.1|7\.1|2\.0|Stereo)[\]\)]?`).ReplaceAllString(title, "")

	// Remove scene/release group tags
	title = regexp.MustCompile(`(?i)(EN-RGB|EniaHD|FLUX|GalaxyRG|DEFLATE|STUTTERSHIT|VARYG)`).ReplaceAllString(title, "")

	// Remove file size indicators
	title = regexp.MustCompile(`(?i)\d+(\.\d+)?\s?(GB|MB|GiB|MiB)`).ReplaceAllString(title, "")

	// Remove language codes and subtitles info
	title = regexp.MustCompile(`(?i)[\[\(]?(Multi|Dual|Eng|English|Sub|Subs|Subtitle)[\]\)]?`).ReplaceAllString(title, "")

	// Remove common abbreviations that appear in weird formats
	title = regexp.MustCompile(`(?i)\s+(MB|DD|HC)\s+`).ReplaceAllString(title, " ")

	// Remove empty brackets and parentheses (including nested ones)
	for i := 0; i < 3; i++ {
		title = regexp.MustCompile(`\[\s*\]`).ReplaceAllString(title, "")
		title = regexp.MustCompile(`\(\s*\)`).ReplaceAllString(title, "")
		title = regexp.MustCompile(`\{\s*\}`).ReplaceAllString(title, "")
	}

	// Remove standalone brackets/parens with only special chars
	title = regexp.MustCompile(`[\[\(][^\w\s]*[\]\)]`).ReplaceAllString(title, "")

	// Remove special characters that are commonly malformed
	title = regexp.MustCompile(`\s*[p\]]+\s*\[+\s*\[+\s*`).ReplaceAllString(title, " ") // Fix "p] [] [" pattern

	// Clean up dots, underscores, and dashes used as separators
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = regexp.MustCompile(`\s*-\s*-+\s*`).ReplaceAllString(title, " ") // Remove multiple dashes

	// Remove multiple spaces
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)

	// Remove trailing/leading brackets, parentheses, and special chars
	title = regexp.MustCompile(`^[\[\(\{]+|[\]\)\}]+$`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`^[\s\-_,]+|[\s\-_,]+$`).ReplaceAllString(title, "")
	title = strings.TrimSpace(title)

	return title, year
}

func (dm *DownloadManager) findVideoFiles(dir string) ([]string, error) {
	var videoFiles []string
	videoExtensions := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
		".m4v": true, ".wmv": true, ".flv": true, ".webm": true,
	}

	// Check if dir is actually a file
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// It's a single file
		ext := strings.ToLower(filepath.Ext(dir))
		if videoExtensions[ext] {
			return []string{dir}, nil
		}
		return nil, nil
	}

	// It's a directory, scan for video files
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip .part files (incomplete downloads)
		if strings.HasSuffix(strings.ToLower(path), ".part") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if videoExtensions[ext] {
			// Skip sample files
			if !strings.Contains(strings.ToLower(path), "sample") {
				videoFiles = append(videoFiles, path)
			}
		}

		return nil
	})

	return videoFiles, err
}

func (dm *DownloadManager) moveSubtitles(videoFile, destVideoFile string) {
	// Look for subtitle files with the same base name
	baseName := strings.TrimSuffix(videoFile, filepath.Ext(videoFile))
	destBaseName := strings.TrimSuffix(destVideoFile, filepath.Ext(destVideoFile))

	subtitleExtensions := []string{".srt", ".sub", ".idx", ".ssa", ".ass"}

	for _, ext := range subtitleExtensions {
		subFile := baseName + ext
		if _, err := os.Stat(subFile); err == nil {
			destSubFile := destBaseName + ext
			if err := os.Rename(subFile, destSubFile); err != nil {
				log.Printf("[POSTPROCESS] Failed to move subtitle %s: %v", subFile, err)
			} else {
				log.Printf("[POSTPROCESS] Moved subtitle: %s -> %s", subFile, destSubFile)
			}
		}
	}
}

func (dm *DownloadManager) recordDownloadHistory(task *models.DownloadTask) error {
	if task.InfoHash == "" {
		return nil
	}

	now := time.Now()
	history := models.DownloadHistory{
		InfoHash:     task.InfoHash,
		Title:        task.Title,
		Size:         task.Size,
		DownloadedAt: now,
		DeletedAt:    now, // Set to current time when first recorded
		Reason:       "completed",
	}

	// Use FirstOrCreate to avoid duplicates
	result := dm.db.GetDB().Where("info_hash = ?", task.InfoHash).FirstOrCreate(&history)
	if result.Error != nil {
		log.Printf("[HISTORY] Failed to record download history for %s: %v", task.Title, result.Error)
		return result.Error
	}

	log.Printf("[HISTORY] Recorded download: %s (hash: %s)", task.Title, task.InfoHash)
	return nil
}

// GetAllTasks retrieves all download tasks from the database
func (dm *DownloadManager) GetAllTasks() []models.DownloadTask {
	var tasks []models.DownloadTask
	result := dm.db.GetDB().Order("created_at DESC").Find(&tasks)
	if result.Error != nil {
		log.Printf("[DM] Error fetching tasks: %v", result.Error)
		return []models.DownloadTask{}
	}
	return tasks
}

// DeleteTask deletes a download task from the database
func (dm *DownloadManager) DeleteTask(taskID uint) error {
	var task models.DownloadTask

	// First, find the task
	result := dm.db.GetDB().First(&task, taskID)
	if result.Error != nil {
		return fmt.Errorf("task not found: %w", result.Error)
	}

	// If the task is currently downloading, try to cancel it first
	if task.Status == "downloading" && task.TorrentID > 0 && dm.torrentClient != nil {
		log.Printf("[DM] Removing active torrent for task %d before deletion", taskID)
		if err := dm.torrentClient.RemoveTorrent(task.TorrentID, false); err != nil {
			log.Printf("[DM] Warning: Failed to remove torrent: %v", err)
		}
	}

	// Delete the task from the database
	result = dm.db.GetDB().Delete(&task)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}

	log.Printf("[DM] Deleted task %d: %s", taskID, task.Title)
	dm.notifyUpdate()
	return nil
}

// DeleteTasksByStatus deletes all download tasks with the given status
func (dm *DownloadManager) DeleteTasksByStatus(status string) (int, error) {
	var tasks []models.DownloadTask

	// Find all tasks with the specified status
	result := dm.db.GetDB().Where("status = ?", status).Find(&tasks)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to query tasks: %w", result.Error)
	}

	count := 0
	for _, task := range tasks {
		if err := dm.DeleteTask(task.ID); err != nil {
			log.Printf("[DM] Failed to delete task %d: %v", task.ID, err)
			continue
		}
		count++
	}

	log.Printf("[DM] Deleted %d tasks with status '%s'", count, status)
	return count, nil
}

// ReprocessCompletedDownloads runs post-processing on all completed tasks that haven't been organized yet
func (dm *DownloadManager) ReprocessCompletedDownloads() (int, error) {
	// Find all video files in Downloads directory
	downloadsDir := "/mnt/media/Downloads"

	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read downloads directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		fullPath := filepath.Join(downloadsDir, entry.Name())

		// Find video files
		videoFiles, err := dm.findVideoFiles(fullPath)
		if err != nil || len(videoFiles) == 0 {
			continue
		}

		// Create a mock task with the folder/file name as title
		mockTask := &models.DownloadTask{
			Title: entry.Name(),
		}

		log.Printf("[REPROCESS] Processing: %s (%d video files)", entry.Name(), len(videoFiles))
		dm.postProcessForPlex(mockTask)
		count++
	}

	log.Printf("[REPROCESS] Completed reprocessing %d downloads", count)
	return count, nil
}

// ReprocessLibraryFiles renames existing movies in /mnt/media/Movies using the enhanced filename cleaning
func (dm *DownloadManager) ReprocessLibraryFiles() (int, error) {
	moviesDir := "/mnt/media/Movies"

	entries, err := os.ReadDir(moviesDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read movies directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		oldFolderPath := filepath.Join(moviesDir, entry.Name())

		// Find video files in this folder
		videoFiles, err := dm.findVideoFiles(oldFolderPath)
		if err != nil || len(videoFiles) == 0 {
			log.Printf("[LIBRARY-REPROCESS] Skipping %s: no video files", entry.Name())
			continue
		}

		// Extract clean movie name and year from folder name
		movieName, year := dm.extractMovieNameAndYear(entry.Name())
		if movieName == "" {
			log.Printf("[LIBRARY-REPROCESS] Skipping %s: could not extract movie name", entry.Name())
			continue
		}

		// Build new folder name
		var newFolderName string
		if year != "" {
			newFolderName = fmt.Sprintf("%s (%s)", movieName, year)
		} else {
			newFolderName = movieName
		}

		// Skip if name hasn't changed
		if newFolderName == entry.Name() {
			log.Printf("[LIBRARY-REPROCESS] Skipping %s: name already clean", entry.Name())
			continue
		}

		newFolderPath := filepath.Join(moviesDir, newFolderName)

		// Check if target folder already exists
		if _, err := os.Stat(newFolderPath); err == nil {
			log.Printf("[LIBRARY-REPROCESS] Skipping %s: target folder %s already exists", entry.Name(), newFolderName)
			continue
		}

		// Rename the folder
		log.Printf("[LIBRARY-REPROCESS] Renaming: %s -> %s", entry.Name(), newFolderName)
		if err := os.Rename(oldFolderPath, newFolderPath); err != nil {
			log.Printf("[LIBRARY-REPROCESS] Failed to rename %s: %v", entry.Name(), err)
			continue
		}

		// Rename video files inside the folder
		newVideoFiles, _ := dm.findVideoFiles(newFolderPath)
		for _, videoFile := range newVideoFiles {
			oldBaseName := filepath.Base(videoFile)
			ext := filepath.Ext(videoFile)

			var newFileName string
			if year != "" {
				newFileName = fmt.Sprintf("%s (%s)%s", movieName, year, ext)
			} else {
				newFileName = fmt.Sprintf("%s%s", movieName, ext)
			}

			// Skip if filename is already correct
			if oldBaseName == newFileName {
				continue
			}

			newFilePath := filepath.Join(newFolderPath, newFileName)
			log.Printf("[LIBRARY-REPROCESS] Renaming file: %s -> %s", oldBaseName, newFileName)

			if err := os.Rename(videoFile, newFilePath); err != nil {
				log.Printf("[LIBRARY-REPROCESS] Failed to rename file %s: %v", oldBaseName, err)
			}
		}

		count++
	}

	log.Printf("[LIBRARY-REPROCESS] Completed reprocessing %d library folders", count)

	// Trigger Jellyfin library refresh if available
	if dm.jellyfinClient != nil {
		log.Println("[LIBRARY-REPROCESS] Triggering Jellyfin library refresh")
		if err := dm.jellyfinClient.RefreshLibrary(); err != nil {
			log.Printf("[LIBRARY-REPROCESS] Failed to refresh Jellyfin: %v", err)
		}
	}

	return count, nil
}

// GetJellyfinClient returns the Jellyfin client (may be nil if not configured)
func (dm *DownloadManager) GetJellyfinClient() *jellyfin.Client {
	return dm.jellyfinClient
}
