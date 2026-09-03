package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/retry"
)

// Sender sends compressed metric batches to the server.
type Sender struct {
	httpClient *http.Client
	baseURL    string
	signer     *crypto.SHA256Signer
}

// NewSender creates a metrics sender using httpClient and baseURL.
func NewSender(httpClient *http.Client, baseURL string, signer *crypto.SHA256Signer) *Sender {
	return &Sender{
		httpClient: httpClient,
		baseURL:    baseURL,
		signer:     signer,
	}
}

// Send compresses and sends metrics to the batch update endpoint, retrying
// temporary network failures.
func (a *Sender) Send(ctx context.Context, metrics []models.Metrics) error {
	err := retry.Do(ctx, func() error {
		return a.sendMetrics(ctx, metrics)
	}, func(err error) bool {
		if err == nil {
			return false
		}

		var netErr net.Error
		return errors.As(err, &netErr)
	}, []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	})

	if err != nil {
		return fmt.Errorf("error sending metrics: %w", err)
	}

	return nil
}

func (a *Sender) sendMetrics(ctx context.Context, metric []models.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err = gz.Write(body); err != nil {
		return fmt.Errorf("gzip write: %w", err)
	}
	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip closing: %w", err)
	}

	updateUrl, err := url.JoinPath(a.baseURL, "/updates")
	if err != nil {
		return fmt.Errorf("create url: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, updateUrl, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	if a.signer != nil {
		req.Header.Set("HashSHA256", a.signer.Sign(body))
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send update: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send update: %d", resp.StatusCode)
	}

	return nil
}
