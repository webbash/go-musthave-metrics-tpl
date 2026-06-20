package get_value_list

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	file "github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
)

func TestHandler_ServeHTTP(t *testing.T) {
	storage := repository.NewMemStorage()
	ctx := context.Background()
	storage.UpdateGauge(ctx, "test_gauge", 10.6)
	storage.IncrementCounter(ctx, "test_counter", 10)
	fileStorage := &file.MockFileStorage{}

	metricsService := service.NewMetricsService(storage, fileStorage, 0)
	handler := NewHandler(metricsService)
	r := chi.NewRouter()
	r.Get("/", handler.ServeHTTP)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/html", recorder.Header().Get("Content-Type"))

	body, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)

	bodyStr := string(body)
	assert.True(t, strings.Contains(bodyStr, "test_gauge: 10.6"))
	assert.True(t, strings.Contains(bodyStr, "test_counter: 10"))
}
