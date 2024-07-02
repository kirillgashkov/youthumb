// Package cache provides a key-value cache for thumbnails with expiration.
// The cache is backed by an SQLite database.
package cache

import (
	_ "github.com/mattn/go-sqlite3"
	"time"
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
}

// New creates a new Cache with the given path to the SQLite database file.
func New(path string) *Cache {
	return &Cache{Path: path}
}

// Get returns the thumbnail for the given video ID from the cache.
func (c *Cache) Get(videoID string) ([]bool, error) {
	return nil, nil
}

// Set sets the thumbnail for the given video ID in the cache with the given
// expiration time.
func (c *Cache) Set(videoID string, thumbnail *Thumbnail, expireTime time.Time) error {
	return nil
}
