package get_value

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func TestHandler_ServeHTTP(t *testing.T) {
	storage := repository.NewMemStorage()
	ctx := context.Background()
	storage.UpdateGauge(ctx, "test_gauge", 10.6)
	storage.IncrementCounter(ctx, "test_counter", 10)

	metricsService := service.NewMetricsService(storage)
	handler := NewHandler(metricsService)
	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", handler.ServeHTTP)

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Get gauge",
			url:  "/value/gauge/test_gauge",
			want: want{
				code: http.StatusOK,
				body: "10.6",
			},
		},
		{
			name: "Get counter",
			url:  "/value/counter/test_counter",
			want: want{
				code: http.StatusOK,
				body: "10",
			},
		},
		{
			name: "Get unknown gauge",
			url:  "/value/gauge/unknown",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get unknown counter",
			url:  "/value/counter/unknown",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Unknown metric type",
			url:  "/value/unknown/test",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tt.want.code, recorder.Code)

			if tt.want.body != "" {
				body, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}
