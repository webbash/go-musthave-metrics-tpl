package main

import (
	"context"
	"errors"
	"flag"
	"log"
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
)

func main() {
	addr := flag.String("a", "localhost:8080", "server endpoint")
	flag.Parse()

	r := chi.NewRouter()
	storage := repository.NewMemStorage()
	updateHandlerConcrete := update_handler.NewHandler(storage)
	getValueHandlerConcrete := get_value.NewHandler(storage)
	getValueListHandlerConcrete := get_value_list.NewHandler(storage)

	r.Get("/", getValueListHandlerConcrete.ServeHTTP)
	r.Get("/value/{metricType}/{metricName}", getValueHandlerConcrete.ServeHTTP)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandlerConcrete.ServeHTTP)

	srv := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful start
	go func() {
		log.Printf("Server starting on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Grace period для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server stopped gracefully")
}
