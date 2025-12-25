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

// ExtractMovieInfo extracts movie name and year from torrent title
func (t *TMDbService) ExtractMovieInfo(title string) (name string, year int) {
	// Extract year first (4 digits)
	yearRegex := regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	yearMatch := yearRegex.FindString(title)
	if yearMatch != "" {
		year, _ = strconv.Atoi(yearMatch)
	}

	// Remove common release group tags and quality markers
	cleanTitle := title

	// Remove everything in brackets and parentheses (content AND the brackets themselves)
	cleanTitle = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(cleanTitle, " ")
	cleanTitle = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(cleanTitle, " ")

	// Remove year
	cleanTitle = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`).ReplaceAllString(cleanTitle, "")

	// Remove quality markers and codecs (more comprehensive)
	cleanTitle = regexp.MustCompile(`(?i)\b(1080p|720p|2160p|4K|UHD|BluRay|BRRip|WEBRip|WEB-DL|WEB|HDRip|HDTV|DVDRip|BDRip|TS|TELES|CAM|SCREENER|HC|mkv|mp4|avi)\b`).ReplaceAllString(cleanTitle, "")
	// Remove WEB DL variations
	cleanTitle = regexp.MustCompile(`(?i)\bWEB\s+DL\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(x264|x265|H\.?264|H\.?265|HEVC|AVC|10bit|8bit)\b`).ReplaceAllString(cleanTitle, "")
	// Remove bit depth patterns with various separators
	cleanTitle = regexp.MustCompile(`(?i)\b10\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b8\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	// Remove codec patterns with spaces (e.g., "H 264", "x 265")
	cleanTitle = regexp.MustCompile(`(?i)\b[Hh]\s+26[45]\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b[xX]\s+26[45]\b`).ReplaceAllString(cleanTitle, "")

	// Remove file size indicators
	cleanTitle = regexp.MustCompile(`\b\d+\.?\d*\s?(GB|MB)\b`).ReplaceAllString(cleanTitle, "")

	// Remove audio formats (including standalone numbers like "5 1" or "2 0")
	cleanTitle = regexp.MustCompile(`(?i)\b(AAC|AC3|DDP5?\.?1?|DD5?\.?1?|DDP|DD|Atmos|DTS|TrueHD|FLAC|MP3)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[257]\s*\.?\s*[01]\b`).ReplaceAllString(cleanTitle, "") // Remove 5.1, 7.1, 2.0, etc.
	// Remove DDP patterns with spaces (e.g., "DDP5 1", "DD5 1")
	cleanTitle = regexp.MustCompile(`(?i)\bDD[P]?\s*5\s+1\b`).ReplaceAllString(cleanTitle, "")

	// Remove release groups and source tags
	cleanTitle = regexp.MustCompile(`(?i)\b(YIFY|YTS|RARBG|ETRG|BONE|FGT|CMRG|EVO|ION10|SPARKS|AMZN|NF|IMAX|RGB|EniaHD|GalaxyRG|RMTeam|KyoGo|Asiimov|BYNDR|MA|FLUX|NeoNoir)\b`).ReplaceAllString(cleanTitle, "")

	// Remove language names and codes
	cleanTitle = regexp.MustCompile(`(?i)\b(EN|ENG|ENGLISH|MULTI|DUAL|SUBBED?|DUBBED?|CHINESE|KOREAN|JAPANESE|SPANISH|FRENCH|GERMAN|ITALIAN|RUSSIAN)\b`).ReplaceAllString(cleanTitle, "")

	// Remove version tags
	cleanTitle = regexp.MustCompile(`(?i)\bV\d+\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(REPACK|PROPER|EXTENDED|UNRATED|DIRECTORS?\.?CUT|DC)\b`).ReplaceAllString(cleanTitle, "")

	// Remove standalone single letters followed by ] or other artifacts (e.g., "p]", "p ")
	cleanTitle = regexp.MustCompile(`\b[a-z]\]`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[a-z]\s+\[`).ReplaceAllString(cleanTitle, " ")

	// Remove remaining brackets and parentheses
	cleanTitle = strings.ReplaceAll(cleanTitle, "[", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "]", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "(", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, ")", " ")

	// Clean up separators (dots, dashes, underscores) -> spaces
	cleanTitle = regexp.MustCompile(`[._-]+`).ReplaceAllString(cleanTitle, " ")

	// Remove multiple spaces
	cleanTitle = regexp.MustCompile(`\s+`).ReplaceAllString(cleanTitle, " ")

	// Remove standalone single letters at word boundaries (except common words like "A")
	cleanTitle = regexp.MustCompile(`\b[b-z]\b`).ReplaceAllString(cleanTitle, " ")

	// Remove remaining standalone numbers (artifacts from codec/audio patterns)
	cleanTitle = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(cleanTitle, " ")

	// Remove "DL" standalone (remaining from WEB-DL)
	cleanTitle = regexp.MustCompile(`(?i)\bDL\b`).ReplaceAllString(cleanTitle, " ")

	// Remove multiple spaces again after single letter removal
	cleanTitle = regexp.MustCompile(`\s+`).ReplaceAllString(cleanTitle, " ")

	// Trim and return
	cleanTitle = strings.TrimSpace(cleanTitle)

	return cleanTitle, year
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

func (t *TMDbService) GetPosterURL(movie *TMDbMovie) string {
	if movie == nil || movie.PosterPath == "" {
		return ""
	}
	return TMDbImageBaseURL + movie.PosterPath
}

func (t *TMDbService) GetBackdropURL(movie *TMDbMovie) string {
	if movie == nil || movie.BackdropPath == "" {
		return ""
	}
	return "https://image.tmdb.org/t/p/w1280" + movie.BackdropPath
}

// FetchPosterForTask fetches poster URL for a download task
// Now with database caching for referential integrity
func (t *TMDbService) FetchPosterForTask(title string) (posterURL string, tmdbID int, err error) {
	movieName, year := t.ExtractMovieInfo(title)

	// Create a clean title for better matching
	cleanTitle := t.CleanTitle(movieName)

	// Check if we already have this movie in our metadata cache
	if t.db != nil {
		var metadata models.MovieMetadata
		result := t.db.GetDB().Where("clean_title = ? OR title = ?", cleanTitle, movieName).First(&metadata)
		if result.Error == nil && metadata.PosterURL != "" {
			log.Printf("[TMDB] Found cached poster for '%s': %s", movieName, metadata.PosterURL)
			return metadata.PosterURL, metadata.TMDbID, nil
		}
	}

	// Not in cache, fetch from TMDb API
	log.Printf("[TMDB] Searching TMDb for: '%s' (%d)", movieName, year)

	movie, err := t.SearchMovie(movieName, year)
	if err != nil {
		log.Printf("[TMDB] Search failed for '%s': %v", movieName, err)
		return "", 0, err
	}

	posterURL = t.GetPosterURL(movie)
	backdropURL := t.GetBackdropURL(movie)
	log.Printf("[TMDB] Found poster for '%s': %s", movieName, posterURL)

	// Save to database cache for future use
	if t.db != nil {
		metadata := models.MovieMetadata{
			Title:       movieName,
			CleanTitle:  cleanTitle,
			Year:        year,
			TMDbID:      movie.ID,
			PosterURL:   posterURL,
			BackdropURL: backdropURL,
			Overview:    movie.Overview,
			Rating:      movie.VoteAverage,
		}

		// Use FirstOrCreate to avoid duplicates
		if err := t.db.GetDB().Where("tm_db_id = ?", movie.ID).FirstOrCreate(&metadata).Error; err != nil {
			log.Printf("[TMDB] Warning: Failed to cache metadata for '%s': %v", movieName, err)
		} else {
			log.Printf("[TMDB] Cached metadata for '%s' (TMDb ID: %d)", movieName, movie.ID)
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
