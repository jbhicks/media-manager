package models

import (
	"time"
)

// User represents an authenticated user of the media manager
// This enables multi-user support, watch history, and personalized recommendations
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email,omitempty" gorm:"uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"not null"` // Never serialize password
	DisplayName  string    `json:"display_name,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Role         string    `json:"role" gorm:"default:'user'"` // admin, user
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WatchHistory tracks what users have watched and where they left off
type WatchHistory struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MediaType   string    `json:"media_type" gorm:"not null"` // movie, tv_episode
	MediaID     uint      `json:"media_id" gorm:"not null"`   // MovieMetadata.ID or Episode.ID
	Progress    float64   `json:"progress" gorm:"default:0"` // 0.0 to 1.0 (percentage watched)
	Position    int       `json:"position" gorm:"default:0"`   // Last position in seconds
	Duration    int       `json:"duration" gorm:"default:0"` // Total duration in seconds
	IsCompleted bool      `json:"is_completed" gorm:"default:false"`
	WatchedAt   time.Time `json:"watched_at" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Watchlist allows users to save movies/shows they want to watch later
type Watchlist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MediaType string    `json:"media_type" gorm:"not null"` // movie, tv_show
	TMDbID    int       `json:"tmdb_id" gorm:"column:tmdb_id;default:0"`
	Title     string    `json:"title"`
	PosterURL string    `json:"poster_url,omitempty"`
	AddedAt   time.Time `json:"added_at" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

// UserPreference stores user settings and preferences
type UserPreference struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	User            *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PreferredQuality string   `json:"preferred_quality,omitempty"` // 720p, 1080p, 4k
	PreferredCodec   string    `json:"preferred_codec,omitempty"`   // h264, hevc, av1
	AutoPlayNext    bool      `json:"auto_play_next" gorm:"default:true"`
	SubtitlesEnabled bool     `json:"subtitles_enabled" gorm:"default:true"`
	DefaultLanguage string    `json:"default_language,omitempty"`  // en, es, etc.
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
