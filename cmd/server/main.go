package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config/db"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()
	sugar := logger.NewLogger()

	var database *sql.DB
	if cfg.DatabaseDSN != "" {
		var err error

		database, err = db.NewPGConnector(cfg.DatabaseDSN).Connect()
		if err != nil {
			sugar.Errorw("failed to connect to database", "err", err)
			os.Exit(1)
		}

		defer database.Close()
	}

	fileStorage := storage.NewFileStorage(cfg.FileStoragePath)
	repo := buildRepository(cfg, sugar, fileStorage, database)
	metricsService := service.NewMetricsService(repo)

	r := internal.NewRouter(cfg, sugar, metricsService, repo).Init()

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Grace period для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
					metrics, err := repo.GetAllMetrics(context.Background())
					if err != nil {
						sugar.Errorw("failed to get all metrics", "err", err)
						continue
					}
					err = fileStorage.Save(metrics)
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
	database *sql.DB,
) service.MetricsRepository {
	if database != nil {
		return repository.NewPostgresRepository(database)
	}

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
