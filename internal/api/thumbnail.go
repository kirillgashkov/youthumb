package api

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/api/errors"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
)

// chunkSize is the size of the chunks that are sent to the client. It is chosen
// to be 3 MB to fit within 4 MB gRPC message size limit.
const chunkSize = 3 * 1024 * 1024

type thumbnailServiceServer struct {
	youthumbpb.UnimplementedThumbnailServiceServer
}

func (*thumbnailServiceServer) GetThumbnail(req *youthumbpb.GetThumbnailRequest, stream youthumbpb.ThumbnailService_GetThumbnailServer) error {
	if req.VideoUrl == "" {
		return status.Errorf(codes.InvalidArgument, "video URL is required")
	}

	resp, err := http.Get(req.VideoUrl)
	if err != nil {
		return errors.ErrGRPCInternal
	}
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return status.Errorf(codes.NotFound, "video not found")
		}
		return errors.ErrGRPCInternal
	}

	if err := stream.Send(&youthumbpb.ThumbnailChunk{ContentType: resp.Header.Get("Content-Type")}); err != nil {
		return errors.ErrGRPCInternal
	}

	buf := make([]byte, chunkSize)
	for {
		slog.Debug("reading chunk")
		n, err := resp.Body.Read(buf)
		if err == io.EOF {
			slog.Debug("no more data to read")
			break
		}
		if err != nil {
			slog.Debug("failed to read chunk", "error", err)
			return errors.ErrGRPCInternal
		}

		if n == 0 {
			slog.Debug("no more data to read")
			break
		}

		if err := stream.Send(&youthumbpb.ThumbnailChunk{Data: buf[:n]}); err != nil {
			slog.Debug("failed to send chunk", "error", err)
			return errors.ErrGRPCInternal
		}

		buf = make([]byte, chunkSize)
	}

	return nil
}
