package service

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
	"gorm.io/gorm"
)

type SuggestionService struct {
	db              *db.Database
	downloadManager *DownloadManager
	tmdbService     *TMDbService
}

func NewSuggestionService(database *db.Database, dm *DownloadManager, tmdb *TMDbService) *SuggestionService {
	return &SuggestionService{
		db:              database,
		downloadManager: dm,
		tmdbService:     tmdb,
	}
}

// CalculateTorrentQualityScore calculates a quality score for a torrent result
// Balances file size, seeders, and quality markers to pick the best version
func (s *SuggestionService) CalculateTorrentQualityScore(title string, size int64, seeders int) float64 {
	score := 0.0

	// 1. Seeder score (40% weight) - more seeders is better
	seederScore := float64(seeders) / 100.0
	if seederScore > 100.0 {
		seederScore = 100.0
	}
	score += seederScore * 0.4

	// 2. File size score (35% weight) - optimal range is 2-4 GB for 1080p
	sizeGB := float64(size) / (1024 * 1024 * 1024)
	sizeScore := 0.0
	if sizeGB >= 2.0 && sizeGB <= 4.0 {
		sizeScore = 100.0 // Perfect size
	} else if sizeGB >= 1.5 && sizeGB < 2.0 {
		sizeScore = 85.0 // Acceptable smaller size
	} else if sizeGB > 4.0 && sizeGB <= 6.0 {
		sizeScore = 85.0 // Acceptable larger size
	} else if sizeGB > 6.0 && sizeGB <= 8.0 {
		sizeScore = 60.0 // Too large but might be high quality
	} else if sizeGB < 1.5 {
		sizeScore = 50.0 // Too small, likely lower quality
	} else {
		sizeScore = 30.0 // Way too large (>8GB)
	}
	score += sizeScore * 0.35

	// 3. Quality markers (25% weight)
	qualityScore := 50.0 // Base score
	titleLower := strings.ToLower(title)

	// Source quality
	if strings.Contains(titleLower, "web-dl") {
		qualityScore += 20.0 // Best quality
	} else if strings.Contains(titleLower, "webrip") {
		qualityScore += 15.0 // Good quality
	} else if strings.Contains(titleLower, "bluray") || strings.Contains(titleLower, "brrip") {
		qualityScore += 18.0 // Excellent quality
	} else if strings.Contains(titleLower, "hdrip") {
		qualityScore += 10.0 // Decent quality
	} else if strings.Contains(titleLower, "ts") || strings.Contains(titleLower, "cam") {
		qualityScore -= 30.0 // Poor quality
	}

	// Codec efficiency (x265/HEVC is more efficient)
	if strings.Contains(titleLower, "x265") || strings.Contains(titleLower, "hevc") {
		qualityScore += 15.0 // Bonus for efficient codec
	}

	// Audio quality
	if strings.Contains(titleLower, "atmos") || strings.Contains(titleLower, "ddp5.1") || strings.Contains(titleLower, "dd5.1") {
		qualityScore += 10.0
	} else if strings.Contains(titleLower, "5.1") {
		qualityScore += 5.0
	}

	// Cap quality score at 100
	if qualityScore > 100.0 {
		qualityScore = 100.0
	}
	score += qualityScore * 0.25

	return score
}

// CalculateQualityScore calculates a quality score for a DownloadSuggestion.
// Returns a value on a 0-10 scale balancing seeders, size, and quality markers.
func (s *SuggestionService) CalculateQualityScore(suggestion *models.DownloadSuggestion) float64 {
	score := 1.5

	// Seeders: logarithmic scale so huge numbers don't dominate
	score += math.Log10(float64(suggestion.Seeders + 1))

	// Size: prefer 2-20 GB; penalize very small or very large
	sizeGB := float64(suggestion.Size) / (1024 * 1024 * 1024)
	if sizeGB >= 2.0 && sizeGB <= 20.0 {
		score += 1.0
	} else {
		score -= 1.0
	}

	// Quality markers
	qualityLower := strings.ToLower(suggestion.Quality)
	if strings.Contains(qualityLower, "bluray") || strings.Contains(qualityLower, "brrip") {
		score += 0.6
	} else if strings.Contains(qualityLower, "web-dl") || strings.Contains(qualityLower, "webrip") {
		score += 0.45
	}

	return score
}

