package audit

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type observerFunc func(Event) error

func (f observerFunc) Observe(event Event) error {
	return f(event)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestFileObserver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	event := Event{TS: 1, Metrics: []string{"temperature"}, IPAddress: "127.0.0.1"}

	require.NoError(t, NewFileObserver(path).Observe(event))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"temperature"`)

	err = NewFileObserver(filepath.Join(t.TempDir(), "missing", "audit.log")).Observe(event)
	assert.Error(t, err)
}

func TestHTTPObserver(t *testing.T) {
	event := Event{TS: 1, Metrics: []string{"temperature"}}
	observer := NewHTTPObserver("http://audit.test/events")
	observer.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "temperature")
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	require.NoError(t, observer.Observe(event))

	observer.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadGateway, Status: "502 Bad Gateway", Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	assert.Error(t, observer.Observe(event))

	observer.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})}
	assert.Error(t, observer.Observe(event))

	assert.Error(t, NewHTTPObserver("://invalid").Observe(event))
}

func TestSubjectNotify(t *testing.T) {
	subject := NewSubject(zap.NewNop().Sugar())
	called := false
	subject.AddObserver(observerFunc(func(Event) error {
		called = true
		return nil
	}))
	subject.AddObserver(observerFunc(func(Event) error {
		return errors.New("observer failed")
	}))

	require.NoError(t, subject.Notify(Event{}))
	assert.True(t, called)
}
