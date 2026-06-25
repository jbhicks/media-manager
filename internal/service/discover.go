package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TMDb Discover endpoints for browsing movies and TV shows
// These endpoints wrap the TMDb API to provide curated content

// TMDbDiscoverResult represents a paginated discover response
type TMDbDiscoverResult struct {
	Page         int                `json:"page"`
	Results      []TMDbDiscoverItem `json:"results"`
	TotalPages   int                `json:"total_pages"`
	TotalResults int                `json:"total_results"`
}

// TMDbDiscoverItem represents a movie or TV show in discover results
type TMDbDiscoverItem struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`           // For movies
	Name            string  `json:"name"`            // For TV shows
	OriginalTitle   string  `json:"original_title"`
	OriginalName    string  `json:"original_name"`
	PosterPath      string  `json:"poster_path"`
	BackdropPath    string  `json:"backdrop_path"`
	Overview        string  `json:"overview"`
	ReleaseDate     string  `json:"release_date"`    // For movies
	FirstAirDate    string  `json:"first_air_date"`  // For TV
	VoteAverage     float64 `json:"vote_average"`
	VoteCount       int     `json:"vote_count"`
	Popularity      float64 `json:"popularity"`
	GenreIDs        []int   `json:"genre_ids"`
	MediaType       string  `json:"media_type,omitempty"` // movie, tv
}

// TMDbGenreList represents the list of genres
type TMDbGenreList struct {
	Genres []TMDbGenre `json:"genres"`
}

// DiscoverEndpoints handles all TMDB discover/browse endpoints
type DiscoverEndpoints struct {
	tmdbService *TMDbService
}

func NewDiscoverEndpoints(tmdb *TMDbService) *DiscoverEndpoints {
	return &DiscoverEndpoints{tmdbService: tmdb}
}

func (d *DiscoverEndpoints) RegisterRoutes(mux *http.ServeMux) {
	// Movie discover endpoints
	mux.HandleFunc("/api/discover/movies/trending", d.handleTrendingMovies)
	mux.HandleFunc("/api/discover/movies/popular", d.handlePopularMovies)
	mux.HandleFunc("/api/discover/movies/now_playing", d.handleNowPlayingMovies)
	mux.HandleFunc("/api/discover/movies/upcoming", d.handleUpcomingMovies)
	mux.HandleFunc("/api/discover/movies/top_rated", d.handleTopRatedMovies)
	mux.HandleFunc("/api/discover/movies/by_genre", d.handleMoviesByGenre)

	// TV show discover endpoints
	mux.HandleFunc("/api/discover/tv/trending", d.handleTrendingTV)
	mux.HandleFunc("/api/discover/tv/popular", d.handlePopularTV)
	mux.HandleFunc("/api/discover/tv/on_the_air", d.handleOnTheAirTV)
	mux.HandleFunc("/api/discover/tv/airing_today", d.handleAiringTodayTV)
	mux.HandleFunc("/api/discover/tv/top_rated", d.handleTopRatedTV)
	mux.HandleFunc("/api/discover/tv/by_genre", d.handleTVByGenre)

	// Mixed endpoints
	mux.HandleFunc("/api/discover/trending", d.handleTrendingAll)
	mux.HandleFunc("/api/discover/genres", d.handleGenres)

	// Detail endpoints
	mux.HandleFunc("/api/discover/movie/", d.handleMovieDetails)
	mux.HandleFunc("/api/discover/tv/", d.handleTVDetails)
}

// Helper to fetch from TMDB API
func (d *DiscoverEndpoints) fetchTMDB(endpoint string) (*TMDbDiscoverResult, error) {
	apiKey := d.tmdbService.apiKey
	if apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY not configured")
	}
	url := fmt.Sprintf("%s%s?api_key=%s&language=en-US", TMDbAPIBaseURL, endpoint, apiKey)
	
	resp, err := d.tmdbService.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var result TMDbDiscoverResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Helper to fetch movie details from TMDB
func (d *DiscoverEndpoints) fetchTMDBMovieDetails(movieID int) (*TMDbMovieDetails, error) {
	url := fmt.Sprintf("%s/movie/%d?api_key=%s&language=en-US&append_to_response=credits,videos,similar", 
		TMDbAPIBaseURL, movieID, d.tmdbService.apiKey)
	
	resp, err := d.tmdbService.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var details TMDbMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// Helper to fetch TV details from TMDB
func (d *DiscoverEndpoints) fetchTMDBTVDetails(tvID int) (*TMDbTVDetails, error) {
	url := fmt.Sprintf("%s/tv/%d?api_key=%s&language=en-US&append_to_response=credits,videos,similar", 
		TMDbAPIBaseURL, tvID, d.tmdbService.apiKey)
	
	resp, err := d.tmdbService.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var details TMDbTVDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// Helper to get page parameter
func getPageParam(r *http.Request) int {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		return 1
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	return page
}

// === MOVIE ENDPOINTS ===

func (d *DiscoverEndpoints) handleTrendingMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/trending/movie/week?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch trending movies: %v", err), http.StatusInternalServerError)
		return
	}

	// Add media_type to each item
	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handlePopularMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/movie/popular?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch popular movies: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleNowPlayingMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/movie/now_playing?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch now playing movies: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleUpcomingMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/movie/upcoming?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch upcoming movies: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleTopRatedMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/movie/top_rated?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch top rated movies: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	genreID := r.URL.Query().Get("genre_id")
	if genreID == "" {
		jsonError(w, "genre_id parameter is required", http.StatusBadRequest)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/discover/movie?with_genres=%s&page=%d", genreID, getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch movies by genre: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "movie"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// === TV SHOW ENDPOINTS ===

func (d *DiscoverEndpoints) handleTrendingTV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/trending/tv/week?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch trending TV shows: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handlePopularTV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/tv/popular?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch popular TV shows: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleOnTheAirTV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/tv/on_the_air?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch on-the-air TV shows: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleAiringTodayTV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/tv/airing_today?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch airing today TV shows: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleTopRatedTV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/tv/top_rated?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch top rated TV shows: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleTVByGenre(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	genreID := r.URL.Query().Get("genre_id")
	if genreID == "" {
		jsonError(w, "genre_id parameter is required", http.StatusBadRequest)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/discover/tv?with_genres=%s&page=%d", genreID, getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch TV shows by genre: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range result.Results {
		result.Results[i].MediaType = "tv"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// === MIXED ENDPOINTS ===

func (d *DiscoverEndpoints) handleTrendingAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := d.fetchTMDB(fmt.Sprintf("/trending/all/week?page=%d", getPageParam(r)))
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch trending content: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (d *DiscoverEndpoints) handleGenres(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mediaType := r.URL.Query().Get("type")
	if mediaType == "" {
		mediaType = "movie"
	}

	endpoint := fmt.Sprintf("/genre/%s/list", mediaType)
	url := fmt.Sprintf("%s%s?api_key=%s&language=en-US", TMDbAPIBaseURL, endpoint, d.tmdbService.apiKey)

	resp, err := d.tmdbService.httpClient.Get(url)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch genres: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		jsonError(w, fmt.Sprintf("TMDB API returned status %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	var result TMDbGenreList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		jsonError(w, fmt.Sprintf("Failed to decode genres: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// === DETAIL ENDPOINTS ===

func (d *DiscoverEndpoints) handleMovieDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract movie ID from path: /api/discover/movie/123
	path := r.URL.Path
	prefix := "/api/discover/movie/"
	if len(path) <= len(prefix) {
		jsonError(w, "Movie ID required", http.StatusBadRequest)
		return
	}

	movieIDStr := path[len(prefix):]
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		jsonError(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	details, err := d.fetchTMDBMovieDetails(movieID)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch movie details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

func (d *DiscoverEndpoints) handleTVDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract TV ID from path: /api/discover/tv/123
	path := r.URL.Path
	prefix := "/api/discover/tv/"
	if len(path) <= len(prefix) {
		jsonError(w, "TV ID required", http.StatusBadRequest)
		return
	}

	// Handle season endpoint: /api/discover/tv/123/season/1
	pathParts := strings.Split(path[len(prefix):], "/")
	tvID, err := strconv.Atoi(pathParts[0])
	if err != nil {
		jsonError(w, "Invalid TV ID", http.StatusBadRequest)
		return
	}

	// Check if this is a season request
	if len(pathParts) >= 3 && pathParts[1] == "season" {
		seasonNum, err := strconv.Atoi(pathParts[2])
		if err != nil {
			jsonError(w, "Invalid season number", http.StatusBadRequest)
			return
		}
		d.handleSeasonDetails(w, r, tvID, seasonNum)
		return
	}

	details, err := d.fetchTMDBTVDetails(tvID)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch TV details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

func (d *DiscoverEndpoints) handleSeasonDetails(w http.ResponseWriter, r *http.Request, tvID, seasonNum int) {
	url := fmt.Sprintf("%s/tv/%d/season/%d?api_key=%s&language=en-US", 
		TMDbAPIBaseURL, tvID, seasonNum, d.tmdbService.apiKey)
	
	resp, err := d.tmdbService.httpClient.Get(url)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to fetch season details: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		jsonError(w, fmt.Sprintf("TMDB API returned status %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	var result TMDbSeasonDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		jsonError(w, fmt.Sprintf("Failed to decode season details: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// TMDbSeasonDetails represents detailed season information
type TMDbSeasonDetails struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	SeasonNumber  int              `json:"season_number"`
	Overview      string           `json:"overview"`
	AirDate       string           `json:"air_date"`
	PosterPath    string           `json:"poster_path"`
	Episodes      []TMDbEpisode    `json:"episodes"`
}

// TMDbEpisode represents a single episode
type TMDbEpisode struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	EpisodeNumber int     `json:"episode_number"`
	Overview      string  `json:"overview"`
	AirDate       string  `json:"air_date"`
	StillPath     string  `json:"still_path"`
	Runtime       int     `json:"runtime"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
}

