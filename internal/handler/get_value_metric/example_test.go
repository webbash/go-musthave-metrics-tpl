package get_value_metric

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func ExampleHandler() {
	storage := repository.NewMemStorage()
	_ = storage.UpdateGauge(context.Background(), "temperature", 23.5)
	handler := NewHandler(service.NewMetricsService(storage))

	router := chi.NewRouter()
	router.Post("/value", handler.ServeHTTP)

	requestBody := `{"id":"temperature","type":"gauge"}`
	request := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	body, _ := io.ReadAll(recorder.Body)
	fmt.Println(recorder.Code)
	fmt.Println(string(body))
	// Output:
	// 200
	// {"id":"temperature","type":"gauge","value":23.5}
}
