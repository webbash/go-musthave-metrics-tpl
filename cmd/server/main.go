package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/get_value_list"
	update_handler "github.com/webbash/go-musthave-metrics-tpl.git/internal/handler/update"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
)

func main() {
	r := chi.NewRouter()
	storage := repository.NewMemStorage()
	updateHandlerConcrete := update_handler.NewHandler(storage)
	getValueHandlerConcrete := get_value.NewHandler(storage)
	getValueListHandlerConcrete := get_value_list.NewHandler(storage)

	r.Get("/", getValueListHandlerConcrete.ServeHTTP)
	r.Get("/value/{metricType}/{metricName}", getValueHandlerConcrete.ServeHTTP)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", updateHandlerConcrete.ServeHTTP)

	log.Fatal(http.ListenAndServe(":8080", r))
}
