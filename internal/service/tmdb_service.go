package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

const (
	TMDbAPIBaseURL   = "https://api.themoviedb.org/3"
	TMDbImageBaseURL = "https://image.tmdb.org/t/p/w500"
)

type TMDbService struct {
	apiKey     string
	httpClient *http.Client
	db         *db.Database
}

type TMDbSearchResult struct {
	Page         int         `json:"page"`
	Results      []TMDbMovie `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type TMDbMovie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	Overview      string  `json:"overview"`
	VoteAverage   float64 `json:"vote_average"`
	Popularity    float64 `json:"popularity"`
}

type TMDbMovieDetails struct {
	ID            int         `json:"id"`
	Title         string      `json:"title"`
	OriginalTitle string      `json:"original_title"`
	Tagline       string      `json:"tagline"`
	Overview      string      `json:"overview"`
	ReleaseDate   string      `json:"release_date"`
	Runtime       int         `json:"runtime"`
	VoteAverage   float64     `json:"vote_average"`
	VoteCount     int         `json:"vote_count"`
	PosterPath    string      `json:"poster_path"`
	BackdropPath  string      `json:"backdrop_path"`
	Genres        []TMDbGenre `json:"genres"`
	Credits       TMDbCredits `json:"credits"`
	Status        string      `json:"status"`
	Budget        int64       `json:"budget"`
	Revenue       int64       `json:"revenue"`
}

type TMDbGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TMDbCredits struct {
	Cast []TMDbCastMember `json:"cast"`
	Crew []TMDbCrewMember `json:"crew"`
}

type TMDbCastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

type TMDbCrewMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
}

// TV Show related types
type TMDbTVSearchResult struct {
	Page         int        `json:"page"`
	Results      []TMDbShow `json:"results"`
	TotalPages   int        `json:"total_pages"`
	TotalResults int        `json:"total_results"`
}

type TMDbShow struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	FirstAirDate string  `json:"first_air_date"`
	Overview     string  `json:"overview"`
	VoteAverage  float64 `json:"vote_average"`
	Popularity   float64 `json:"popularity"`
}

// MediaInfo holds extracted information from a title
type MediaInfo struct {
	Name    string
	Year    int
	Season  int
	Episode int
	IsTV    bool
}

func NewTMDbService(database *db.Database) *TMDbService {
	// Get API key from environment variable, fallback to empty string
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		log.Println("[TMDb] Warning: TMDB_API_KEY environment variable not set. Poster fetching will be disabled.")
	}

	return &TMDbService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		db: database,
	}
}

// ExtractMediaInfo extracts show/movie name, year, season, and episode from torrent title
func (t *TMDbService) ExtractMediaInfo(title string) MediaInfo {
	info := MediaInfo{}

	// Detect TV show pattern (S01E01, 1x01, etc.)
	tvPatterns := []string{
		`(?i)S(\d+)E(\d+)`,                     // S01E01
		`(?i)S(\d+)\s*E(\d+)`,                  // S01 E01
		`(?i)(\d+)x(\d+)`,                      // 1x01
		`(?i)Season\s*(\d+).*?Episode\s*(\d+)`, // Season 1 Episode 1
	}

	for _, pattern := range tvPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(title); len(matches) >= 3 {
			info.IsTV = true
			info.Season, _ = strconv.Atoi(matches[1])
			info.Episode, _ = strconv.Atoi(matches[2])
			break
		}
	}

	// Extract year (4 digits, 1900-2099)
	yearRegex := regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	if yearMatch := yearRegex.FindString(title); yearMatch != "" {
		info.Year, _ = strconv.Atoi(yearMatch)
	}

	// Clean title
	cleanTitle := title

	// For TV shows, remove everything after the episode pattern (removes episode titles)
	if info.IsTV {
		for _, pattern := range tvPatterns {
			re := regexp.MustCompile(pattern)
			if loc := re.FindStringIndex(cleanTitle); loc != nil {
				// Keep everything before the pattern
				cleanTitle = cleanTitle[:loc[0]]
				break
			}
		}
	}

	// Remove everything in brackets and parentheses
	cleanTitle = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(cleanTitle, " ")
	cleanTitle = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(cleanTitle, " ")

	// Remove year from title if it's not a TV show (TV shows often have year in name like "All Creatures Great and Small 2020")
	if !info.IsTV {
		cleanTitle = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`).ReplaceAllString(cleanTitle, "")
	}

	// Remove quality markers and codecs
	cleanTitle = regexp.MustCompile(`(?i)\b(1080p|720p|2160p|4K|UHD|BluRay|BRRip|WEBRip|WEB-DL|WEB|HDRip|HDTV|DVDRip|BDRip|TS|TELES|CAM|SCREENER|HC|mkv|mp4|avi)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\bWEB\s+DL\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(x264|x265|H\.?264|H\.?265|HEVC|AVC|10bit|8bit)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b10\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b8\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b[Hh]\s+26[45]\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b[xX]\s+26[45]\b`).ReplaceAllString(cleanTitle, "")

	// Remove file size indicators
	cleanTitle = regexp.MustCompile(`\b\d+\.?\d*\s?(GB|MB)\b`).ReplaceAllString(cleanTitle, "")

	// Remove audio formats
	cleanTitle = regexp.MustCompile(`(?i)\b(AAC|AC3|DDP5?\.?1?|DD5?\.?1?|DDP|DD|Atmos|DTS|TrueHD|FLAC|MP3)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[257]\s*\.?\s*[01]\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\bDD[P]?\s*5\s+1\b`).ReplaceAllString(cleanTitle, "")

	// Remove release groups and source tags
	cleanTitle = regexp.MustCompile(`(?i)\b(YIFY|YTS|RARBG|ETRG|BONE|FGT|CMRG|EVO|ION10|SPARKS|AMZN|NF|IMAX|RGB|EniaHD|GalaxyRG|RMTeam|KyoGo|Asiimov|BYNDR|MA|FLUX|NeoNoir|NTb|DOLORES|GRACE|ETHEL|MeGusta|DARKFLiX)\b`).ReplaceAllString(cleanTitle, "")

	// Remove language names and codes
	cleanTitle = regexp.MustCompile(`(?i)\b(EN|ENG|ENGLISH|MULTI|DUAL|SUBBED?|DUBBED?|CHINESE|KOREAN|JAPANESE|SPANISH|FRENCH|GERMAN|ITALIAN|RUSSIAN)\b`).ReplaceAllString(cleanTitle, "")

	// Remove version tags
	cleanTitle = regexp.MustCompile(`(?i)\bV\d+\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(REPACK|PROPER|EXTENDED|UNRATED|DIRECTORS?\.?CUT|DC)\b`).ReplaceAllString(cleanTitle, "")

	// Remove standalone single letters followed by artifacts
	cleanTitle = regexp.MustCompile(`\b[a-z]\]`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[a-z]\s+\[`).ReplaceAllString(cleanTitle, " ")

	// Remove remaining brackets and parentheses
	cleanTitle = strings.ReplaceAll(cleanTitle, "[", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "]", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "(", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, ")", " ")

	// Clean up separators
	cleanTitle = regexp.MustCompile(`[._-]+`).ReplaceAllString(cleanTitle, " ")

	// Remove multiple spaces
	cleanTitle = regexp.MustCompile(`\s+`).ReplaceAllString(cleanTitle, " ")

	// Remove standalone single letters (except 'A' and 'I')
	cleanTitle = regexp.MustCompile(`\b[b-hj-z]\b`).ReplaceAllString(cleanTitle, " ")

	// For TV shows, don't remove all numbers (years might be part of the name)
	if !info.IsTV {
		// Remove remaining standalone numbers
		cleanTitle = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(cleanTitle, " ")
	}

	// Remove "DL" standalone
	cleanTitle = regexp.MustCompile(`(?i)\bDL\b`).ReplaceAllString(cleanTitle, " ")

	// Remove multiple spaces again
	cleanTitle = regexp.MustCompile(`\s+`).ReplaceAllString(cleanTitle, " ")

	// Trim and set name
	info.Name = strings.TrimSpace(cleanTitle)

	return info
}

// ExtractMovieInfo extracts movie name and year from torrent title (kept for backward compatibility)
func (t *TMDbService) ExtractMovieInfo(title string) (name string, year int) {
	info := t.ExtractMediaInfo(title)
	return info.Name, info.Year
}

func (t *TMDbService) SearchMovie(title string, year int) (*TMDbMovie, error) {
	// Check if API key is set
	if t.apiKey == "" {
		return nil, fmt.Errorf("TMDb API key not configured")
	}

	params := url.Values{}
	params.Add("api_key", t.apiKey)
	params.Add("query", title)
	if year > 0 {
		params.Add("year", strconv.Itoa(year))
	}

	searchURL := fmt.Sprintf("%s/search/movie?%s", TMDbAPIBaseURL, params.Encode())

	resp, err := t.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search TMDb: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status %d", resp.StatusCode)
	}

	var searchResult TMDbSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode TMDb response: %w", err)
	}

	if len(searchResult.Results) == 0 {
		return nil, fmt.Errorf("no results found for '%s'", title)
	}

	// Return the first (most relevant) result
	return &searchResult.Results[0], nil
}

// SearchTV searches for TV shows on TMDb
func (t *TMDbService) SearchTV(title string, year int) (*TMDbShow, error) {
	// Check if API key is set
	if t.apiKey == "" {
		return nil, fmt.Errorf("TMDb API key not configured")
	}

	params := url.Values{}
	params.Add("api_key", t.apiKey)
	params.Add("query", title)
	if year > 0 {
		params.Add("first_air_date_year", strconv.Itoa(year))
	}

	searchURL := fmt.Sprintf("%s/search/tv?%s", TMDbAPIBaseURL, params.Encode())

	resp, err := t.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search TMDb TV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API returned status %d", resp.StatusCode)
	}

	var searchResult TMDbTVSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode TMDb TV response: %w", err)
	}

	if len(searchResult.Results) == 0 {
		// If no results and we have a year in the title, try removing it
		if year > 0 {
			yearStr := strconv.Itoa(year)
			if strings.Contains(title, yearStr) {
				cleanTitle := strings.ReplaceAll(title, yearStr, "")
				cleanTitle = strings.TrimSpace(cleanTitle)
				log.Printf("[TMDB] Retrying TV search without year: '%s'", cleanTitle)
				return t.SearchTV(cleanTitle, 0) // Retry without year
			}
		}
		return nil, fmt.Errorf("no TV results found for '%s'", title)
	}

	// Return the first (most relevant) result
	return &searchResult.Results[0], nil
}

func (t *TMDbService) GetPosterURL(movie *TMDbMovie) string {
	if movie == nil || movie.PosterPath == "" {
		return ""
	}
	return TMDbImageBaseURL + movie.PosterPath
}

// GetTVPosterURL returns the full poster URL for a TV show
func (t *TMDbService) GetTVPosterURL(show *TMDbShow) string {
	if show == nil || show.PosterPath == "" {
		return ""
	}
	return TMDbImageBaseURL + show.PosterPath
}

func (t *TMDbService) GetBackdropURL(movie *TMDbMovie) string {
	if movie == nil || movie.BackdropPath == "" {
		return ""
	}
	return "https://image.tmdb.org/t/p/w1280" + movie.BackdropPath
}

// GetTVBackdropURL returns the full backdrop URL for a TV show
func (t *TMDbService) GetTVBackdropURL(show *TMDbShow) string {
	if show == nil || show.BackdropPath == "" {
		return ""
	}
	return "https://image.tmdb.org/t/p/w1280" + show.BackdropPath
}

// FetchPosterForTask fetches poster URL for a download task (supports both movies and TV shows)
// Now with database caching for referential integrity
func (t *TMDbService) FetchPosterForTask(title string) (posterURL string, tmdbID int, err error) {
	// Extract media info (detects TV vs Movie automatically)
	mediaInfo := t.ExtractMediaInfo(title)

	// Create a clean title for better matching
	cleanTitle := t.CleanTitle(mediaInfo.Name)

	// Check if we already have this in our metadata cache
	if t.db != nil {
		var metadata models.MovieMetadata
		result := t.db.GetDB().Where("clean_title = ? OR title = ?", cleanTitle, mediaInfo.Name).First(&metadata)
		if result.Error == nil && metadata.PosterURL != "" {
			log.Printf("[TMDB] Found cached poster for '%s': %s", mediaInfo.Name, metadata.PosterURL)
			return metadata.PosterURL, metadata.TMDbID, nil
		}
	}

	// Not in cache, fetch from TMDb API
	if mediaInfo.IsTV {
		// Search for TV show
		log.Printf("[TMDB] Searching TMDb TV for: '%s' (%d) [S%02dE%02d]",
			mediaInfo.Name, mediaInfo.Year, mediaInfo.Season, mediaInfo.Episode)

		show, err := t.SearchTV(mediaInfo.Name, mediaInfo.Year)
		if err != nil {
			log.Printf("[TMDB] TV search failed for '%s': %v", mediaInfo.Name, err)
			return "", 0, err
		}

		posterURL = t.GetTVPosterURL(show)
		backdropURL := t.GetTVBackdropURL(show)
		log.Printf("[TMDB] Found TV poster for '%s': %s", mediaInfo.Name, posterURL)

		// Save to database cache for future use
		if t.db != nil {
			metadata := models.MovieMetadata{
				Title:       mediaInfo.Name,
				CleanTitle:  cleanTitle,
				Year:        mediaInfo.Year,
				TMDbID:      show.ID,
				PosterURL:   posterURL,
				BackdropURL: backdropURL,
				Overview:    show.Overview,
				Rating:      show.VoteAverage,
			}

			// Use FirstOrCreate to avoid duplicates
			if err := t.db.GetDB().Where("tm_db_id = ?", show.ID).FirstOrCreate(&metadata).Error; err != nil {
				log.Printf("[TMDB] Warning: Failed to cache TV metadata for '%s': %v", mediaInfo.Name, err)
			} else {
				log.Printf("[TMDB] Cached TV metadata for '%s' (TMDb ID: %d)", mediaInfo.Name, show.ID)
			}
		}

		return posterURL, show.ID, nil
	}

	// Search for movie
	log.Printf("[TMDB] Searching TMDb Movie for: '%s' (%d)", mediaInfo.Name, mediaInfo.Year)

	movie, err := t.SearchMovie(mediaInfo.Name, mediaInfo.Year)
	if err != nil {
		log.Printf("[TMDB] Movie search failed for '%s': %v", mediaInfo.Name, err)
		return "", 0, err
	}

	posterURL = t.GetPosterURL(movie)
	backdropURL := t.GetBackdropURL(movie)
	log.Printf("[TMDB] Found movie poster for '%s': %s", mediaInfo.Name, posterURL)

	// Save to database cache for future use
	if t.db != nil {
		metadata := models.MovieMetadata{
			Title:       mediaInfo.Name,
			CleanTitle:  cleanTitle,
			Year:        mediaInfo.Year,
			TMDbID:      movie.ID,
			PosterURL:   posterURL,
			BackdropURL: backdropURL,
			Overview:    movie.Overview,
			Rating:      movie.VoteAverage,
		}

		// Use FirstOrCreate to avoid duplicates
		if err := t.db.GetDB().Where("tm_db_id = ?", movie.ID).FirstOrCreate(&metadata).Error; err != nil {
			log.Printf("[TMDB] Warning: Failed to cache metadata for '%s': %v", mediaInfo.Name, err)
		} else {
			log.Printf("[TMDB] Cached metadata for '%s' (TMDb ID: %d)", mediaInfo.Name, movie.ID)
		}
	}

	return posterURL, movie.ID, nil
}

// CleanTitle normalizes a title for better matching
func (t *TMDbService) CleanTitle(title string) string {
	clean := title
	clean = strings.ReplaceAll(clean, "[", " ")
	clean = strings.ReplaceAll(clean, "]", " ")
	clean = strings.ReplaceAll(clean, "(", " ")
	clean = strings.ReplaceAll(clean, ")", " ")
	clean = strings.TrimSpace(clean)
	clean = strings.ToLower(clean)
	return clean
}

// GetMovieDetails fetches detailed information about a movie by TMDB ID
func (t *TMDbService) GetMovieDetails(tmdbID int) (*TMDbMovieDetails, error) {
	if t.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	// Fetch movie details with credits appended
	url := fmt.Sprintf("%s/movie/%d?api_key=%s&append_to_response=credits", TMDbAPIBaseURL, tmdbID, t.apiKey)

	resp, err := t.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var details TMDbMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &details, nil
}
