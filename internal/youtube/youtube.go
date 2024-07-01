package youtube

import (
	"fmt"
	"net/url"
)

func ThumbnailURLFromVideoURL(videoURL string) (string, error) {
	videoID, err := parseVideoID(videoURL)
	if err != nil {
		return "", err
	}
	return thumbnailURL(videoID), nil
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

func thumbnailURL(videoID string) string {
	if videoID == "" {
		return ""
	}
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID)
}
