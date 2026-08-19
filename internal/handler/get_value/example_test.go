package get_value

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func ExampleHandler() {
	storage := repository.NewMemStorage()
	_ = storage.UpdateGauge(context.Background(), "temperature", 23.5)
	handler := NewHandler(service.NewMetricsService(storage))

	router := chi.NewRouter()
	router.Get("/value/{metricType}/{metricName}", handler.ServeHTTP)

	request := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var value string
	_, _ = fmt.Fscan(recorder.Body, &value)
	fmt.Println(recorder.Code)
	fmt.Println(value)
	// Output:
	// 200
	// 23.5
}
