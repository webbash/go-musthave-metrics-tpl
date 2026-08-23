package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/retry"
)

var errServerFailure = errors.New("audit server failure")
var errConnectionFailure = errors.New("audit connection failure")

// HTTPObserver sends audit events as JSON to an HTTP endpoint.
type HTTPObserver struct {
	url            string
	client         *http.Client
	retryIntervals []time.Duration
}

// NewHTTPObserver creates an observer that sends audit events to url.
func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		retryIntervals: []time.Duration{
			time.Second,
			5 * time.Second,
			10 * time.Second,
		},
	}
}

// Observe sends an audit event to the configured HTTP endpoint, retrying
// network errors and 5xx responses.
func (o *HTTPObserver) Observe(e Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	return retry.Do(context.Background(), func() error {
		req, err := http.NewRequest(http.MethodPost, o.url, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create audit request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := o.client.Do(req)
		if err != nil {
			// Connection errors are retried.
			if _, ok := errors.AsType[*url.Error](err); ok {
				return fmt.Errorf("%w: %w", errConnectionFailure, err)
			}

			return fmt.Errorf("send audit request: %w", err)
		}
		defer resp.Body.Close()

		switch {
		case resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode < 600:
			return fmt.Errorf("%w: %s", errServerFailure, resp.Status)
		case resp.StatusCode != http.StatusOK:
			return fmt.Errorf(
				"audit server returned status %s",
				resp.Status,
			)
		default:
			return nil
		}
	}, func(err error) bool {
		if errors.Is(err, errServerFailure) || errors.Is(err, errConnectionFailure) {
			return true
		}

		return false
	}, o.retryIntervals)
}
