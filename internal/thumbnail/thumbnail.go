package thumbnail

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a thumbnail is not found in cache or remote server.
	ErrNotFound = errors.New("thumbnail not found")
)

// Thumbnail represents a thumbnail image.
type Thumbnail struct {
	ContentType string
	Data        []byte
	Expiration  time.Time
}
