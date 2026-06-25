package models

import (
	"time"
)

// TVShow represents a TV series with metadata from TMDb
type TVShow struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Title           string    `json:"title" gorm:"index;not null"`
	CleanTitle      string    `json:"clean_title" gorm:"index"`
	OriginalName    string    `json:"original_name,omitempty"`
	Overview        string    `json:"overview,omitempty"`
	TMDbID          int       `json:"tmdb_id" gorm:"column:tm_db_id;uniqueIndex"`
	IMDbID          string    `json:"imdb_id,omitempty"`
	PosterURL       string    `json:"poster_url,omitempty"`
	BackdropURL     string    `json:"backdrop_url,omitempty"`
	FirstAirDate    string    `json:"first_air_date,omitempty"`
	LastAirDate     string    `json:"last_air_date,omitempty"`
	Status          string    `json:"status,omitempty"` // Returning, Ended, Canceled, etc.
	Type            string    `json:"type,omitempty"`   // Scripted, Reality, etc.
	NumberOfSeasons int       `json:"number_of_seasons,omitempty"`
	TotalEpisodes   int       `json:"total_episodes,omitempty"`
	Rating          float64   `json:"rating,omitempty"`
	VoteCount       int       `json:"vote_count,omitempty"`
	Genres          string    `json:"genres,omitempty"` // Comma-separated genre names
	Network         string    `json:"network,omitempty"`
	CreatedBy       string    `json:"created_by,omitempty"`
	ContentRating   string    `json:"content_rating,omitempty"` // TV-MA, TV-14, etc.
	Seasons         []Season  `json:"seasons,omitempty" gorm:"foreignKey:TVShowID"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Season represents a single season of a TV show
type Season struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TVShowID     uint      `json:"tv_show_id" gorm:"index;not null"`
	TVShow       *TVShow   `json:"tv_show,omitempty" gorm:"foreignKey:TVShowID"`
	SeasonNumber int       `json:"season_number" gorm:"not null"`
	Name         string    `json:"name,omitempty"`
	Overview     string    `json:"overview,omitempty"`
	TMDbID       int       `json:"tmdb_id,omitempty"`
	PosterURL    string    `json:"poster_url,omitempty"`
	AirDate      string    `json:"air_date,omitempty"`
	EpisodeCount int       `json:"episode_count,omitempty"`
	Episodes     []Episode `json:"episodes,omitempty" gorm:"foreignKey:SeasonID"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Episode represents a single episode of a TV season
type Episode struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SeasonID       uint      `json:"season_id" gorm:"index;not null"`
	Season         *Season   `json:"season,omitempty" gorm:"foreignKey:SeasonID"`
	EpisodeNumber  int       `json:"episode_number" gorm:"not null"`
	Name           string    `json:"name"`
	Overview       string    `json:"overview,omitempty"`
	TMDbID         int       `json:"tmdb_id,omitempty"`
	StillURL       string    `json:"still_url,omitempty"` // Episode thumbnail
	AirDate        string    `json:"air_date,omitempty"`
	Runtime        int       `json:"runtime,omitempty"` // Minutes
	Rating         float64   `json:"rating,omitempty"`
	VoteCount      int       `json:"vote_count,omitempty"`
	FilePath       string    `json:"file_path,omitempty"` // Local file if downloaded
	IsDownloaded   bool      `json:"is_downloaded" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CastMember represents an actor/actress in a movie or TV show
type CastMember struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TMDbID      int       `json:"tmdb_id" gorm:"uniqueIndex"`
	Name        string    `json:"name"`
	Character   string    `json:"character,omitempty"`
	ProfileURL  string    `json:"profile_url,omitempty"`
	Order       int       `json:"order" gorm:"default:0"` // Billing order
	MediaType   string    `json:"media_type"`               // movie, tv_show
	MediaID     uint      `json:"media_id" gorm:"index"`    // MovieMetadata.ID or TVShow.ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CrewMember represents a director, writer, producer, etc.
type CrewMember struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TMDbID     int       `json:"tmdb_id"`
	Name       string    `json:"name"`
	Job        string    `json:"job"`        // Director, Writer, Producer, etc.
	Department string    `json:"department"` // Directing, Writing, Production, etc.
	ProfileURL string    `json:"profile_url,omitempty"`
	MediaType  string    `json:"media_type"`            // movie, tv_show
	MediaID    uint      `json:"media_id" gorm:"index"` // MovieMetadata.ID or TVShow.ID
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Genre represents a movie/TV genre from TMDb
type Genre struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TMDbID    int       `json:"tmdb_id" gorm:"uniqueIndex"`
	Name      string    `json:"name"`
	MediaType string    `json:"media_type" gorm:"default:'movie'"` // movie, tv
	CreatedAt time.Time `json:"created_at"`
}

// Video represents a trailer or clip for a movie/TV show
type Video struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TMDbID    string    `json:"tmdb_id" gorm:"column:video_tmdb_id"` // TMDb video ID
	Name      string    `json:"name"`
	Key       string    `json:"key"`        // YouTube key
	Site      string    `json:"site"`       // YouTube, Vimeo
	Type      string    `json:"type"`       // Trailer, Teaser, Clip, Featurette
	Size      int       `json:"size"`       // 360, 480, 720, 1080
	Official  bool      `json:"official"`
	PublishedAt string  `json:"published_at,omitempty"`
	MediaType string    `json:"media_type"`            // movie, tv_show
	MediaID   uint      `json:"media_id" gorm:"index"` // MovieMetadata.ID or TVShow.ID
	CreatedAt time.Time `json:"created_at"`
}
