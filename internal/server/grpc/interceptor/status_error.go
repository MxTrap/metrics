package interceptors

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/server/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StatusErrorInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}
	if errors.Is(err, models.ErrNotFoundMetric) {
		return nil, status.Error(codes.NotFound, "")
	}
	if errors.Is(err, models.ErrUnknownMetricType) {
		return nil, status.Error(codes.FailedPrecondition, "")
	}
	if errors.Is(err, models.ErrWrongMetricValue) {
		return nil, status.Error(codes.FailedPrecondition, "")
	}

	return nil, status.Error(codes.Internal, "")
}