func (s *SuggestionService) GenerateSuggestions(rule *models.DownloadRule, skipAlreadyDownloaded bool) (int, error) {
	log.Printf("[SUGGESTIONS] Generating suggestions for rule: %s (skipAlreadyDownloaded=%v)", rule.Name, skipAlreadyDownloaded)
	startTime := time.Now()

	searchResult, err := s.downloadManager.SearchWithoutDownload(rule)
	if err != nil {
		return 0, fmt.Errorf("search failed: %w", err)
	}

	results := searchResult.Results
	log.Printf("[SUGGESTIONS] Found %d results to process", len(results))

	// First pass: fetch TMDB data for all results and group by TMDB ID
	type candidateSuggestion struct {
		result       models.SearchResult
		tmdbID       int
		poster       string
		qualityScore float64
	}

	// Parallelize TMDb fetches using worker pool
	tmdbStart := time.Now()
	numWorkers := 10 // 10 concurrent TMDb API requests (TMDb allows 50 req/s)
	if len(results) < numWorkers {
		numWorkers = len(results)
	}

	type tmdbResult struct {
		index     int
		posterURL string
		tmdbID    int
		err       error
	}

	resultsChan := make(chan int, len(results))
	tmdbChan := make(chan tmdbResult, len(results))

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for idx := range resultsChan {
				if s.tmdbService != nil {
					posterURL, tmdbID, err := s.tmdbService.FetchPosterForTask(results[idx].Title)
					tmdbChan <- tmdbResult{index: idx, posterURL: posterURL, tmdbID: tmdbID, err: err}
				} else {
					tmdbChan <- tmdbResult{index: idx, err: fmt.Errorf("no tmdb service")}
				}
			}
		}()
	}

	// Send work to workers
	for i := range results {
		resultsChan <- i
	}
	close(resultsChan)

	// Collect results
	tmdbResults := make(map[int]tmdbResult)
	for range results {
		tr := <-tmdbChan
		tmdbResults[tr.index] = tr
		if tr.err != nil && tr.err.Error() != "no tmdb service" {
			log.Printf("[SUGGESTIONS] Failed to fetch TMDb for %s: %v", results[tr.index].Title, tr.err)
		}
	}
	close(tmdbChan)

	tmdbDuration := time.Since(tmdbStart)
	log.Printf("[SUGGESTIONS] TMDb metadata fetched for %d results (%.2fs)", len(results), tmdbDuration.Seconds())

	// Build candidates with TMDb data
	candidatesByTMDbID := make(map[int][]*candidateSuggestion)
	candidatesWithoutTMDb := []*candidateSuggestion{}

	for i, result := range results {
		candidate := &candidateSuggestion{
			result:       result,
			qualityScore: s.CalculateTorrentQualityScore(result.Title, result.Size, result.Seeders),
		}

		if tr, ok := tmdbResults[i]; ok && tr.err == nil {
			candidate.tmdbID = tr.tmdbID
			candidate.poster = tr.posterURL
		}

		if candidate.tmdbID > 0 {
			candidatesByTMDbID[candidate.tmdbID] = append(candidatesByTMDbID[candidate.tmdbID], candidate)
		} else {
			candidatesWithoutTMDb = append(candidatesWithoutTMDb, candidate)
		}
	}

	// Second pass: deduplicate by TMDB ID, keeping the one with highest quality score
	uniqueCandidates := []*candidateSuggestion{}

	for tmdbID, candidates := range candidatesByTMDbID {
		if len(candidates) == 1 {
			uniqueCandidates = append(uniqueCandidates, candidates[0])
		} else {
			// Multiple torrents for the same movie - pick the one with best quality score
			best := candidates[0]
			for _, c := range candidates[1:] {
				if c.qualityScore > best.qualityScore {
					best = c
				}
			}
			sizeGB := float64(best.result.Size) / (1024 * 1024 * 1024)
			log.Printf("[SUGGESTIONS] Deduplicating TMDB ID %d: kept '%s' (%.1f GB, %d seeders, score %.1f), dropped %d other versions",
				tmdbID, best.result.Title, sizeGB, best.result.Seeders, best.qualityScore, len(candidates)-1)
			uniqueCandidates = append(uniqueCandidates, best)
		}
	}

	// Add candidates without TMDB ID (they'll be deduplicated by info_hash in DB)
	uniqueCandidates = append(uniqueCandidates, candidatesWithoutTMDb...)

	log.Printf("[SUGGESTIONS] After deduplication: %d unique movies from %d results", len(uniqueCandidates), len(results))

	// Third pass: create suggestions in database
	created := 0
	skipped := 0
	for _, candidate := range uniqueCandidates {
		result := candidate.result

		// Skip if already downloaded (checks both download_tasks and download_history)
		// Only check if skipAlreadyDownloaded is true
		if skipAlreadyDownloaded {
			alreadyDownloaded, err := s.downloadManager.CheckIfAlreadyDownloaded(result.InfoHash)
			if err != nil {
				log.Printf("[WARNING] Failed to check download history for %s: %v", result.Title, err)
				continue
			}
			if alreadyDownloaded {
				log.Printf("[SUGGESTIONS] Skipping %s - already downloaded", result.Title)
				skipped++
				continue
			}
		}

		suggestion := &models.DownloadSuggestion{
			SourceID:   result.SourceID,
			Title:      result.Title,
			InfoHash:   result.InfoHash,
			MagnetLink: result.MagnetLink,
			TorrentURL: result.TorrentURL,
			Size:       result.Size,
			Seeders:    result.Seeders,
			Leechers:   result.Leechers,
			Category:   result.Category,
			UploadDate: result.UploadDate,
			Status:     "pending",
			PosterURL:  candidate.poster,
			TMDbID:     candidate.tmdbID,
		}

		if err := s.db.GetDB().
			Where("info_hash = ?", suggestion.InfoHash).
			FirstOrCreate(suggestion).Error; err != nil {
			log.Printf("[WARNING] Failed to create suggestion: %v", err)
			continue
		}
		created++
	}

	totalDuration := time.Since(startTime)
	log.Printf("[SUGGESTIONS] Created %d new suggestions, skipped %d already downloaded (total: %.2fs)", created, skipped, totalDuration.Seconds())
	return created, nil
}

