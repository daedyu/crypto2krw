package grpc

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	elapsed := time.Since(start)

	code := status.Code(err)
	slog.Info("grpc call",
		"method", info.FullMethod,
		"duration_ms", elapsed.Milliseconds(),
		"code", code.String(),
	)

	return resp, err
}
