package thumbnail

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// download downloads a thumbnail from a given URL.
func download(url string) (*Thumbnail, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return fromResponse(resp)
}

// fromResponse creates a Thumbnail from an HTTP response.
//
// The response must be successful (status code 200).
func fromResponse(resp *http.Response) (*Thumbnail, error) {
	expiration, err := time.Parse(time.RFC1123, resp.Header.Get("Expires"))
	if err != nil {
		return nil, err
	}

	data := bytes.NewBuffer(nil)
	if _, err := io.Copy(data, resp.Body); err != nil {
		return nil, err
	}

	t := &Thumbnail{
		ContentType: resp.Header.Get("Content-Type"),
		Data:        data.Bytes(),
		Expiration:  expiration,
	}
	return t, nil
}
