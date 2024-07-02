package thumbnail

import (
	"errors"
	"github.com/kirillgashkov/assignment-youthumb/internal/rpc/message"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

const (
	// maxChunkSize is the max size of the chunks that are sent to the client.
	maxChunkSize = 64 * 1024
)

var (
	errStatusMissingVideoURL = status.Errorf(codes.InvalidArgument, "video URL is required")
	errStatusInvalidVideoURL = status.Errorf(codes.InvalidArgument, "video URL is invalid")
	errStatusNotFound        = status.Errorf(codes.NotFound, "video or thumbnail not found")
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
		return errStatusMissingVideoURL
	}

	videoID, err := ParseVideoID(req.VideoUrl)
	if err != nil {
		return errStatusInvalidVideoURL
	}
	thumbnailURL, err := URLFromVideoURL(req.VideoUrl)
	if err != nil {
		return errStatusInvalidVideoURL
	}

	t, err := s.cache.GetThumbnail(videoID)
	if errors.Is(err, errNotFound) {
		downloadedThumbnail, expirationTime, err := download(thumbnailURL)
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

	if err := send(stream, t); err != nil {
		return message.ErrStatusInternal
	}

	return nil
}

// send sends the thumbnail data to the client in chunks.
func send(stream youthumbpb.ThumbnailService_GetThumbnailServer, t *Thumbnail) error {
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