// TMDbTVDetails represents detailed TV show information
type TMDbTVDetails struct {
	ID               int                `json:"id"`
	Name             string             `json:"name"`
	OriginalName     string             `json:"original_name"`
	Tagline          string             `json:"tagline"`
	Overview         string             `json:"overview"`
	FirstAirDate     string             `json:"first_air_date"`
	LastAirDate      string             `json:"last_air_date"`
	Status           string             `json:"status"`
	Type             string             `json:"type"`
	NumberOfSeasons  int                `json:"number_of_seasons"`
	NumberOfEpisodes int                `json:"number_of_episodes"`
	VoteAverage      float64            `json:"vote_average"`
	VoteCount        int                `json:"vote_count"`
	PosterPath       string             `json:"poster_path"`
	BackdropPath     string             `json:"backdrop_path"`
	Genres           []TMDbGenre        `json:"genres"`
	Credits          TMDbCredits        `json:"credits"`
	Videos           TMDbVideos         `json:"videos"`
	Similar          TMDbSimilarResult  `json:"similar"`
	Seasons          []TMDbSeasonInfo   `json:"seasons"`
	Networks         []TMDbNetwork      `json:"networks"`
	CreatedBy        []TMDbCreator      `json:"created_by"`
	ContentRatings   TMDbContentRatings `json:"content_ratings"`
	EpisodeRunTime   []int              `json:"episode_run_time"`
}

