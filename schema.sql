-- Database Schema for Media Manager

-- Media files table
CREATE TABLE media_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    filename TEXT NOT NULL,
    size INTEGER NOT NULL,
    mod_time DATETIME NOT NULL,
    file_type TEXT NOT NULL, -- 'image' or 'video'
    mime_type TEXT NOT NULL,
    preview_path TEXT,
    width INTEGER,
    height INTEGER,
    duration INTEGER, -- for videos, in seconds
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tags table
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    color TEXT DEFAULT '#007bff', -- hex color for UI
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Many-to-many relationship between files and tags
CREATE TABLE file_tags (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (file_id, tag_id),
    FOREIGN KEY (file_id) REFERENCES media_files(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Folders table for organizing scanned directories
CREATE TABLE folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    last_scanned DATETIME,
    file_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_media_files_type ON media_files(file_type);
CREATE INDEX idx_media_files_mod_time ON media_files(mod_time);
CREATE INDEX idx_media_files_size ON media_files(size);
CREATE INDEX idx_file_tags_file_id ON file_tags(file_id);
CREATE INDEX idx_file_tags_tag_id ON file_tags(tag_id);
CREATE INDEX idx_folders_path ON folders(path);