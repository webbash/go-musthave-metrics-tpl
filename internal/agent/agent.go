package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type Agent struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	ms             runtime.MemStats
	httpClient     *http.Client
	basicURL       string
}

func NewAgent(basicURL string, pollInterval, reportInterval time.Duration, httpClient *http.Client) *Agent {
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
	var resultErr error

	for metricName, value := range a.gaugeMetrics {
		metric := models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		}

		err := a.sendUpdateMetric(metric)

		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("send gauge metric %s: %w", metricName, err))
		}
	}

	for metricName, value := range a.counterMetrics {
		metric := models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		}
		err := a.sendUpdateMetric(metric)
		if err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("send counter metric %s: %w", metricName, err))
		}
	}

	return resultErr
}

func (a *Agent) sendUpdateMetric(metric models.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}
	url := fmt.Sprintf("%s/update", a.basicURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
