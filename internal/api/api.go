package api

import (
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ThumbnailServiceServer struct {
	youthumbpb.UnimplementedThumbnailServiceServer
}

func (*ThumbnailServiceServer) GetThumbnail(*youthumbpb.GetThumbnailRequest, youthumbpb.ThumbnailService_GetThumbnailServer) error {
	return status.Errorf(codes.Unimplemented, "method GetThumbnail not implemented")
}