func (s *SuggestionService) ListSuggestions(status string, limit, offset int) ([]models.DownloadSuggestion, int64, error) {
	var suggestions []models.DownloadSuggestion
	var total int64

	query := s.db.GetDB().Model(&models.DownloadSuggestion{}).Preload("Source")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("seeders DESC").
		Limit(limit).
		Offset(offset).
		Find(&suggestions).Error; err != nil {
		return nil, 0, err
	}

	return suggestions, total, nil
}

// normalizeTitle strips quality markers, years, and common suffixes to create a grouping key
func normalizeTitle(title string) string {
	lower := strings.ToLower(title)

	// Remove year patterns (2019, (2019), [2019])
	replacements := []string{
		"1080p", "720p", "2160p", "4k", "uhd",
		"bluray", "blu-ray", "brrip", "bdrip",
		"web-dl", "webdl", "webrip", "web",
		"hdtv", "hdrip", "dvdrip", "cam", "ts", "tc",
		"x264", "x265", "hevc", "avc", "h.264", "h.265",
		"aac", "ac3", "dts", "dd5.1", "ddp5.1", "atmos",
		"remux", "repack", "proper", "extended", "unrated",
	}

	for _, r := range replacements {
		lower = strings.ReplaceAll(lower, r, "")
	}

	// Remove common separators and extra spaces
	lower = strings.ReplaceAll(lower, ".", " ")
	lower = strings.ReplaceAll(lower, "_", " ")
	lower = strings.ReplaceAll(lower, "-", " ")
	lower = strings.ReplaceAll(lower, "[", " ")
	lower = strings.ReplaceAll(lower, "]", " ")
	lower = strings.ReplaceAll(lower, "(", " ")
	lower = strings.ReplaceAll(lower, ")", " ")

	// Remove year numbers (4 digits that look like years)
	// Simple approach: remove standalone 4-digit numbers
	words := strings.Fields(lower)
	var filtered []string
	for _, w := range words {
		if len(w) == 4 {
			// Check if it's a year (1900-2099)
			if w[0] == '1' || w[0] == '2' {
				continue
			}
		}
		filtered = append(filtered, w)
	}

	result := strings.Join(filtered, " ")
	return strings.TrimSpace(result)
}

