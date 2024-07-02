// Package cache provides a key-value cache for thumbnails with expiration.
// The cache is backed by an SQLite database.
package cache

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Thumbnail represents a thumbnail image.
type Thumbnail struct {
	ContentType string
	Data        []byte
}

// Cache is a key-value cache for thumbnails with expiration.
type Cache struct {
	// Path to the SQLite database file.
	Path string
	db   *sql.DB
}

// Open opens the cache at the given path. The caller is responsible for
// closing the cache.
func Open(path string) (*Cache, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS cache (
			video_id TEXT PRIMARY KEY,
			content_type TEXT NOT NULL,
			data BLOB NOT NULL,
			expires_at INTEGER NOT NULL
		)
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, err
	}

	return &Cache{Path: path, db: db, getStmt: getStmt, setStmt: setStmt}, nil
}

// Close closes the cache.
func (c *Cache) Close() error {
	return c.db.Close()
}

// Get returns the thumbnail for the given video ID from the cache.
func (c *Cache) Get(videoID string) (*Thumbnail, error) {
	query := `SELECT content_type, data FROM cache WHERE video_id = ? AND expires_at > ?`

	var contentType string
	var data []byte
	err := c.db.QueryRow(query, videoID, time.Now().Unix()).Scan(&contentType, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &Thumbnail{ContentType: contentType, Data: data}, nil
}

// Set sets the thumbnail for the given video ID in the cache with the given
// expiration time.
func (c *Cache) Set(videoID string, thumbnail *Thumbnail, expiration time.Time) error {
	query := `INSERT OR REPLACE INTO cache (video_id, content_type, data, expires_at) VALUES (?, ?, ?, ?)`
	_, err := c.db.Exec(query, videoID, thumbnail.ContentType, thumbnail.Data, expiration.Unix())
	return err
}
