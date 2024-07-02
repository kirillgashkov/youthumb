package message

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Predefined gRPC errors that are transmitted to the client.
var (
	ErrGRPCInternal = status.Errorf(codes.Internal, "internal server error")
)
