package thumbnail

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Cache is a cache for thumbnail images.
type Cache struct {
	// db is the SQLite database connection pool.
	db *sql.DB
}

// OpenCache opens a new cache.
// The given DSN must be a SQLite DSN.
func OpenCache(dsn string) (*Cache, error) {
	db, err := sql.Open("sqlite3", dsn)
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

	return &Cache{db: db}, nil
}

// Close closes the cache.
func (c *Cache) Close() error {
	return c.db.Close()
}

// GetThumbnail returns a thumbnail from the cache.
// If the thumbnail is not found in the cache, it returns ErrNotFound.
func (c *Cache) GetThumbnail(videoID string) (*Thumbnail, error) {
	query := `SELECT content_type, data, expires_at FROM cache WHERE video_id = ? AND expires_at > ?`
	row := c.db.QueryRow(query, videoID, time.Now().Unix())

	var contentType string
	var data []byte
	var expiration int64
	err := row.Scan(&contentType, &data, &expiration)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	t := &Thumbnail{
		ContentType: contentType,
		Data:        data,
		Expiration:  time.Unix(expiration, 0),
	}
	return t, nil
}

// SetThumbnail sets a thumbnail in the cache.
func (c *Cache) SetThumbnail(videoID string, t *Thumbnail) error {
	query := `INSERT OR REPLACE INTO cache (video_id, content_type, data, expires_at) VALUES (?, ?, ?, ?)`

	if _, err := c.db.Exec(query, videoID, t.ContentType, t.Data, t.Expiration.Unix()); err != nil {
		return err
	}

	return nil
}
