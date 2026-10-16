package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	updatebatch "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_batch"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
)

type metricsServer interface {
	Serve() error
	Shutdown(context.Context) error
}

// newMetricsServer allocates only the configured metrics listener.
func newMetricsServer(cfg config.Config, logger *zap.SugaredLogger, metrics *service.MetricsService, repo service.MetricsRepository, database *sql.DB, subject *audit.Subject) (metricsServer, error) {
	if err := config.ValidateTransport(cfg.Transport); err != nil {
		return nil, err
	}
	if cfg.Transport == config.GRPC {
		return newGRPCMetricsServer(cfg.GRPCAddress, cfg.TrustedSubnet, updatebatch.NewGRPCHandler(metrics, logger))
	}
	var decryptor *crypto.Decryptor
	if cfg.CryptoKey != "" {
		privateKey, err := crypto.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			return nil, fmt.Errorf("load private crypto key: %w", err)
		}
		decryptor = crypto.NewDecryptor(privateKey)
	}
	router := internal.NewRouter(cfg, logger, metrics, repo, database, subject, decryptor).Init()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("listen HTTP: %w", err)
	}
	return &httpMetricsServer{
		listener: listener,
		server: &http.Server{Addr: cfg.Address, Handler: router,
			ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 120 * time.Second},
	}, nil
}

type httpMetricsServer struct {
	server   *http.Server
	listener net.Listener
}

func (s *httpMetricsServer) Serve() error {
	err := s.server.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *httpMetricsServer) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Join(err, s.server.Close())
	}
	return nil
}
