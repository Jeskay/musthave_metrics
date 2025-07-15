package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewLoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		t := time.Now()
		resp, err = handler(ctx, req)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}
		logger.Info(
			"incoming request",
			slog.String("uri", md["uri"][0]),
			slog.String("method", info.FullMethod),
			slog.Duration("latency", time.Since(t)),
		)
		if err != nil {
			logger.Info(
				"response",
				slog.String("error", err.Error()),
			)
		}
		return resp, err
	}
}
