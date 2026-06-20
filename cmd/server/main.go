package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/webbash/go-musthave-metrics-tpl.git/cmd/server/middleware"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_metric"
	update_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	update_metric_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update_metric"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
	"go.uber.org/zap"
)

func main() {
	var addr string
	var restore bool
	var storeInterval int
	var fileStoragePath string

	flag.StringVar(&addr, "a", "localhost:8080", "server endpoint")
	flag.IntVar(&storeInterval, "i", 300, "metric collection to file interval")
	flag.StringVar(&fileStoragePath, "f", "./tmp/temporary.json", "file memRepository path")
	flag.BoolVar(&restore, "r", false, "restore metrics from file")

	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		addr = envAddress
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if seconds, err := strconv.Atoi(envStoreInterval); err == nil {
			storeInterval = seconds
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if r, err := strconv.ParseBool(envRestore); err == nil {
			restore = r
		}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	r := chi.NewRouter()
	fileStorage := storage.NewFileStorage(fileStoragePath)

	var memRepository *repository.MemStorage
	if restore {
		metrics, err := fileStorage.Load()
		if err != nil {
			sugar.Errorw("failed to restore metrics from file: %w", err)
			return
		}
		memRepository = repository.NewMemStorageFromMetrics(metrics)
	} else {
		memRepository = repository.NewMemStorage()
	}

	metricsService := service.NewMetricsService(memRepository, fileStorage, storeInterval)
	updateH := update_handler.NewHandler(metricsService)
	updateMetricH := update_metric_handler.NewHandler(metricsService)
	getValueMetricH := get_value_metric.NewHandler(metricsService)
	getValueH := get_value.NewHandler(metricsService)
	getValueListH := get_value_list.NewHandler(metricsService)

	r.Use(middleware.LoggingMiddleware(sugar))
	r.Use(middleware.GzipMiddleware())

	r.Get("/", getValueListH.ServeHTTP)
	r.Post("/value", getValueMetricH.ServeHTTP)
	r.Post("/value/", getValueMetricH.ServeHTTP)
	r.Post("/update", updateMetricH.ServeHTTP)
	r.Post("/update/", updateMetricH.ServeHTTP)
	r.Get("/value/{metricType}/{metricName}", getValueH.ServeHTTP)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateH.ServeHTTP)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful start
	go func() {
		sugar.Infow("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Infow("server closed", "err", err)
			os.Exit(1)
		}
	}()

	if storeInterval != 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					metrics := memRepository.GetAllMetrics()
					err := fileStorage.Save(metrics)
					if err != nil {
						sugar.Errorw(fmt.Sprintf("failed to save metrics to file: %s", err))
						return
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
