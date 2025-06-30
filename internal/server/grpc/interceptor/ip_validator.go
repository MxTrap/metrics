package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

func IpValidator(cidr string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if cidr == "" {
			return handler(ctx, req)
		}

		errPermissionDenied := status.Error(codes.PermissionDenied, "")
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errPermissionDenied
		}

		ips := md.Get("X-Real-Ip")
		if len(ips) == 0 {
			return nil, errPermissionDenied
		}

		ip := net.ParseIP(ips[0])
		if ip == nil {
			return nil, errPermissionDenied
		}

		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		if !ipNet.Contains(ip) {
			return nil, errPermissionDenied
		}

		return handler(ctx, req)
	}
}
