package db

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/user/media-manager/pkg/models"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate database schema
	err = db.AutoMigrate(&models.MediaFile{}, &models.Tag{}, &models.Folder{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) GetMediaFiles(limit, offset int) ([]models.MediaFile, error) {
	var files []models.MediaFile
	err := d.db.Preload("Tags").Limit(limit).Offset(offset).Find(&files).Error
	return files, err
}

func (d *Database) CreateMediaFile(file *models.MediaFile) error {
	return d.db.Create(file).Error
}

func (d *Database) GetTags() ([]models.Tag, error) {
	var tags []models.Tag
	err := d.db.Find(&tags).Error
	return tags, err
}

func (d *Database) CreateFolder(folder *models.Folder) error {
	return d.db.Create(folder).Error
}

func (d *Database) CreateTag(tag *models.Tag) error {
	return d.db.Create(tag).Error
}

func (d *Database) DeleteMediaFilesByDirectory(dirPath string) error {
	return d.db.Where("path LIKE ?", dirPath+"%").Delete(&models.MediaFile{}).Error
}

func (d *Database) Close() error {
	db, err := d.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
