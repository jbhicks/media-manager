package models

import (
	"time"
)

type DownloadSource struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	APIKey      string    `json:"api_key,omitempty"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	Priority    int       `json:"priority" gorm:"default:0"`
	LastChecked time.Time `json:"last_checked"`
	LastError   string    `json:"last_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DownloadRule struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Name               string    `json:"name"`
	Enabled            bool      `json:"enabled" gorm:"default:true"`
	MediaType          string    `json:"media_type"`
	SearchQuery        string    `json:"search_query,omitempty"`
	Quality            string    `json:"quality,omitempty"`
	Resolution         string    `json:"resolution,omitempty"`
	MinSeeders         int       `json:"min_seeders" gorm:"default:0"`
	MaxSeeders         int       `json:"max_seeders,omitempty"`
	MinSize            int64     `json:"min_size,omitempty"`
	MaxSize            int64     `json:"max_size,omitempty"`
	MinUploadAge       int       `json:"min_upload_age,omitempty"`
	MaxUploadAge       int       `json:"max_upload_age,omitempty"`
	SortBy             string    `json:"sort_by" gorm:"default:'seeders'"`
	MaxResults         int       `json:"max_results" gorm:"default:1"`
	MaxResultsPerTitle int       `json:"max_results_per_title" gorm:"default:1"`
	AutoDownload       bool      `json:"auto_download" gorm:"default:false"`
	DestinationPath    string    `json:"destination_path"`
	Schedule           string    `json:"schedule,omitempty"`
	LastRun            time.Time `json:"last_run"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DownloadTask struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	RuleID       uint            `json:"rule_id,omitempty"`
	Rule         *DownloadRule   `json:"rule,omitempty" gorm:"foreignKey:RuleID"`
	SourceID     uint            `json:"source_id,omitempty"`
	Source       *DownloadSource `json:"source,omitempty" gorm:"foreignKey:SourceID"`
	Title        string          `json:"title"`
	InfoHash     string          `json:"info_hash,omitempty" gorm:"index"`
	MagnetLink   string          `json:"magnet_link,omitempty"`
	TorrentURL   string          `json:"torrent_url,omitempty"`
	TorrentID    int             `json:"torrent_id,omitempty"`
	Size         int64           `json:"size"`
	Seeders      int             `json:"seeders"`
	Leechers     int             `json:"leechers"`
	Status       string          `json:"status" gorm:"default:'pending'"`
	Progress     float64         `json:"progress" gorm:"default:0"`
	DownloadPath string          `json:"download_path,omitempty"`
	PosterURL    string          `json:"poster_url,omitempty"`
	TMDbID       int             `json:"tmdb_id,omitempty"`
	Error        string          `json:"error,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type SearchResult struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	SourceID   uint            `json:"source_id"`
	Source     *DownloadSource `json:"source,omitempty" gorm:"foreignKey:SourceID"`
	Title      string          `json:"title"`
	InfoHash   string          `json:"info_hash,omitempty" gorm:"index"`
	MagnetLink string          `json:"magnet_link,omitempty"`
	TorrentURL string          `json:"torrent_url,omitempty"`
	Size       int64           `json:"size"`
	Seeders    int             `json:"seeders"`
	Leechers   int             `json:"leechers"`
	Category   string          `json:"category,omitempty"`
	UploadDate time.Time       `json:"upload_date,omitempty"`
	ExpiresAt  time.Time       `json:"expires_at"`
	CreatedAt  time.Time       `json:"created_at"`
}

type DownloadSuggestion struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	SourceID     uint            `json:"source_id"`
	Source       *DownloadSource `json:"source,omitempty" gorm:"foreignKey:SourceID"`
	Title        string          `json:"title" gorm:"index"`
	InfoHash     string          `json:"info_hash" gorm:"index;uniqueIndex"`
	MagnetLink   string          `json:"magnet_link,omitempty"`
	TorrentURL   string          `json:"torrent_url,omitempty"`
	Size         int64           `json:"size"`
	Seeders      int             `json:"seeders" gorm:"index"`
	Leechers     int             `json:"leechers"`
	Category     string          `json:"category,omitempty"`
	Resolution   string          `json:"resolution,omitempty"`
	Quality      string          `json:"quality,omitempty"`
	UploadDate   time.Time       `json:"upload_date,omitempty" gorm:"index"`
	PosterURL    string          `json:"poster_url,omitempty"`
	TMDbID       int             `json:"tmdb_id,omitempty"`
	Status       string          `json:"status" gorm:"default:'pending';index"`
	RejectedAt   *time.Time      `json:"rejected_at,omitempty"`
	ApprovedAt   *time.Time      `json:"approved_at,omitempty"`
	DownloadedAt *time.Time      `json:"downloaded_at,omitempty"`
	Notes        string          `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type DownloadHistory struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	InfoHash     string    `json:"info_hash" gorm:"index;uniqueIndex"`
	Title        string    `json:"title"`
	Size         int64     `json:"size"`
	DownloadedAt time.Time `json:"downloaded_at"`
	DeletedAt    time.Time `json:"deleted_at"`
	Reason       string    `json:"reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
