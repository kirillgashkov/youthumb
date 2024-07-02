package thumbnail

import "errors"

var (
	// errNotFound is returned when a thumbnail is not found.
	errNotFound = errors.New("thumbnail not found")
)

// Thumbnail represents a thumbnail image.
type Thumbnail struct {
	ContentType string
	Data        []byte
}