// groupSuggestions groups suggestions by TMDB ID or normalized title
func (s *SuggestionService) groupSuggestions(suggestions []models.DownloadSuggestion) []models.SuggestionGroup {
	// Group by TMDB ID first, then by normalized title
	byTMDbID := make(map[int][]models.DownloadSuggestion)
	byTitle := make(map[string][]models.DownloadSuggestion)

	for _, sug := range suggestions {
		if sug.TMDbID > 0 {
			byTMDbID[sug.TMDbID] = append(byTMDbID[sug.TMDbID], sug)
		} else {
			norm := normalizeTitle(sug.Title)
			if norm != "" {
				byTitle[norm] = append(byTitle[norm], sug)
			} else {
				byTitle[sug.Title] = append(byTitle[sug.Title], sug)
			}
		}
	}

	var groups []models.SuggestionGroup

	// Process TMDB groups
	for tmdbID, items := range byTMDbID {
		group := s.buildGroup(tmdbID, items)
		groups = append(groups, group)
	}

	// Process title-based groups
	for _, items := range byTitle {
		group := s.buildGroup(0, items)
		groups = append(groups, group)
	}

	return groups
}

// buildGroup creates a SuggestionGroup from a list of suggestions, sorting by quality score
func (s *SuggestionService) buildGroup(tmdbID int, items []models.DownloadSuggestion) models.SuggestionGroup {
	if len(items) == 0 {
		return models.SuggestionGroup{}
	}

	// Find the best poster URL from any item in the group
	var bestPosterURL string
	var bestTitle string
	for _, item := range items {
		if bestPosterURL == "" && item.PosterURL != "" {
			bestPosterURL = item.PosterURL
		}
		if bestTitle == "" && item.Title != "" {
			bestTitle = item.Title
		}
		if bestPosterURL != "" && bestTitle != "" {
			break
		}
	}

	// Sort by quality score descending
	sort.Slice(items, func(i, j int) bool {
		scoreI := s.CalculateTorrentQualityScore(items[i].Title, items[i].Size, items[i].Seeders)
		scoreJ := s.CalculateTorrentQualityScore(items[j].Title, items[j].Size, items[j].Seeders)
		return scoreI > scoreJ
	})

	primary := items[0]
	var alternates []models.DownloadSuggestion
	if len(items) > 1 {
		alternates = items[1:]
	}

	// Use primary title if no better title found, but prefer a title with poster
	groupTitle := bestTitle
	if groupTitle == "" {
		groupTitle = primary.Title
	}

	// Use the best poster from the group, not just primary
	groupPoster := bestPosterURL
	if groupPoster == "" {
		groupPoster = primary.PosterURL
	}

	return models.SuggestionGroup{
		TMDbID:       tmdbID,
		Title:        groupTitle,
		PosterURL:    groupPoster,
		Primary:      primary,
		Alternates:   alternates,
		TotalOptions: len(items),
	}
}

func (s *SuggestionService) ListGroupedSuggestions(status string, limit, offset int) ([]models.SuggestionGroup, int64, error) {
	// Fetch all suggestions for the status (we need them all to group properly)
	var suggestions []models.DownloadSuggestion
	var total int64

	query := s.db.GetDB().Model(&models.DownloadSuggestion{}).Preload("Source")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("seeders DESC").Find(&suggestions).Error; err != nil {
		return nil, 0, err
	}

	groups := s.groupSuggestions(suggestions)

	// Sort groups by primary seeders descending for deterministic pagination
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Primary.Seeders > groups[j].Primary.Seeders
	})

	// Apply pagination at the group level
	totalGroups := int64(len(groups))
	start := offset
	end := offset + limit
	if start > len(groups) {
		start = len(groups)
	}
	if end > len(groups) {
		end = len(groups)
	}

	return groups[start:end], totalGroups, nil
}

