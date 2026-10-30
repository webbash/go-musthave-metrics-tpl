package agent

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc/credentials"
)

// AuthInfo belongs to one transport, so reconnecting never replaces another
// in-flight RPC's local address. No mutable IP is shared between workers.
type localIPAuth struct {
	credentials.TLSInfo
	ip string
}

type localIPTransport struct {
	credentials.TransportCredentials
}

func newLocalIPTransport(certFile string) (credentials.TransportCredentials, error) {
	if certFile == "" {
		return nil, errors.New("gRPC TLS certificate is required")
	}
	tlsCredentials, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		return nil, fmt.Errorf("load gRPC TLS certificate: %w", err)
	}
	return localIPTransport{TransportCredentials: tlsCredentials}, nil
}

func (c localIPTransport) ClientHandshake(ctx context.Context, authority string, raw net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, auth, err := c.TransportCredentials.ClientHandshake(ctx, authority, raw)
	if err != nil {
		return nil, nil, err
	}
	tlsInfo, ok := auth.(credentials.TLSInfo)
	if !ok {
		_ = conn.Close()
		return nil, nil, errors.New("gRPC connection has no TLS information")
	}
	address, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok || address.IP == nil {
		_ = conn.Close()
		return nil, nil, errors.New("gRPC connection has no local TCP IP")
	}
	return conn, localIPAuth{
		TLSInfo: tlsInfo,
		ip:      address.IP.String(),
	}, nil
}

func (c localIPTransport) Clone() credentials.TransportCredentials {
	return localIPTransport{TransportCredentials: c.TransportCredentials.Clone()}
}

type localIPMetadata struct{}

func (localIPMetadata) RequireTransportSecurity() bool { return true }

func (localIPMetadata) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	request, ok := credentials.RequestInfoFromContext(ctx)
	if !ok {
		return nil, errors.New("gRPC request information missing")
	}
	address, ok := request.AuthInfo.(localIPAuth)
	if !ok || address.ip == "" {
		return nil, errors.New("gRPC local IP missing")
	}
	return map[string]string{"x-real-ip": address.ip}, nil
}
