package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"sync"
	"testing"
	"time"

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

func TestSendMetricsRequestContext(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		name := "timeout"
		if canceled {
			name = "parent cancellation"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				deadline, ok := req.Context().Deadline()
				require.True(t, ok, "each request must have a deadline")
				require.InDelta(t, 10*time.Second, time.Until(deadline), float64(time.Second))
				if canceled {
					cancel()
					require.ErrorIs(t, req.Context().Err(), context.Canceled)
					return nil, req.Context().Err()
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(nil)), Header: make(http.Header)}, nil
			})}
			sender := NewSender(client, "http://metrics.test", nil, nil)
			err := sender.sendMetrics(ctx, nil)
			if canceled {
				require.ErrorIs(t, err, context.Canceled)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSenderRealIP(t *testing.T) {
	for _, tc := range []struct {
		name, network, address, subnet string
	}{
		{"IPv4", "tcp4", "127.0.0.1:0", "127.0.0.0/8"},
		{"IPv6", "tcp6", "[::1]:0", "::1/128"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen(tc.network, tc.address)
			if err != nil && tc.network == "tcp6" {
				t.Skipf("IPv6 unavailable: %v", err)
			}
			require.NoError(t, err)
			addresses := make(chan string, 6)
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				peer, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil || r.Header.Get("X-Real-IP") != peer {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				addresses <- peer
				w.WriteHeader(http.StatusOK)
			})
			server := httptest.NewUnstartedServer(middleware.TrustedSubnetMiddleware(tc.subnet)(next))
			require.NoError(t, server.Listener.Close())
			server.Listener = listener
			server.Start()
			t.Cleanup(server.Close)
			client := server.Client()
			t.Cleanup(client.CloseIdleConnections)
			sender := NewSender(client, server.URL, nil, nil)
			// Exercise both a reused connection and concurrent worker requests.
			reused := make(chan bool, 2)
			ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) { reused <- info.Reused },
			})
			require.NoError(t, sender.Send(ctx, nil))
			require.NoError(t, sender.Send(ctx, nil))
			require.False(t, <-reused, "the first request opens a connection")
			require.True(t, <-reused, "the second request must reuse the connection")
			var workers sync.WaitGroup
			errors := make(chan error, 4)
			for range 4 {
				workers.Go(func() { errors <- sender.Send(t.Context(), nil) })
			}
			workers.Wait()
			close(errors)
			for err := range errors {
				require.NoError(t, err)
			}
			require.Len(t, addresses, 6)
		})
	}
}
