package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
)

func TestServerStartsOnlySelectedTransport(t *testing.T) {
	for _, transport := range []string{config.HTTP, config.GRPC} {
		t.Run(transport, func(t *testing.T) {
			occupied, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			t.Cleanup(func() { _ = occupied.Close() })
			cfg := config.Config{Transport: transport, Address: "127.0.0.1:0", GRPCAddress: "127.0.0.1:0"}
			if transport == config.HTTP {
				cfg.GRPCAddress = occupied.Addr().String()
			} else {
				cfg.GRPCTLSCert, cfg.GRPCTLSKey = testTLSFiles(t)
				cfg.Address = occupied.Addr().String()
				cfg.CryptoKey = "absent-http-only-key.pem"
			}
			logger := zap.NewNop().Sugar()
			subject := audit.NewSubject(t.Context(), logger)
			t.Cleanup(subject.Close)
			s, err := newMetricsServer(cfg, logger, service.NewMetricsService(nil), nil, nil, subject)
			require.NoError(t, err, "unused transport must not bind or load crypto keys")
			done := make(chan error, 1)
			go func() { done <- s.Serve() }()
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			require.NoError(t, s.Shutdown(ctx))
			require.NoError(t, <-done)
		})
	}
}

func TestServerRejectsInvalidConfigurationAndOccupiedPort(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer occupied.Close()
	certFile, keyFile := testTLSFiles(t)
	for _, cfg := range []config.Config{
		{Transport: "unknown"},
		{Transport: config.GRPC, GRPCAddress: "127.0.0.1:0", TrustedSubnet: "bad"},
		{Transport: config.GRPC, GRPCAddress: ""},
		{Transport: config.GRPC, GRPCAddress: occupied.Addr().String()},
		{Transport: config.HTTP, Address: occupied.Addr().String()},
	} {
		cfg.GRPCTLSCert, cfg.GRPCTLSKey = certFile, keyFile
		_, err := newMetricsServer(cfg, zap.NewNop().Sugar(), service.NewMetricsService(nil), nil, nil, nil)
		require.Error(t, err)
	}
}
