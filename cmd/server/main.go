package main

import (
	"log"
	"net/http"

	updateHandler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service/update"
)

func main() {
	storage := repository.NewMemStorage()
	service := update.NewService(storage)
	updateHandlerConcrete := updateHandler.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", updateHandlerConcrete.ServeHTTP)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
