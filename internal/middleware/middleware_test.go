package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
)

func gzipData(t *testing.T, data []byte) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return &buf
}

func TestGzipMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write(body)
	})
	handler := GzipMiddleware()(next)

	t.Run("compresses json response", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"ok":true}`))
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
		reader, err := gzip.NewReader(recorder.Body)
		require.NoError(t, err)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		assert.Equal(t, `{"ok":true}`, string(body))
	})

	t.Run("passes plain response through", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("plain"))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Empty(t, recorder.Header().Get("Content-Encoding"))
		assert.Equal(t, "plain", recorder.Body.String())
	})

	t.Run("decompresses request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", gzipData(t, []byte("compressed request")))
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		reader, err := gzip.NewReader(recorder.Body)
		require.NoError(t, err)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		assert.Equal(t, "compressed request", string(body))
	})

	t.Run("rejects invalid compressed request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid"))
		request.Header.Set("Content-Encoding", "gzip")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}

func TestHashCheckMiddleware(t *testing.T) {
	signer := crypto.NewSHA256Signer("secret")
	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write(body)
	})
	handler := HashCheckMiddleware(signer, zap.NewNop().Sugar())(next)

	t.Run("passes request without hash", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("plain"))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, "plain", recorder.Body.String())
	})

	t.Run("verifies and signs request", func(t *testing.T) {
		body := []byte("signed payload")
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		request.Header.Set("HashSHA256", signer.Sign(body))
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Equal(t, string(body), recorder.Body.String())
		assert.True(t, signer.Verify(body, recorder.Header().Get("HashSHA256")))
	})

	t.Run("rejects invalid hash", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("payload"))
		request.Header.Set("HashSHA256", "invalid")
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte("ok"))
	})
	handler := LoggingMiddleware(zap.NewNop().Sugar())(next)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, "ok", recorder.Body.String())
}
