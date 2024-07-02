package rpc

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/rpc/interceptor"
	"github.com/kirillgashkov/assignment-youthumb/internal/thumbnail"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewServer creates a new gRPC server.
func NewServer(cache *thumbnail.Cache, cfg *config.Config) *grpc.Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.NewUnaryServerLog(),
			interceptor.NewUnaryServerRecover(),
		),
		grpc.ChainStreamInterceptor(
			interceptor.NewStreamServerLog(),
			interceptor.NewStreamServerRecover(),
		),
	)

	if cfg.Mode == config.ModeDevelopment {
		reflection.Register(srv)
	}
	youthumbpb.RegisterThumbnailServiceServer(srv, thumbnail.NewService(cache))

	return srv
}
