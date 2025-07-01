package interceptors

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"testing"
)

func TestIPValidatorEmptyCIDR(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := IPValidator("")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestIPValidatorMissingMetadata(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := context.Background()

	interceptor := IPValidator("192.168.1.0/24")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}

func TestIPValidatorMissingXRealIP(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{}))

	interceptor := IPValidator("192.168.1.0/24")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}

func TestIPValidatorInvalidXRealIP(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	md := metadata.New(map[string]string{"X-Real-Ip": "invalid-ip"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := IPValidator("192.168.1.0/24")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}

func TestIPValidatorInvalidCIDR(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	md := metadata.New(map[string]string{"X-Real-Ip": "192.168.1.100"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := IPValidator("invalid-cidr")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	_, ok := status.FromError(err)
	assert.False(t, ok) // Ошибка парсинга CIDR не является gRPC-ошибкой
}

func TestIPValidatorIPInCIDR(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	md := metadata.New(map[string]string{"X-Real-Ip": "192.168.1.100"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := IPValidator("192.168.1.0/24")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestIPValidatorIPNotInCIDR(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
	md := metadata.New(map[string]string{"X-Real-Ip": "10.0.0.100"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := IPValidator("192.168.1.0/24")
	resp, err := interceptor(ctx, "request", info, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, s.Code())
}
