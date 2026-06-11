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
	updateH := update_handler.NewHandler(metricsService)
	getValueH := get_value.NewHandler(metricsService)
	getValueListH := get_value_list.NewHandler(metricsService)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	r.Use(withLogging(sugar))

	r.Get("/", getValueListH.ServeHTTP)
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

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.responseData.status = status
}

func withLogging(sugar *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			responseData := &responseData{
				status: http.StatusOK,
				size:   0,
			}

			lw := &loggingResponseWriter{ResponseWriter: w, responseData: responseData}
			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			sugar.Infow(
				"request completed",
				"uri", uri,
				"method", method,
				"duration", duration,
				"status", responseData.status,
				"size", responseData.size,
			)
		})
	}
}
