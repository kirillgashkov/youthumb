package interceptor

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"runtime/debug"
)

// NewUnaryRecover returns a new unary server interceptor that recovers from
// panics.
func NewUnaryRecover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("unary recovered from panic", "method", info.FullMethod, "panic", p, "stack", string(debug.Stack()))
				resp, err = nil, status.Error(codes.Internal, "internal server error")
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

type recoverStream struct {
	grpc.ServerStream
}

func (s *recoverStream) RecvMsg(m any) (err error) {
	defer func() {
		if p := recover(); p != nil {
			slog.Error("stream recovered from panic", "panic", p, "stack", string(debug.Stack()))
			err = status.Error(codes.Internal, "internal server error")
		}
	}()
	err = s.ServerStream.RecvMsg(m)
	return
}

func (s *recoverStream) SendMsg(m any) (err error) {
	defer func() {
		if p := recover(); p != nil {
			slog.Error("stream recovered from panic", "panic", p, "stack", string(debug.Stack()))
			err = status.Error(codes.Internal, "internal server error")
		}
	}()
	err = s.ServerStream.SendMsg(m)
	return
}

// NewStreamRecover returns a new stream server interceptor that recovers from
// panics.
func NewStreamRecover() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("stream recovered from panic", "method", info.FullMethod, "panic", p, "stack", string(debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		err = handler(srv, &recoverStream{ss})
		return
	}
}
