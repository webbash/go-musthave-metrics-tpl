package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTrustedSubnetInterceptor(t *testing.T) {
	for _, tc := range []struct {
		name, subnet string
		addresses    []string
		code         codes.Code
	}{
		{"disabled", "", nil, codes.OK},
		{"IPv4", "127.0.0.0/8", []string{"127.0.0.1"}, codes.OK},
		{"IPv6", "::1/128", []string{"::1"}, codes.OK},
		{"outside", "10.0.0.0/8", []string{"127.0.0.1"}, codes.PermissionDenied},
		{"missing", "127.0.0.0/8", nil, codes.PermissionDenied},
		{"empty", "127.0.0.0/8", []string{""}, codes.PermissionDenied},
		{"malformed", "127.0.0.0/8", []string{"127.0.0.1:8081"}, codes.PermissionDenied},
		{"multiple", "127.0.0.0/8", []string{"127.0.0.1", "127.0.0.1"}, codes.PermissionDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interceptor, err := TrustedSubnetInterceptor(tc.subnet)
			require.NoError(t, err)
			ctx := metadata.NewIncomingContext(t.Context(), metadata.MD{"x-real-ip": tc.addresses})
			called := false
			_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) { called = true; return nil, nil })
			require.Equal(t, tc.code, status.Code(err))
			require.Equal(t, tc.code == codes.OK, called)
		})
	}
	_, err := TrustedSubnetInterceptor("invalid-CIDR")
	require.Error(t, err)
}