// TMDbVideos represents video results (trailers, etc.)
type TMDbVideos struct {
	Results []TMDbVideoResult `json:"results"`
}

type TMDbVideoResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Site        string `json:"site"`
	Type        string `json:"type"`
	Size        int    `json:"size"`
	Official    bool   `json:"official"`
	PublishedAt string `json:"published_at"`
}

// TMDbSimilarResult represents similar movies/shows
type TMDbSimilarResult struct {
	Page         int                `json:"page"`
	Results      []TMDbDiscoverItem `json:"results"`
	TotalPages   int                `json:"total_pages"`
	TotalResults int                `json:"total_results"`
}

// TMDbSeasonInfo represents a season summary
type TMDbSeasonInfo struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
	PosterPath   string `json:"poster_path"`
	Overview     string `json:"overview"`
}

// TMDbNetwork represents a TV network
type TMDbNetwork struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// TMDbCreator represents a show creator
type TMDbCreator struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
}

// TMDbContentRatings represents content rating certifications
type TMDbContentRatings struct {
	Results []TMDbContentRating `json:"results"`
}

type TMDbContentRating struct {
	ISO3166_1   string `json:"iso_3166_1"`
	Rating      string `json:"rating"`
	Certification string `json:"certification"`
}

// Cache for discover results (5 minute TTL)
var discoverCache = make(map[string]cacheEntry)
var discoverCacheMutex sync.RWMutex

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

func getCachedDiscover(key string) (interface{}, bool) {
	discoverCacheMutex.RLock()
	defer discoverCacheMutex.RUnlock()
	entry, ok := discoverCache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func setCachedDiscover(key string, data interface{}) {
	discoverCacheMutex.Lock()
	defer discoverCacheMutex.Unlock()
	discoverCache[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
}
