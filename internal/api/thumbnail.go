package api

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/api/errs"
	"github.com/kirillgashkov/assignment-youthumb/internal/cache"
	"github.com/kirillgashkov/assignment-youthumb/internal/youtube"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
)

// maxChunkSize is the max size of the chunks that are sent to the client.
const maxChunkSize = 64 * 1024

type thumbnailServiceServer struct {
	youthumbpb.UnimplementedThumbnailServiceServer
	cache *cache.Cache
}

func (s *thumbnailServiceServer) GetThumbnail(req *youthumbpb.GetThumbnailRequest, stream youthumbpb.ThumbnailService_GetThumbnailServer) error {
	if req.VideoUrl == "" {
		return status.Errorf(codes.InvalidArgument, "video URL is required")
	}

	thumbnailURL, err := youtube.ThumbnailURLFromVideoURL(req.VideoUrl)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "video URL is invalid")
	}

	resp, err := http.Get(thumbnailURL)
	if err != nil {
		return errs.ErrGRPCInternal
	}
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return status.Errorf(codes.NotFound, "video or thumbnail not found")
		}
		return errs.ErrGRPCInternal
	}

	if err := stream.Send(&youthumbpb.ThumbnailChunk{ContentType: resp.Header.Get("Content-Type")}); err != nil {
		return errs.ErrGRPCInternal
	}

	buf := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errs.ErrGRPCInternal
		}

		if err := stream.Send(&youthumbpb.ThumbnailChunk{Data: buf[:n]}); err != nil {
			return errs.ErrGRPCInternal
		}

		buf = make([]byte, maxChunkSize)
	}

	return nil
}
