package api

import (
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type thumbnailServiceServer struct {
	youthumbpb.UnimplementedThumbnailServiceServer
}

func (*thumbnailServiceServer) GetThumbnail(*youthumbpb.GetThumbnailRequest, youthumbpb.ThumbnailService_GetThumbnailServer) error {
	return status.Errorf(codes.Unimplemented, "method GetThumbnail not implemented")
}
