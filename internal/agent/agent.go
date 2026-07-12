package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/crypto"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/retry"
)

type Agent struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	ms             runtime.MemStats
	httpClient     *http.Client
	basicURL       string
	signer         *crypto.Sha256Signer
}

func NewAgent(basicURL string, pollInterval, reportInterval time.Duration, httpClient *http.Client, signer *crypto.Sha256Signer) *Agent {
	addr := basicURL
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return &Agent{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		httpClient:     httpClient,
		basicURL:       addr,
		signer:         signer,
	}
}

func (a *Agent) ReadMetrics() {
	runtime.ReadMemStats(&a.ms)

	a.gaugeMetrics["Alloc"] = float64(a.ms.Alloc)
	a.gaugeMetrics["BuckHashSys"] = float64(a.ms.BuckHashSys)
	a.gaugeMetrics["Frees"] = float64(a.ms.Frees)
	a.gaugeMetrics["GCCPUFraction"] = float64(a.ms.GCCPUFraction)
	a.gaugeMetrics["GCSys"] = float64(a.ms.GCSys)
	a.gaugeMetrics["HeapAlloc"] = float64(a.ms.HeapAlloc)
	a.gaugeMetrics["HeapIdle"] = float64(a.ms.HeapIdle)
	a.gaugeMetrics["HeapInuse"] = float64(a.ms.HeapInuse)
	a.gaugeMetrics["HeapObjects"] = float64(a.ms.HeapObjects)
	a.gaugeMetrics["HeapReleased"] = float64(a.ms.HeapReleased)
	a.gaugeMetrics["HeapSys"] = float64(a.ms.HeapSys)
	a.gaugeMetrics["LastGC"] = float64(a.ms.LastGC)
	a.gaugeMetrics["Lookups"] = float64(a.ms.Lookups)
	a.gaugeMetrics["MCacheInuse"] = float64(a.ms.MCacheInuse)
	a.gaugeMetrics["MCacheSys"] = float64(a.ms.MCacheSys)
	a.gaugeMetrics["MSpanInuse"] = float64(a.ms.MSpanInuse)
	a.gaugeMetrics["MSpanSys"] = float64(a.ms.MSpanSys)
	a.gaugeMetrics["Mallocs"] = float64(a.ms.Mallocs)
	a.gaugeMetrics["NextGC"] = float64(a.ms.NextGC)
	a.gaugeMetrics["NumForcedGC"] = float64(a.ms.NumForcedGC)
	a.gaugeMetrics["NumGC"] = float64(a.ms.NumGC)
	a.gaugeMetrics["OtherSys"] = float64(a.ms.OtherSys)
	a.gaugeMetrics["PauseTotalNs"] = float64(a.ms.PauseTotalNs)
	a.gaugeMetrics["StackInuse"] = float64(a.ms.StackInuse)
	a.gaugeMetrics["StackSys"] = float64(a.ms.StackSys)
	a.gaugeMetrics["Sys"] = float64(a.ms.Sys)
	a.gaugeMetrics["TotalAlloc"] = float64(a.ms.TotalAlloc)
	a.gaugeMetrics["RandomValue"] = rand.Float64()

	a.counterMetrics["PollCount"] += 1
}

func (a *Agent) SendMetrics() error {
	var metricsToSend []models.Metrics
	for metricName, value := range a.gaugeMetrics {
		metricsToSend = append(metricsToSend, models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		})
	}
	for metricName, value := range a.counterMetrics {
		metricsToSend = append(metricsToSend, models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		})
	}

	err := retry.Do(context.Background(), func() error {
		return a.sendUpdateMetrics(metricsToSend)
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

func (a *Agent) sendUpdateMetrics(metric []models.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(body); err != nil {
		return fmt.Errorf("gzip write: %w", err)
	}
	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gzip closing: %w", err)
	}

	updateUrl, err := url.JoinPath(a.basicURL, "/updates")
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
