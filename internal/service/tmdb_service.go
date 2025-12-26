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
	TMDbAPIBaseURL    = "https://api.themoviedb.org/3"
	TMDbImageBaseURL  = "https://image.tmdb.org/t/p/w500"
	GoogleBooksAPIURL = "https://www.googleapis.com/books/v1/volumes"
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

// Google Books related types
type GoogleBooksSearchResult struct {
	Kind       string            `json:"kind"`
	TotalItems int               `json:"totalItems"`
	Items      []GoogleBooksItem `json:"items"`
}

type GoogleBooksItem struct {
	ID         string                `json:"id"`
	VolumeInfo GoogleBooksVolumeInfo `json:"volumeInfo"`
}

type GoogleBooksVolumeInfo struct {
	Title               string                  `json:"title"`
	Authors             []string                `json:"authors"`
	Publisher           string                  `json:"publisher"`
	PublishedDate       string                  `json:"publishedDate"`
	Description         string                  `json:"description"`
	ImageLinks          GoogleBooksImageLinks   `json:"imageLinks"`
	IndustryIdentifiers []GoogleBooksIdentifier `json:"industryIdentifiers"`
}

type GoogleBooksImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
}

type GoogleBooksIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

// MediaInfo holds extracted information from a title
type MediaInfo struct {
	Name        string
	Year        int
	Season      int
	Episode     int
	IsTV        bool
	ContentType string // "movie", "tv", "music", "book", "other"
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

// DetectContentType determines if the title is a book, music, movie, TV show, or other
func (t *TMDbService) DetectContentType(title string) string {
	titleLower := strings.ToLower(title)

	// Book patterns - check for common ebook formats and academic keywords
	bookPatterns := []string{
		`\.(pdf|epub|mobi|azw3|djvu|fb2)`, // File extensions
		`\b(matrix calculus|matrix analysis|matrix mathematics|matrix theory|matrix decomposition)`, // Mathematical books
		`\b(linear algebra|boolean matrix|kronecker product|tensor product)`,                        // Math topics
		`\b(\d+ed|edition|workbook|solutions|springer|wiley|cambridge|oxford|press)\b`,              // Publishing
		`\b(ultrastructural pathology|pathology of the cell)\b`,                                     // Academic
	}
	for _, pattern := range bookPatterns {
		if matched, _ := regexp.MatchString(pattern, titleLower); matched {
			return "book"
		}
	}

	// Music patterns - check for audio formats and music keywords
	musicPatterns := []string{
		`\.(flac|mp3|aac|m4a|wav|ogg|opus|wma|ape|alac)\b`,                // Audio file extensions
		`\b(flac|lossless|320kbps|v0|album|ep|single|ost|soundtrack)\b`,   // Audio quality/types
		`\b(live at|concert|records|label|\bcd\b|vinyl|discography)\b`,    // Music releases
		`\b(dj |remix|mix|feat\.|ft\.|artist|band)\b`,                     // Music terminology
		`\d{4}\s+(house|techno|trance|dubstep|hip.?hop|jazz|classical)\b`, // Music genres with year
	}
	for _, pattern := range musicPatterns {
		if matched, _ := regexp.MatchString(pattern, titleLower); matched {
			return "music"
		}
	}

	// Adult content patterns
	adultPatterns := []string{
		`\bxxx\b`,
		`\b(manyvids|onlyfans|pornhub|brazzers)\b`,
		`\b(purple bitch|cosplay xxx)\b`,
	}
	for _, pattern := range adultPatterns {
		if matched, _ := regexp.MatchString(pattern, titleLower); matched {
			return "adult"
		}
	}

	// TV show patterns (will be overridden by S01E01 detection later)
	tvPatterns := []string{
		`\bS\d+E\d+\b`,
		`\b\d+x\d+\b`,
		`\b(season|episode)\b`,
		`\b\d{4}\s+\d{1,2}\s+\d{1,2}\b`, // Date format: YYYY MM DD (daily shows)
	}
	for _, pattern := range tvPatterns {
		if matched, _ := regexp.MatchString(pattern, titleLower); matched {
			return "tv"
		}
	}

	// Movie patterns (default if it has video qualities but no TV markers)
	moviePatterns := []string{
		`\b(1080p|720p|2160p|4k|bluray|webrip|web-dl|dvdrip|brrip|hdtv)\b`,
	}
	for _, pattern := range moviePatterns {
		if matched, _ := regexp.MatchString(pattern, titleLower); matched {
			return "movie"
		}
	}

	// Default to "other" if nothing matches
	return "other"
}

// ExtractMediaInfo extracts show/movie name, year, season, and episode from torrent title
func (t *TMDbService) ExtractMediaInfo(title string) MediaInfo {
	info := MediaInfo{}

	// Detect content type first
	info.ContentType = t.DetectContentType(title)

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
			info.ContentType = "tv" // Override if TV pattern detected
			info.Season, _ = strconv.Atoi(matches[1])
			info.Episode, _ = strconv.Atoi(matches[2])
			break
		}
	}

	for _, pattern := range tvPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(title); len(matches) >= 3 {
			info.IsTV = true
			info.ContentType = "tv" // Override if TV pattern detected
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

	// Remove common words that appear after brackets (genres, categories, etc.)
	cleanTitle = regexp.MustCompile(`(?i)\b(action|thriller|drama|comedy|horror|fantasy|sci-fi|romance|documentary|adventure|crime|mystery|animation|biography|western|musical|sport|war|family)\b$`).ReplaceAllString(cleanTitle, "")

	// Remove year from title if it's not a TV show (TV shows often have year in name like "All Creatures Great and Small 2020")
	if !info.IsTV {
		cleanTitle = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`).ReplaceAllString(cleanTitle, "")
	}

	// Remove quality markers and codecs
	cleanTitle = regexp.MustCompile(`(?i)\b(1080p|720p|2160p|4K|UHD|BluRay|Blu-Ray|BRRip|WEBRip|WEB-DL|WEB|HDRip|HDTV|DVDRip|BDRip|REMUX|TS|TELES|CAM|SCREENER|HC|HDR|SDR|DV|DOV|3D|HSBS|SBS|TAB|mkv|mp4|avi)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\bWEB\s+DL\b`).ReplaceAllString(cleanTitle, "")

	// Remove codec and everything after it (catches release groups like "x264-SPARKS", "H264-GROUP", etc.)
	cleanTitle = regexp.MustCompile(`(?i)\b(x264|x265|H\.?264|H\.?265|HEVC|AVC|AV1|VP9).*$`).ReplaceAllString(cleanTitle, "")

	// Remove bit depth patterns
	cleanTitle = regexp.MustCompile(`(?i)\bH\s+26[45]\b`).ReplaceAllString(cleanTitle, "") // "H 264", "H 265"
	cleanTitle = regexp.MustCompile(`(?i)\b10\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b8\s*Bit\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(10bit|8bit)\b`).ReplaceAllString(cleanTitle, "")
	// Remove leftover fragments from "10bit" becoming "10b" or "H.265" becoming "H"
	cleanTitle = regexp.MustCompile(`\b\d+[a-z]\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[Hh]\b`).ReplaceAllString(cleanTitle, "")

	// Remove file size indicators
	cleanTitle = regexp.MustCompile(`\b\d+\.?\d*\s?(GB|MB)\b`).ReplaceAllString(cleanTitle, "")

	// Remove audio formats (order matters - remove compound formats first)
	cleanTitle = regexp.MustCompile(`(?i)\bDTS-HD\.?MA\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\bDTS-HD\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\bDDPA?\d*\b`).ReplaceAllString(cleanTitle, "") // DDPA, DDP, DDP5
	cleanTitle = regexp.MustCompile(`(?i)\b(AAC\d*|AC3|DDP\d*\.?\d?|DD\d*\.?\d?|Atmos?|DTS|TrueHD|FLAC|MP3|OPUS|Dolby[A-Z]?)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b[2-8]\.?1\b`).ReplaceAllString(cleanTitle, "") // 2.1, 5.1, 7.1, etc
	cleanTitle = regexp.MustCompile(`\b[2-8]CH\b`).ReplaceAllString(cleanTitle, "")   // 6CH, 8CH
	cleanTitle = regexp.MustCompile(`(?i)\bDD[P]?\s*5\s+1\b`).ReplaceAllString(cleanTitle, "")

	// Remove release groups and source tags (including common suffixes like GalaxyRG265)
	// Note: Using (?:^|[\s._-]) to match word boundaries including hyphens
	cleanTitle = regexp.MustCompile(`(?i)(^|[\s._-])(YIFY|YTS|RARBG|RARBM|ETRG|BONE|FGT|CMRG|EVO|ION10|SPARKS|AMZN|NF|IMAX|RGB|EniaHD|Gal(axy)?(RG|UHD)?\d*|Ga|RMTeam|KyoGo|Asiimov|BYNDR|MA|FLUX|NeoNoir|NTb|DOLORES|GRACE|ETHEL|MeGusta|DARKFLiX|HiDt|CtrlHD|hallowed|PiRaTeS|MAX|YAWNTiC|PSA|PCOK|PTV|WhiskeyJack|nickarad|onlyfaffs|Lullozzo|DiRT|GPRS|TURG|DaddyCooL|Classics|NAHOM|SMU|FL|BEN)(\s|[\s._-]|$)`).ReplaceAllString(cleanTitle, " ")

	// Remove common source/platform tags
	cleanTitle = regexp.MustCompile(`(?i)\b(Netflix|AMZN|Amazon|DSNP|Disney|HBO|Hulu|Apple|Paramount)\b`).ReplaceAllString(cleanTitle, "")

	// Remove standalone plus signs and hyphens
	cleanTitle = regexp.MustCompile(`\s*[+\-]\s*`).ReplaceAllString(cleanTitle, " ")

	// Remove language names and codes
	cleanTitle = regexp.MustCompile(`(?i)\b(EN|ENG|ENGLISH|MULTI|DUAL|SUB(BED)?|DUBBED?|ITA|ITALIAN|CHINESE|KOREAN|JAPANESE|SPANISH|FRENCH|GERMAN|RUSSIAN)\b`).ReplaceAllString(cleanTitle, "")

	// Remove version tags
	cleanTitle = regexp.MustCompile(`(?i)\bV\d+\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`(?i)\b(REPACK|PROPER|EXTENDED|UNRATED|DIRECTORS?\.?CUT|DC|REMASTERED)\b`).ReplaceAllString(cleanTitle, "")

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

	// Remove collection-specific text and year ranges
	cleanTitle = regexp.MustCompile(`(?i)\s*-?\s*(the\s+)?complete\s+(collection|series|saga|trilogy|quadrilogy|anthology|boxset|box\s+set)(\s+\d+)?`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\(\d{4}\s*-\s*\d{4}\)`).ReplaceAllString(cleanTitle, "") // Year ranges like (1999-2021)
	cleanTitle = regexp.MustCompile(`\b\d+\s+movie\s+collection\b`).ReplaceAllString(cleanTitle, "")

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

// SearchBook searches Google Books API for a book by title
func (t *TMDbService) SearchBook(title string) (*GoogleBooksItem, error) {
	// Clean the title - remove file extensions and edition info that might hurt search
	cleanTitle := title
	cleanTitle = regexp.MustCompile(`\.(pdf|epub|mobi|azw3|djvu|fb2)\b`).ReplaceAllString(cleanTitle, "")
	cleanTitle = regexp.MustCompile(`\b\d+ed\b`).ReplaceAllString(cleanTitle, "") // Remove "2ed", "3ed"
	cleanTitle = strings.TrimSpace(cleanTitle)

	params := url.Values{}
	params.Add("q", cleanTitle)
	params.Add("maxResults", "1")

	searchURL := fmt.Sprintf("%s?%s", GoogleBooksAPIURL, params.Encode())

	resp, err := t.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search Google Books: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google Books API returned status %d", resp.StatusCode)
	}

	var searchResult GoogleBooksSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode Google Books response: %w", err)
	}

	if len(searchResult.Items) == 0 {
		return nil, fmt.Errorf("no book results found for '%s'", cleanTitle)
	}

	// Return the first (most relevant) result
	return &searchResult.Items[0], nil
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

// GetBookCoverURL returns the thumbnail URL for a book
func (t *TMDbService) GetBookCoverURL(book *GoogleBooksItem) string {
	if book == nil {
		return ""
	}
	// Prefer regular thumbnail over small thumbnail
	if book.VolumeInfo.ImageLinks.Thumbnail != "" {
		return book.VolumeInfo.ImageLinks.Thumbnail
	}
	return book.VolumeInfo.ImageLinks.SmallThumbnail
}

// FetchPosterForTask fetches poster URL for a download task (supports movies, TV shows, and books)
// Now with database caching for referential integrity
func (t *TMDbService) FetchPosterForTask(title string) (posterURL string, tmdbID int, err error) {
	// Extract media info (detects TV vs Movie vs Book automatically)
	mediaInfo := t.ExtractMediaInfo(title)

	// Handle books with Google Books API
	if mediaInfo.ContentType == "book" {
		log.Printf("[GoogleBooks] Searching for book: '%s'", title)
		book, err := t.SearchBook(title)
		if err != nil {
			log.Printf("[GoogleBooks] Book search failed for '%s': %v", title, err)
			return "", 0, err
		}

		coverURL := t.GetBookCoverURL(book)
		if coverURL != "" {
			log.Printf("[GoogleBooks] Found book cover for '%s': %s", title, coverURL)
			return coverURL, 0, nil // Books don't have TMDb IDs
		}
		return "", 0, fmt.Errorf("no cover found for book '%s'", title)
	}

	// Create a clean title for better matching (for movies/TV)
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
