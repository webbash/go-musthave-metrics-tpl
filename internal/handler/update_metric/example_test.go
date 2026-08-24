package update_metric

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func ExampleHandler() {
	logger := zap.NewNop().Sugar()
	storage := repository.NewMemStorage()
	handler := NewHandler(service.NewMetricsService(storage), audit.NewSubject(context.Background(), logger), logger)

	router := chi.NewRouter()
	router.Post("/update", handler.ServeHTTP)

	requestBody := `{"id":"temperature","type":"gauge","value":23.5}`
	request := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(requestBody))
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
