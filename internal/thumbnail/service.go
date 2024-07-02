package thumbnail

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kirillgashkov/assignment-youthumb/internal/rpc/message"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxChunkSize is the max size of the chunks that are sent to the client.
	maxChunkSize = 64 * 1024
)

type Service struct {
	youthumbpb.UnimplementedThumbnailServiceServer
	cache *Cache
}

func NewService(cache *Cache) *Service {
	return &Service{cache: cache}
}

func (s *Service) GetThumbnail(req *youthumbpb.GetThumbnailRequest, stream youthumbpb.ThumbnailService_GetThumbnailServer) error {
	if req.VideoUrl == "" {
		return status.Errorf(codes.InvalidArgument, "video URL is required")
	}

	videoID, err := ParseVideoID(req.VideoUrl)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "video URL is invalid")
	}
	thumbnailURL, err := URLFromVideoURL(req.VideoUrl)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "video URL is invalid")
	}

	t, err := s.cache.GetThumbnail(videoID)
	if errors.Is(err, errNotFound) {
		downloadedThumbnail, expirationTime, err := downloadThumbnail(thumbnailURL)
		if err != nil {
			if errors.Is(err, errNotFound) {
				return status.Errorf(codes.NotFound, "video or thumbnail not found")
			}
			return message.ErrStatusInternal
		}
		t = downloadedThumbnail

		if err := s.cache.SetThumbnail(videoID, downloadedThumbnail, expirationTime); err != nil {
			slog.Error("failed to set thumbnail in cache", "error", err)
		}
	} else if err != nil {
		return message.ErrStatusInternal
	}

	contentTypeSent := false
	for i := 0; i < len(t.Data); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(t.Data) {
			end = len(t.Data)
		}

		chunkData := t.Data[i:end]
		var thumbnailChunk *youthumbpb.ThumbnailChunk
		if !contentTypeSent {
			thumbnailChunk = &youthumbpb.ThumbnailChunk{Data: chunkData, ContentType: t.ContentType}
			contentTypeSent = true
		} else {
			thumbnailChunk = &youthumbpb.ThumbnailChunk{Data: chunkData}
		}

		if err := stream.Send(thumbnailChunk); err != nil {
			return err
		}
	}

	return nil
}

func downloadThumbnail(thumbnailURL string) (*Thumbnail, time.Time, error) {
	resp, err := http.Get(thumbnailURL)
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
