package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"log/slog"
)

// NewUnaryLog returns a new unary server interceptor that logs completed unary
// RPCs.
func NewUnaryLog() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		msg, err := handler(ctx, req)
		if err != nil {
			slog.Error("unary", "method", info.FullMethod, "error", err)
		} else {
			slog.Info("unary", "method", info.FullMethod)
		}

		return msg, err
	}
}

// NewStreamLog returns a new stream server interceptor that logs completed
// streaming RPCs.
func NewStreamLog() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			slog.Error("stream", "method", info.FullMethod, "error", err)
		} else {
			slog.Info("stream", "method", info.FullMethod)
		}

		return err
	}
}
