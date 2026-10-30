package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/agent"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

func TestRouterTrustedSubnet(t *testing.T) {
	for _, tc := range []struct {
		name   string
		subnet string
		path   string
		ips    []string
		body   string
		want   int
	}{
		{name: "trusted IPv4", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"192.168.1.42"}, body: "[]", want: 200},
		{name: "trailing slash", subnet: "192.168.1.0/24", path: "/updates/", ips: []string{"192.168.1.42"}, body: "[]", want: 200},
		{name: "outside subnet", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"192.168.2.42"}, body: "[]", want: 403},
		{name: "missing header", subnet: "192.168.1.0/24", path: "/updates", body: "[]", want: 403},
		{name: "invalid IP", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"invalid"}, body: "[]", want: 403},
		{name: "IP with port", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"192.168.1.42:1234"}, body: "[]", want: 403},
		{name: "comma separated IPs", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"192.168.1.42,192.168.1.43"}, body: "[]", want: 403},
		{name: "multiple headers", subnet: "192.168.1.0/24", path: "/updates", ips: []string{"192.168.1.42", "192.168.1.43"}, body: "[]", want: 403},
		{name: "no RemoteAddr fallback", subnet: "192.0.2.0/24", path: "/updates", body: "[]", want: 403},
		{name: "trusted IPv6", subnet: "2001:db8::/32", path: "/updates", ips: []string{"2001:db8::42"}, body: "[]", want: 200},
		{name: "outside IPv6 subnet", subnet: "2001:db8::/32", path: "/updates", ips: []string{"2001:db9::42"}, body: "[]", want: 403},
		{name: "empty subnet without header", path: "/updates", body: "[]", want: 200},
		{name: "empty subnet with invalid IP", path: "/updates", ips: []string{"invalid"}, body: "[]", want: 200},
		{name: "single JSON update stays open", subnet: "192.168.1.0/24", path: "/update", body: `{"id":"test","type":"gauge","value":1}`, want: 200},
		{name: "single URL update stays open", subnet: "192.168.1.0/24", path: "/update/gauge/test/1", want: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop().Sugar()
			repo := repository.NewMemStorage()
			subject := audit.NewSubject(t.Context(), logger)
			t.Cleanup(subject.Close)
			router := NewRouter(config.Config{TrustedSubnet: tc.subnet}, logger, service.NewMetricsService(repo), repo, nil, subject, nil).Init()
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			for _, ip := range tc.ips {
				req.Header.Add("X-Real-IP", ip)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			require.Equal(t, tc.want, recorder.Code, recorder.Body.String())
		})
	}
}

func TestRouterTrustedSubnetPreservesMetricsOnRejection(t *testing.T) {
	for _, path := range []string{"/updates", "/updates/"} {
		t.Run(path, func(t *testing.T) {
			logger := zap.NewNop().Sugar()
			repo := repository.NewMemStorage()
			require.NoError(t, repo.UpdateGauge(t.Context(), "temperature", 20))
			require.NoError(t, repo.IncrementCounter(t.Context(), "PollCount", 5))
			subject := audit.NewSubject(t.Context(), logger)
			t.Cleanup(subject.Close)
			router := NewRouter(config.Config{TrustedSubnet: "10.0.0.0/8"}, logger, service.NewMetricsService(repo), repo, nil, subject, nil).Init()
			body := `[{"id":"temperature","type":"gauge","value":99},{"id":"PollCount","type":"counter","delta":100}]`
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
			req.Header.Set("X-Real-IP", "192.168.1.42")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusForbidden, recorder.Code)
			gauge, err := repo.GetGauge(t.Context(), "temperature")
			require.NoError(t, err)
			require.Equal(t, float64(20), gauge)
			counter, err := repo.GetCounter(t.Context(), "PollCount")
			require.NoError(t, err)
			require.Equal(t, int64(5), counter)
		})
	}
}

func TestRouterTrustedSubnetAllowsMetricReads(t *testing.T) {
	logger := zap.NewNop().Sugar()
	for _, tc := range []struct {
		method, path, body string
	}{
		{http.MethodGet, "/", ""},
		{http.MethodGet, "/value/gauge/temperature", ""},
		{http.MethodPost, "/value", `{"id":"temperature","type":"gauge"}`},
		{http.MethodPost, "/value/", `{"id":"temperature","type":"gauge"}`},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			repo := repository.NewMemStorage()
			require.NoError(t, repo.UpdateGauge(t.Context(), "temperature", 20))
			subject := audit.NewSubject(t.Context(), logger)
			t.Cleanup(subject.Close)
			router := NewRouter(config.Config{TrustedSubnet: "10.0.0.0/8"}, logger, service.NewMetricsService(repo), repo, nil, subject, nil).Init()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		})
	}
}

func TestRouterAcceptsAgentBatchWithTrustedSubnet(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	logger := zap.NewNop().Sugar()
	repo := repository.NewMemStorage()
	subject := audit.NewSubject(t.Context(), logger)
	t.Cleanup(subject.Close)
	cfg := config.Config{TrustedSubnet: "127.0.0.0/8", HashSecret: "secret"}
	router := NewRouter(cfg, logger, service.NewMetricsService(repo), repo, nil, subject, crypto.NewDecryptor(privateKey)).Init()
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	client := server.Client()
	t.Cleanup(client.CloseIdleConnections)
	sender := agent.NewSender(client, server.URL, crypto.NewSHA256Signer(cfg.HashSecret), crypto.NewEncryptor(&privateKey.PublicKey))
	value, delta := 23.5, int64(7)
	metrics := []models.Metrics{
		{ID: "temperature", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	}
	require.NoError(t, sender.Send(t.Context(), metrics))
	gauge, err := repo.GetGauge(t.Context(), "temperature")
	require.NoError(t, err)
	require.Equal(t, value, gauge)
	counter, err := repo.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)
	require.Equal(t, delta, counter)
}
