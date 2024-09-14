package thumbnail

import (
	"errors"
	"log/slog"

	"github.com/kirillgashkov/assignment-youthumb/internal/rpc/message"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxChunkSize is the max size of the chunks that are sent to the client.
	maxChunkSize = 64 * 1024
)

var (
	ErrStatusMissingVideoURL = status.Errorf(codes.InvalidArgument, "video URL is required")
	ErrStatusInvalidVideoURL = status.Errorf(codes.InvalidArgument, "video URL is invalid")
	ErrStatusNotFound        = status.Errorf(codes.NotFound, "video or thumbnail not found")
)

// Service is a thumbnail service.
type Service struct {
	youthumbpb.UnimplementedThumbnailServiceServer
	cache *Cache
}

// NewService creates a new thumbnail service.
func NewService(cache *Cache) *Service {
	return &Service{cache: cache}
}

// GetThumbnail returns a thumbnail for a given video URL.
func (s *Service) GetThumbnail(
	req *youthumbpb.GetThumbnailRequest,
	stream youthumbpb.ThumbnailService_GetThumbnailServer,
) error {
	if req.VideoUrl == "" {
		return ErrStatusMissingVideoURL
	}

	videoID, err := ParseVideoID(req.VideoUrl)
	if err != nil {
		return ErrStatusInvalidVideoURL
	}

	t, err := s.getByVideoID(videoID)
	if errors.Is(err, ErrNotFound) {
		return ErrStatusNotFound
	} else if err != nil {
		slog.Error("failed to get thumbnail", "error", err)
		return message.ErrStatusInternal
	}

	if err := send(stream, t); err != nil {
		slog.Error("failed to send thumbnail", "error", err)
		return message.ErrStatusInternal
	}

	return nil
}

// getByVideoID returns a thumbnail for a given video URL.
func (s *Service) getByVideoID(videoID string) (*Thumbnail, error) {
	t, err := s.cache.GetThumbnail(videoID)

	// Cache hit.
	if err == nil {
		return t, nil
	}

	// Error other than cache miss.
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	// Cache miss.
	thumbnailURL, err := URL(videoID)
	if err != nil {
		return nil, err
	}

	downloadedThumbnail, err := download(thumbnailURL)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetThumbnail(videoID, downloadedThumbnail); err != nil {
		slog.Error("failed to set thumbnail in cache", "error", err)
	}

	return downloadedThumbnail, nil
}

// send sends the thumbnail data to the client in chunks.
func send(stream youthumbpb.ThumbnailService_GetThumbnailServer, t *Thumbnail) error {
	contentTypeSent := false
	contentType := t.ContentType

	for i := 0; i < len(t.Data); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(t.Data) {
			end = len(t.Data)
		}

		chunkData := t.Data[i:end]
		var thumbnailChunk *youthumbpb.ThumbnailChunk

		// Include ContentType in the first chunk only.
		if !contentTypeSent {
			thumbnailChunk = &youthumbpb.ThumbnailChunk{
				Data:        chunkData,
				ContentType: contentType,
			}
			contentTypeSent = true
		} else {
			thumbnailChunk = &youthumbpb.ThumbnailChunk{Data: chunkData}
		}

		// Send the chunk to the stream.
		if err := stream.Send(thumbnailChunk); err != nil {
			return err
		}
	}

	if !contentTypeSent {
		// Send an empty chunk with ContentType if the thumbnail is empty.
		if err := stream.Send(&youthumbpb.ThumbnailChunk{ContentType: contentType}); err != nil {
			return err
		}
	}

	return nil
}
