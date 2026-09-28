package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config/db"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"

	_ "net/http/pprof"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	cfg := config.NewConfig()
	sugar := logger.NewLogger()

	var database *sql.DB
	if cfg.DatabaseDSN != "" {
		var err error
		database, err = db.NewPGConnector(cfg.DatabaseDSN).Connect()
		if err != nil {
			sugar.Fatalw("failed to connect to database", "err", err)
		}

		err = goose.SetDialect("postgres")
		if err != nil {
			sugar.Fatalw("failed to set sql dialect", "err", err)
		}

		err = goose.Up(database, "migrations")
		if err != nil {
			sugar.Fatalw("failed to run migrations", "err", err)
		}

		defer database.Close()
	}

	auditCtx, cancelAudit := context.WithCancel(context.Background())
	defer cancelAudit()
	obsSubject := audit.NewSubject(auditCtx, sugar)
	if cfg.AuditFile != "" {
		file, err := os.OpenFile(
			cfg.AuditFile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			sugar.Errorw("failed to open audit file, skipping audit for file", "err", err)
		} else {
			defer file.Close()
			fileObserver := audit.NewFileObserver(file)
			obsSubject.AddObserver(fileObserver)
		}
	}

	if cfg.AuditURL != "" {
		httpObserver := audit.NewHTTPObserver(cfg.AuditURL)
		obsSubject.AddObserver(httpObserver)
	}

	fileStorage := storage.NewFileStorage(cfg.FileStoragePath)
	repo := buildRepository(cfg, sugar, fileStorage, database)
	metricsService := service.NewMetricsService(repo)

	var decryptor *crypto.Decryptor
	if cfg.CryptoKey != "" {
		privateKey, err := crypto.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			sugar.Fatalw("failed to load private crypto key", "err", err)
		}
		decryptor = crypto.NewDecryptor(privateKey)
	}

	r := internal.NewRouter(cfg, sugar, metricsService, repo, database, obsSubject, decryptor).Init()

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
		}
	}()

	go func() {
		sugar.Infow("starting pprof server", "addr", "localhost:6060")

		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			sugar.Errorw("pprof server stopped", "err", err)
		}
	}()

	stopSaving := make(chan struct{})
	var savingWG sync.WaitGroup
	if cfg.StoreInterval != 0 {
		savingWG.Add(1)
		go func() {
			defer savingWG.Done()
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-stopSaving:
					return
				case <-ticker.C:
				}
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
		}()
	}

	// Graceful shutdown
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	<-signalCtx.Done()
	sugar.Infow("shutting down server")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()
	shutdownErr := srv.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		sugar.Errorw("failed to shut down server", "err", shutdownErr)
		if err := srv.Close(); err != nil {
			sugar.Errorw("failed to close server connections", "err", err)
		}
	}

	// Дожидаемся текущей записи, чтобы она не перезаписала финальный снимок.
	close(stopSaving)
	savingWG.Wait()

	var saveErr error
	if database == nil && cfg.FileStoragePath != "" {
		metrics, err := repo.GetAllMetrics(context.Background())
		if err != nil {
			saveErr = err
		} else {
			saveErr = fileStorage.Save(metrics)
		}
		if saveErr != nil {
			sugar.Errorw("failed to save final metrics", "err", saveErr)
		}
	}

	obsSubject.Close()
	cancelAudit()
	if shutdownErr != nil || saveErr != nil {
		return
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
