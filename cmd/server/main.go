package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	sugar := logger.NewLogger()

	fileStorage := storage.NewFileStorage(cfg.FileStoragePath)
	repo := buildRepository(cfg, sugar, fileStorage)
	metricsService := service.NewMetricsService(repo)

	r := internal.NewRouter(cfg, sugar, metricsService).Init()

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful start
	go func() {
		sugar.Infow("starting server", "addr", cfg.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Errorw("server closed", "err", err)
			os.Exit(1)
		}
	}()

	if cfg.StoreInterval != 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					metrics := repo.GetAllMetrics()
					err := fileStorage.Save(metrics)
					if err != nil {
						sugar.Errorw("failed to save metrics to file", "err", err)
						continue
					}
				}
			}
		}()
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Infow("shutting down server")

	// Grace period для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorw("server forced to shutdown", "err", err)
		os.Exit(1)
	}

	sugar.Infow("server stopped gracefully")
}

func buildRepository(
	cfg config.Config,
	sugar *zap.SugaredLogger,
	fileStorage *storage.FileStorage,
) service.MetricsRepository {
	var memStorage *repository.MemStorage
	if cfg.Restore {
		metrics, err := fileStorage.Load()
		if err != nil {
			sugar.Errorw("failed to load metrics from file", "err", err)
			memStorage = repository.NewMemStorage()
		} else {
			memStorage = repository.NewMemStorageFromMetrics(metrics)
		}
	} else {
		memStorage = repository.NewMemStorage()
	}

	var repo service.MetricsRepository
	if cfg.StoreInterval == 0 {
		repo = repository.NewFileRepository(memStorage, fileStorage)
	} else {
		repo = memStorage
	}

	return repo
}
