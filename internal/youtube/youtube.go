package youtube

import (
	"fmt"
	"net/url"
)

// ThumbnailURLFromVideoURL returns a URL of a thumbnail for a given YouTube
// video URL.
func ThumbnailURLFromVideoURL(videoURL string) (string, error) {
	videoID, err := parseVideoID(videoURL)
	if err != nil {
		return "", err
	}

	u, err := thumbnailURL(videoID)
	return u, err
}

func parseVideoID(videoURL string) (string, error) {
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

func thumbnailURL(videoID string) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("video ID is required")
	}
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID), nil
}
