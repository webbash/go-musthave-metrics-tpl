package agent

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/middleware"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSenderEncryptsCompressedBody(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	decryptor := crypto.NewDecryptor(privateKey)
	encryptor := crypto.NewEncryptor(publicKey)
	signer := crypto.NewSHA256Signer("secret")

	want := []models.Metrics{{
		ID:    "temperature",
		MType: models.Gauge,
		Value: func() *float64 { value := 23.5; return &value }(),
	}}
	var got []models.Metrics

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &got))
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
	})
	handler := middleware.DecryptMiddleware(decryptor)(
		middleware.GzipMiddleware()(
			middleware.HashCheckMiddleware(signer, zap.NewNop().Sugar())(next),
		),
	)
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return &http.Response{
			StatusCode: recorder.Code,
			Header:     recorder.Header(),
			Body:       io.NopCloser(bytes.NewReader(recorder.Body.Bytes())),
			Request:    request,
		}, nil
	})}
	sender := NewSender(client, "http://metrics.test", signer, encryptor)
	require.NoError(t, sender.Send(t.Context(), want))
	assert.Equal(t, want, got)
}
