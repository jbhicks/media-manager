package models

import (
	"time"
)

type MediaFile struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Path        string    `json:"path" gorm:"uniqueIndex"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	FileType    string    `json:"file_type"` // image, video
	MimeType    string    `json:"mime_type"`
	PreviewPath string    `json:"preview_path"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Duration    int       `json:"duration"` // for videos, in seconds
	Tags        []Tag     `json:"tags" gorm:"many2many:file_tags;"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"uniqueIndex"`
	Color string `json:"color"` // hex color for UI
}

type Folder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Path        string    `json:"path" gorm:"uniqueIndex"`
	Name        string    `json:"name"`
	LastScanned time.Time `json:"last_scanned"`
	FileCount   int       `json:"file_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type ServiceConfig struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	DownloadEnabled        bool      `json:"download_enabled" gorm:"default:false"`
	ScheduleInterval       int       `json:"schedule_interval" gorm:"default:3600"`
	MaxConcurrentDownloads int       `json:"max_concurrent_downloads" gorm:"default:5"`
	TorrentClientType      string    `json:"torrent_client_type" gorm:"default:'transmission'"`
	TorrentClientHost      string    `json:"torrent_client_host" gorm:"default:'localhost:9091'"`
	TorrentClientUser      string    `json:"torrent_client_user,omitempty"`
	TorrentClientPass      string    `json:"torrent_client_pass,omitempty"`
	JellyfinURL            string    `json:"jellyfin_url" gorm:"default:''"`
	JellyfinAPIKey         string    `json:"jellyfin_api_key,omitempty"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// MovieMetadata stores cached TMDb metadata for movies
// This is shared across library, downloads, and suggestions
type MovieMetadata struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"index"`                         // Original title for matching
	CleanTitle  string    `json:"clean_title" gorm:"index"`                   // Normalized title for better matching
	Year        int       `json:"year,omitempty"`                             // Release year
	TMDbID      int       `json:"tmdb_id" gorm:"column:tm_db_id;uniqueIndex"` // TMDb movie ID
	PosterURL   string    `json:"poster_url,omitempty"`                       // Full poster URL from TMDb
	BackdropURL string    `json:"backdrop_url,omitempty"`                     // Backdrop/banner URL
	Overview    string    `json:"overview,omitempty"`                         // Movie description
	Rating      float64   `json:"rating,omitempty"`                           // TMDb rating
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
