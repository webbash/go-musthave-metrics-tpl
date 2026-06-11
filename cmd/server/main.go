package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list"
	update_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
)

func main() {
	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "server endpoint")
	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		addr = envAddress
	}

	r := chi.NewRouter()
	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	updateHandlerConcrete := update_handler.NewHandler(metricsService)
	getValueHandlerConcrete := get_value.NewHandler(metricsService)
	getValueListHandlerConcrete := get_value_list.NewHandler(metricsService)

	r.Get("/", getValueListHandlerConcrete.ServeHTTP)
	r.Get("/value/{metricType}/{metricName}", getValueHandlerConcrete.ServeHTTP)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandlerConcrete.ServeHTTP)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := *logger.Sugar()

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful start
	go func() {
		sugar.Infow("starting server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sugar.Infow("server closed", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Infow("shutting down server")

	// Grace period для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorw("server forced to shutdown", zap.Error(err))
		os.Exit(1)
	}

	sugar.Infow("server stopped gracefully")
}
