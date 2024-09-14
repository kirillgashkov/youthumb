package thumbnail

import (
	"errors"
	"time"
)

var (
	// errNotFound is returned when a thumbnail is not found in cache or remote server.
	errNotFound = errors.New("thumbnail not found")
)

// Thumbnail represents a thumbnail image.
type Thumbnail struct {
	ContentType string
	Data        []byte
	Expiration  time.Time
}
