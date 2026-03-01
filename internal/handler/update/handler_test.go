package update

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service/update"
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
	}

	storage := repository.NewMemStorage()
	handler := NewHandler(update.NewService(storage))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", handler.ServeHTTP)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			res := response.Result()

			assert.Equal(t, tt.args.want, response.Result().StatusCode)

			res.Body.Close()
		})
	}
}
