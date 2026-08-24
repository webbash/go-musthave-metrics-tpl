package update_batch

import (
	"context"
	"fmt"
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
	handler := NewHandler(service.NewMetricsService(storage), logger, audit.NewSubject(context.Background(), logger))

	router := chi.NewRouter()
	router.Post("/updates", handler.ServeHTTP)

	requestBody := `[{"id":"temperature","type":"gauge","value":23.5},{"id":"requests","type":"counter","delta":1}]`
	request := httptest.NewRequest(http.MethodPost, "/updates", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	fmt.Println(recorder.Code)
	// Output:
	// 200
}
