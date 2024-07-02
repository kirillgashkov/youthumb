package thumbnail

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// download downloads a thumbnail from a given URL.
func download(url string) (*Thumbnail, time.Time, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, time.Time{}, errNotFound
		}
		return nil, time.Time{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	expiresHeader := resp.Header.Get("Expires")
	expirationTime, err := time.Parse(time.RFC1123, expiresHeader)
	if err != nil {
		return nil, time.Time{}, err
	}

	sb := &strings.Builder{}
	if _, err := io.Copy(sb, resp.Body); err != nil {
		return nil, time.Time{}, err
	}

	t := &Thumbnail{
		ContentType: resp.Header.Get("Content-Type"), Data: []byte(sb.String()),
	}

	return t, expirationTime, nil
}
