package api

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/api/interceptor"
	"github.com/kirillgashkov/assignment-youthumb/internal/config"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(cfg *config.Config) *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.NewUnaryRecover(),
			interceptor.NewUnaryLog(),
		),
		grpc.ChainStreamInterceptor(
			interceptor.NewStreamRecover(),
			interceptor.NewStreamLog(),
		),
	)

	if cfg.Mode == config.ModeDevelopment {
		reflection.Register(srv)
	}
	youthumbpb.RegisterThumbnailServiceServer(srv, &thumbnailServiceServer{})

	return srv
}