func (s *SuggestionService) SearchGroupedSuggestions(searchQuery, status, sortBy string, minSeeders, limit, offset int) ([]models.SuggestionGroup, int64, error) {
	var suggestions []models.DownloadSuggestion
	var total int64

	query := s.db.GetDB().Model(&models.DownloadSuggestion{}).Preload("Source")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if searchQuery != "" {
		query = query.Where("title LIKE ?", "%"+searchQuery+"%")
	}

	if minSeeders > 0 {
		query = query.Where("seeders >= ?", minSeeders)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "seeders DESC"
	switch sortBy {
	case "size":
		orderClause = "size ASC"
	case "upload_date":
		orderClause = "upload_date DESC"
	case "quality_score":
		orderClause = "seeders DESC, upload_date DESC"
	default:
		orderClause = "seeders DESC"
	}

	if err := query.Order(orderClause).Find(&suggestions).Error; err != nil {
		return nil, 0, err
	}

	groups := s.groupSuggestions(suggestions)

	// Sort groups by primary seeders descending for deterministic pagination
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Primary.Seeders > groups[j].Primary.Seeders
	})

	// Apply pagination at the group level
	totalGroups := int64(len(groups))
	start := offset
	end := offset + limit
	if start > len(groups) {
		start = len(groups)
	}
	if end > len(groups) {
		end = len(groups)
	}

	return groups[start:end], totalGroups, nil
}

func (s *SuggestionService) SearchSuggestions(searchQuery, status, sortBy string, minSeeders, limit, offset int) ([]models.DownloadSuggestion, int64, error) {
	var suggestions []models.DownloadSuggestion
	var total int64

	query := s.db.GetDB().Model(&models.DownloadSuggestion{}).Preload("Source")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if searchQuery != "" {
		query = query.Where("title LIKE ?", "%"+searchQuery+"%")
	}

	if minSeeders > 0 {
		query = query.Where("seeders >= ?", minSeeders)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "seeders DESC"
	switch sortBy {
	case "size":
		orderClause = "size ASC"
	case "upload_date":
		orderClause = "upload_date DESC"
	case "quality_score":
		orderClause = "seeders DESC, upload_date DESC"
	default:
		orderClause = "seeders DESC"
	}

	if err := query.Order(orderClause).Limit(limit).Offset(offset).Find(&suggestions).Error; err != nil {
		return nil, 0, err
	}

	return suggestions, total, nil
}

func (s *SuggestionService) GetTopRecommendations(limit int) ([]models.DownloadSuggestion, error) {
	var suggestions []models.DownloadSuggestion

	if err := s.db.GetDB().
		Model(&models.DownloadSuggestion{}).
		Preload("Source").
		Where("status = ?", "pending").
		Where("seeders >= ?", 10).
		Order("seeders DESC, upload_date DESC").
		Limit(limit).
		Find(&suggestions).Error; err != nil {
		return nil, err
	}

	return suggestions, nil
}

func (s *SuggestionService) GetSuggestion(id uint) (*models.DownloadSuggestion, error) {
	var suggestion models.DownloadSuggestion
	if err := s.db.GetDB().Preload("Source").First(&suggestion, id).Error; err != nil {
		return nil, err
	}
	return &suggestion, nil
}

func (s *SuggestionService) ApproveSuggestion(id uint, notes string, autoStart bool) error {
	var suggestion models.DownloadSuggestion
	if err := s.db.GetDB().First(&suggestion, id).Error; err != nil {
		return err
	}

	if suggestion.Status != "pending" {
		return fmt.Errorf("suggestion is not pending (status: %s)", suggestion.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      "approved",
		"approved_at": now,
	}
	if notes != "" {
		updates["notes"] = notes
	}

	return s.db.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&suggestion).Updates(updates).Error; err != nil {
			return err
		}

		task := &models.DownloadTask{
			SourceID:   suggestion.SourceID,
			Title:      suggestion.Title,
			InfoHash:   suggestion.InfoHash,
			MagnetLink: suggestion.MagnetLink,
			TorrentURL: suggestion.TorrentURL,
			Size:       suggestion.Size,
			Seeders:    suggestion.Seeders,
			Leechers:   suggestion.Leechers,
			Status:     "pending",
			PosterURL:  suggestion.PosterURL,
			TMDbID:     suggestion.TMDbID,
		}

		if err := tx.Create(task).Error; err != nil {
			return fmt.Errorf("failed to create download task: %w", err)
		}

		// Auto-start download if enabled and high seeders
		if autoStart && suggestion.Seeders >= 50 {
			task.Status = "downloading"
			startTime := time.Now()
			task.StartedAt = &startTime
			if err := tx.Save(task).Error; err != nil {
				log.Printf("[WARNING] Failed to update task status for auto-start: %v", err)
			}

			// Start download in background
			go func() {
				if err := s.downloadManager.StartTask(task.ID); err != nil {
					log.Printf("[ERROR] Auto-start failed for task %d: %v", task.ID, err)
					// Update status back to pending
					s.db.GetDB().Model(&models.DownloadTask{}).
						Where("id = ?", task.ID).
						Updates(map[string]interface{}{
							"status":     "pending",
							"started_at": nil,
						})
				}
			}()

			log.Printf("[SUGGESTIONS] Approved suggestion %d and auto-started download task %d (seeders: %d)", id, task.ID, suggestion.Seeders)
		} else {
			log.Printf("[SUGGESTIONS] Approved suggestion %d and created download task %d", id, task.ID)
		}

		return nil
	})
}

