package interceptors

import (
	"context"
	"errors"
	"github.com/MxTrap/metrics/internal/server/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestStatusErrorInterceptorSuccess(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := StatusErrorInterceptor
	resp, err := interceptor(ctx, "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestStatusErrorInterceptorNotFound(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, models.ErrNotFoundMetric
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := StatusErrorInterceptor
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, s.Code())
}

func TestStatusErrorInterceptorUnknownMetricType(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, models.ErrUnknownMetricType
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := StatusErrorInterceptor
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
}

func TestStatusErrorInterceptorWrongMetricValue(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, models.ErrWrongMetricValue
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := StatusErrorInterceptor
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
}

func TestStatusErrorInterceptorOtherError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("unknown error")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := StatusErrorInterceptor
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, s.Code())
}
