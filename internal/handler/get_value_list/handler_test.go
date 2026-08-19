package get_value_list

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
)

func TestHandler_ServeHTTP(t *testing.T) {
	storage := repository.NewMemStorage()
	ctx := context.Background()
	err := storage.UpdateGauge(ctx, "test_gauge", 10.6)
	require.NoError(t, err)
	err = storage.IncrementCounter(ctx, "test_counter", 10)
	require.NoError(t, err)

	handler := NewHandler(storage)
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

// Benchmark test
func BenchmarkHandler_buildHTML(b *testing.B) {
	gauges, counters := benchmarkMetrics()

	b.Run("concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buildHTMLConcat(gauges, counters)
		}
	})
	b.Run("builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buildHTMLBuilder(gauges, counters)
		}
	})
}

func benchmarkMetrics() (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64, 1000)
	counters := make(map[string]int64, 1000)

	for i := 0; i < 1000; i++ {
		name := "metric_" + strconv.Itoa(i)
		gauges[name] = float64(i)
		counters[name] = int64(i)
	}

	return gauges, counters
}
