package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type observerFunc func(Event) error

func (f observerFunc) Observe(_ context.Context, event Event) error {
	return f(event)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestFileObserver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	event := Event{TS: 1, Metrics: []string{"temperature"}, IPAddress: "127.0.0.1"}
	observer := NewFileObserver(file)

	require.NoError(t, observer.Observe(context.Background(), event))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"temperature"`)

	require.NoError(t, file.Close())
	assert.Error(t, observer.Observe(context.Background(), event))
}

func TestFileObserverConcurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = file.Close()
	})

	observer := NewFileObserver(file)
	const eventCount = 100

	var wg sync.WaitGroup
	errs := make(chan error, eventCount)
	for range eventCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- observer.Observe(context.Background(), Event{TS: 1, Metrics: []string{"temperature"}})
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, eventCount)
	for _, line := range lines {
		var event Event
		require.NoError(t, json.Unmarshal([]byte(line), &event))
		assert.Equal(t, []string{"temperature"}, event.Metrics)
	}
}

func TestHTTPObserverSuccess(t *testing.T) {
	event := Event{TS: 1, Metrics: []string{"temperature"}}
	observer := newTestHTTPObserver(roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "temperature")
		return response(http.StatusOK), nil
	}))

	require.NoError(t, observer.Observe(context.Background(), event))
}

func TestHTTPObserverRetriesServerFailure(t *testing.T) {
	observer := NewHTTPObserver("http://audit.test/events")
	observer.retryIntervals = []time.Duration{0, 0, 0}
	calls := 0
	observer.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return response(http.StatusBadGateway), nil
	})}

	err := observer.Observe(context.Background(), Event{})
	require.ErrorIs(t, err, errServerFailure)
	assert.Contains(t, err.Error(), "502 Bad Gateway")
	assert.Equal(t, 4, calls)
}

func TestHTTPObserverRetriesConnectionFailure(t *testing.T) {
	observer := NewHTTPObserver("http://audit.test/events")
	observer.retryIntervals = []time.Duration{0, 0, 0}
	calls := 0
	observer.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("connection refused")
	})}

	err := observer.Observe(context.Background(), Event{})
	require.ErrorIs(t, err, errConnectionFailure)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, 4, calls)
}

func TestHTTPObserverDoesNotRetryClientFailure(t *testing.T) {
	observer := NewHTTPObserver("http://audit.test/events")
	observer.retryIntervals = []time.Duration{0, 0, 0}
	calls := 0
	observer.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return response(http.StatusBadRequest), nil
	})}

	err := observer.Observe(context.Background(), Event{})
	require.Error(t, err)
	assert.NotErrorIs(t, err, errServerFailure)
	assert.Equal(t, 1, calls)
}

func TestHTTPObserverInvalidURL(t *testing.T) {
	assert.Error(t, NewHTTPObserver("://invalid").Observe(context.Background(), Event{}))
}

func newTestHTTPObserver(transport http.RoundTripper) *HTTPObserver {
	observer := NewHTTPObserver("http://audit.test/events")
	observer.client = &http.Client{Transport: transport}
	observer.retryIntervals = []time.Duration{0, 0, 0}
	return observer
}

func response(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func TestSubjectNotify(t *testing.T) {
	subject := NewSubject(context.Background(), zap.NewNop().Sugar())
	t.Cleanup(subject.Close)

	firstObserverEvents := make(chan Event, 1)
	secondObserverEvents := make(chan Event, 1)

	subject.AddObserver(observerFunc(func(event Event) error {
		firstObserverEvents <- event
		return nil
	}))
	subject.AddObserver(observerFunc(func(event Event) error {
		secondObserverEvents <- event
		return errors.New("observer failed")
	}))

	want := Event{TS: 1, Metrics: []string{"temperature"}, IPAddress: "127.0.0.1"}
	subject.Notify(want)

	assert.Equal(t, want, receiveEvent(t, firstObserverEvents))
	assert.Equal(t, want, receiveEvent(t, secondObserverEvents))
}

func TestSubjectCloseDrainsBufferedEvents(t *testing.T) {
	subject := NewSubject(context.Background(), zap.NewNop().Sugar())
	processed := make(chan Event, 3)
	subject.AddObserver(observerFunc(func(event Event) error {
		processed <- event
		return nil
	}))

	for i := range 3 {
		subject.Notify(Event{TS: int64(i)})
	}

	subject.Close()
	subject.Close()

	require.Len(t, processed, 3)
	subject.Notify(Event{TS: 4})
	assert.Len(t, processed, 3)
}

func TestSubjectNotifyDropsEventWhenChannelIsFull(t *testing.T) {
	subject := NewSubject(context.Background(), zap.NewNop().Sugar())
	started := make(chan struct{})
	release := make(chan struct{})
	processed := make(chan Event, 6)
	var startOnce sync.Once

	subject.AddObserver(observerFunc(func(event Event) error {
		startOnce.Do(func() { close(started) })
		<-release
		processed <- event
		return nil
	}))

	subject.Notify(Event{TS: 1})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("observer did not start")
	}

	for i := 2; i <= 6; i++ {
		subject.Notify(Event{TS: int64(i)})
	}
	subject.Notify(Event{TS: 7})

	close(release)
	subject.Close()

	assert.Len(t, processed, 6)
}

func receiveEvent(t *testing.T, events <-chan Event) Event {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatal("observer did not receive event")
		return Event{}
	}
}
