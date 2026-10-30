package main

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/middleware"
	pb "github.com/webbash/go-musthave-metrics-tpl.git/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// grpcMetricsServer owns the listener and waits for handlers during shutdown.
type grpcMetricsServer struct {
	server   *grpc.Server
	listener net.Listener
}

func newGRPCMetricsServer(address, trustedSubnet, certFile, keyFile string, handler pb.MetricsServer) (*grpcMetricsServer, error) {
	interceptor, err := middleware.TrustedSubnetInterceptor(trustedSubnet)
	if err != nil {
		return nil, err
	}
	if address == "" {
		return nil, fmt.Errorf("gRPC address must not be empty")
	}
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("gRPC TLS certificate and private key are required")
	}
	tlsCredentials, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load gRPC TLS certificate and private key: %w", err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen gRPC: %w", err)
	}
	server := grpc.NewServer(grpc.Creds(tlsCredentials), grpc.UnaryInterceptor(interceptor), grpc.WaitForHandlers(true))
	pb.RegisterMetricsServer(server, handler)
	return &grpcMetricsServer{server: server, listener: listener}, nil
}

func (s *grpcMetricsServer) Serve() error {
	err := s.server.Serve(s.listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

// Shutdown gives active RPCs the caller's grace period, then cancels them.
// WaitForHandlers prevents a cancelled database write from racing final cleanup.
func (s *grpcMetricsServer) Shutdown(ctx context.Context) error {
	// Also release the listener if shutdown wins the race with Serve starting.
	defer s.listener.Close()
	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		<-done
		return ctx.Err()
	}
}
