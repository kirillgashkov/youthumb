package interceptor

import (
	"context"
	"github.com/kirillgashkov/assignment-youthumb/internal/api/errs"
	"google.golang.org/grpc"
	"log/slog"
	"runtime/debug"
)

// NewUnaryServerRecover returns a new unary server interceptor that recovers
// from panics.
func NewUnaryServerRecover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("unary caught panic", "method", info.FullMethod, "panic", p, "stack", string(debug.Stack()))
				resp, err = nil, errs.ErrGRPCInternal
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

// NewStreamServerRecover returns a new stream server interceptor that recovers
// from panics. Message receive and send operations are not recovered.
func NewStreamServerRecover() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("stream caught panic", "method", info.FullMethod, "panic", p, "stack", string(debug.Stack()))
				err = errs.ErrGRPCInternal
			}
		}()
		err = handler(srv, ss)
		return
	}
}
