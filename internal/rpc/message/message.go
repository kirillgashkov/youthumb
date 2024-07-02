// Package message provides predefined gRPC status errors.
package message

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrStatusInternal is an internal server error.
	ErrStatusInternal = status.Errorf(codes.Internal, "internal server error")
)