func (s *SuggestionService) RejectSuggestion(id uint, notes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "rejected",
		"rejected_at": now,
	}
	if notes != "" {
		updates["notes"] = notes
	}

	result := s.db.GetDB().Model(&models.DownloadSuggestion{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("suggestion not found or not pending")
	}

	log.Printf("[SUGGESTIONS] Rejected suggestion %d", id)
	return nil
}

func (s *SuggestionService) BulkApprove(ids []uint, notes string) (int, error) {
	approved := 0
	for _, id := range ids {
		if err := s.ApproveSuggestion(id, notes, false); err != nil {
			log.Printf("[WARNING] Failed to approve suggestion %d: %v", id, err)
			continue
		}
		approved++
	}
	log.Printf("[SUGGESTIONS] Bulk approved %d/%d suggestions", approved, len(ids))
	return approved, nil
}

func (s *SuggestionService) BulkReject(ids []uint, notes string) (int, error) {
	rejected := 0
	for _, id := range ids {
		if err := s.RejectSuggestion(id, notes); err != nil {
			log.Printf("[WARNING] Failed to reject suggestion %d: %v", id, err)
			continue
		}
		rejected++
	}
	log.Printf("[SUGGESTIONS] Bulk rejected %d/%d suggestions", rejected, len(ids))
	return rejected, nil
}

func (s *SuggestionService) ClearRejected() (int64, error) {
	result := s.db.GetDB().Where("status = ?", "rejected").Delete(&models.DownloadSuggestion{})
	if result.Error != nil {
		return 0, result.Error
	}
	log.Printf("[SUGGESTIONS] Cleared %d rejected suggestions", result.RowsAffected)
	return result.RowsAffected, nil
}

func (s *SuggestionService) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	statuses := []string{"pending", "approved", "rejected"}
	for _, status := range statuses {
		var count int64
		if err := s.db.GetDB().Model(&models.DownloadSuggestion{}).
			Where("status = ?", status).
			Count(&count).Error; err != nil {
			return nil, err
		}
		stats[status] = count
	}

	var total int64
	if err := s.db.GetDB().Model(&models.DownloadSuggestion{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}
