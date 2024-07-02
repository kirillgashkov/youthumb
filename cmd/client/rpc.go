package main

import (
	"net"
	"strconv"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"

	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// newThumbnailServiceClient creates a new gRPC client.
func newThumbnailServiceClient(conn *grpc.ClientConn) (youthumbpb.ThumbnailServiceClient, error) {
	return youthumbpb.NewThumbnailServiceClient(conn), nil
}

// newClient creates a new gRPC client connection. Caller is responsible for
// closing the connection.
func newClient(cfg config.GRPCConfig) (*grpc.ClientConn, error) {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
