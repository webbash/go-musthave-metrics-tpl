package get_value_list

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
)

func TestHandler_ServeHTTP(t *testing.T) {
	storage := repository.NewMemStorage()
	storage.UpdateGauge("test_gauge", 10.6)
	storage.IncrementCounter("test_counter", 10)

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
