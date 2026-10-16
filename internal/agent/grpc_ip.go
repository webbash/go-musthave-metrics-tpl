package agent

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthInfo belongs to one transport, so reconnecting never replaces another
// in-flight RPC's local address. No mutable IP is shared between workers.
type localIPAuth struct {
	credentials.CommonAuthInfo
	ip string
}

func (localIPAuth) AuthType() string { return "insecure" }

type localIPTransport struct {
	credentials.TransportCredentials
}

func newLocalIPTransport() credentials.TransportCredentials {
	return localIPTransport{TransportCredentials: insecure.NewCredentials()}
}

func (c localIPTransport) ClientHandshake(ctx context.Context, authority string, raw net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, _, err := c.TransportCredentials.ClientHandshake(ctx, authority, raw)
	if err != nil {
		return nil, nil, err
	}
	address, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok || address.IP == nil {
		_ = conn.Close()
		return nil, nil, errors.New("gRPC connection has no local TCP IP")
	}
	return conn, localIPAuth{
		CommonAuthInfo: credentials.CommonAuthInfo{SecurityLevel: credentials.NoSecurity},
		ip:             address.IP.String(),
	}, nil
}

func (c localIPTransport) Clone() credentials.TransportCredentials {
	return localIPTransport{TransportCredentials: c.TransportCredentials.Clone()}
}

type localIPMetadata struct{}

func (localIPMetadata) RequireTransportSecurity() bool { return false }

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
