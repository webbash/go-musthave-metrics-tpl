package middleware

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnetInterceptor checks the declared IP before a handler can write.
func TrustedSubnetInterceptor(trustedSubnet string) (grpc.UnaryServerInterceptor, error) {
	var subnet *net.IPNet
	if trustedSubnet != "" {
		var err error
		_, subnet, err = net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("parse trusted subnet: %w", err)
		}
	}
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if subnet != nil {
			md, _ := metadata.FromIncomingContext(ctx)
			addresses := md.Get("x-real-ip")
			if len(addresses) != 1 || !subnet.Contains(net.ParseIP(addresses[0])) {
				return nil, status.Error(codes.PermissionDenied, "agent IP is outside the trusted subnet")
			}
		}
		return handler(ctx, req)
	}, nil
}
