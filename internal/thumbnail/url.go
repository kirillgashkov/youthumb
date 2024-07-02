package thumbnail

import (
	"fmt"
	"net/url"
)

// URLFromVideoURL returns a URL of a thumbnail for a given YouTube video URL.
func URLFromVideoURL(videoURL string) (string, error) {
	videoID, err := ParseVideoID(videoURL)
	if err != nil {
		return "", err
	}

	u, err := URL(videoID)
	return u, err
}

// ParseVideoID extracts a video ID from a YouTube video URL.
func ParseVideoID(videoURL string) (string, error) {
	u, err := url.Parse(videoURL)
	if err != nil {
		return "", err
	}

	switch u.Host {
	case "www.youtube.com", "youtube.com":
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return "", err
		}
		return q.Get("v"), nil
	case "youtu.be":
		return u.Path[1:], nil
	}

	return "", fmt.Errorf("unknown video URL: %s", videoURL)
}

// URL returns a URL of a thumbnail for a given YouTube video ID.
func URL(videoID string) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("video ID is required")
	}
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", videoID), nil
}
