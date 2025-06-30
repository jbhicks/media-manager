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
