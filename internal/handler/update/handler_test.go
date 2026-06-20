package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	file "github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
)

func TestHandler_ServeHTTP(t *testing.T) {
	type args struct {
		url    string
		method string
		want   int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Gauge with float",
			args: args{
				url:    "/update/gauge/test/10.6",
				method: http.MethodPost,
				want:   http.StatusOK,
			},
		},
		{
			name: "Gauge with string",
			args: args{
				url:    "/update/gauge/test/abc",
				method: http.MethodPost,
				want:   http.StatusBadRequest,
			},
		},
		{
			name: "Counter with int",
			args: args{
				url:    "/update/counter/test/10",
				method: http.MethodPost,
				want:   http.StatusOK,
			},
		},
		{
			name: "Counter with float",
			args: args{
				url:    "/update/counter/test/10.6",
				method: http.MethodPost,
				want:   http.StatusBadRequest,
			},
		},
		{
			name: "Unknown metric type",
			args: args{
				url:    "/update/unknown/test/10",
				method: http.MethodPost,
				want:   http.StatusNotImplemented,
			},
		},
	}

	storage := repository.NewMemStorage()
	fileStorage := &file.MockFileStorage{}

	metricsService := service.NewMetricsService(storage, fileStorage, 0)
	handler := NewHandler(metricsService)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.ServeHTTP)
			req := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			assert.Equal(t, tt.args.want, recorder.Code)

			if recorder.Code != tt.args.want {
				fmt.Printf("Error body: %s\n", recorder.Body.String())
			}
		})
	}
}
