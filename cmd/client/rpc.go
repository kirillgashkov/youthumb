package main

import (
	"net/netip"

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
	addr, err := netip.ParseAddr(cfg.Host)
	if err != nil {
		return nil, err
	}
	addrPort := netip.AddrPortFrom(addr, uint16(cfg.Port))

	conn, err := grpc.NewClient(addrPort.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
