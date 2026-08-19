package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func ExampleHandler() {
	logger := zap.NewNop().Sugar()
	storage := repository.NewMemStorage()
	handler := NewHandler(service.NewMetricsService(storage), audit.NewSubject(logger), logger)

	router := chi.NewRouter()
	router.Post("/update/{metricType}/{metricName}/{metricValue}", handler.ServeHTTP)

	request := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/23.5", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	fmt.Println(recorder.Code)
	// Output:
	// 200
}
