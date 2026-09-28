package agent

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"go.uber.org/zap"
)

func TestSendSnapshotAfterCollectorsStop(t *testing.T) {
	collectCtx, stop := context.WithCancel(t.Context())
	stop()
	var requests int
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.NoError(t, req.Context().Err())
		reader, err := gzip.NewReader(req.Body)
		require.NoError(t, err)
		defer reader.Close()
		var metrics []models.Metrics
		require.NoError(t, json.NewDecoder(reader).Decode(&metrics))
		require.Len(t, metrics, 1)
		require.Equal(t, "Alloc", metrics[0].ID)
		require.Equal(t, 150.0, *metrics[0].Value)
		requests++
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	a := NewAgent("http://metrics.test", time.Hour, time.Hour, client, nil, 2, zap.NewNop().Sugar(), nil)
	a.runtimeCollector.gaugeMetrics["Alloc"] = 150
	stopBatches := make(chan struct{})
	collectors, workers := a.Start(collectCtx, t.Context(), stopBatches)
	collectors.Wait()
	close(stopBatches)
	workers.Wait()
	require.NoError(t, a.SendSnapshot(t.Context()))
	require.Equal(t, 1, requests)
}

func TestWorkersFinishBeforeFinalSend(t *testing.T) {
	collectCtx, stop := context.WithCancel(t.Context())
	defer stop()
	sendCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var calls atomic.Int32
	var finished atomic.Int32
	var premature atomic.Bool
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := calls.Add(1)
		if n <= 2 {
			started <- struct{}{}
			select {
			case <-release:
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
			finished.Add(1)
		} else if finished.Load() != 2 {
			premature.Store(true)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	a := NewAgent("http://metrics.test", time.Hour, time.Millisecond, client, nil, 2, zap.NewNop().Sugar(), nil)
	done := make(chan struct{})
	stopBatches := make(chan struct{})
	collectors, workers := a.Start(collectCtx, sendCtx, stopBatches)
	finalErr := make(chan error, 1)
	go func() {
		collectors.Wait()
		close(stopBatches)
		workers.Wait()
		finalErr <- a.SendSnapshot(sendCtx)
		close(done)
	}()
	for range 2 {
		select {
		case <-started:
		case <-sendCtx.Done():
			t.Fatal("workers did not start requests")
		}
	}
	stop()
	select {
	case <-done:
		t.Fatal("shutdown returned while requests were still running")
	case <-time.After(20 * time.Millisecond):
	}
	require.EqualValues(t, 2, calls.Load(), "final send must wait for workers")
	close(release)
	select {
	case <-done:
	case <-sendCtx.Done():
		t.Fatal("shutdown did not finish after requests completed")
	}
	require.NoError(t, <-finalErr)
	require.GreaterOrEqual(t, calls.Load(), int32(3), "final snapshot must be sent")
	require.False(t, premature.Load(), "final send overlapped earlier requests")
}
