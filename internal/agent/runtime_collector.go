package agent

import (
	"math/rand"
	"runtime"
	"sync"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type RuntimeCollector struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	mu             sync.RWMutex
}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (rc *RuntimeCollector) Collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.gaugeMetrics["Alloc"] = float64(ms.Alloc)
	rc.gaugeMetrics["BuckHashSys"] = float64(ms.BuckHashSys)
	rc.gaugeMetrics["Frees"] = float64(ms.Frees)
	rc.gaugeMetrics["GCCPUFraction"] = float64(ms.GCCPUFraction)
	rc.gaugeMetrics["GCSys"] = float64(ms.GCSys)
	rc.gaugeMetrics["HeapAlloc"] = float64(ms.HeapAlloc)
	rc.gaugeMetrics["HeapIdle"] = float64(ms.HeapIdle)
	rc.gaugeMetrics["HeapInuse"] = float64(ms.HeapInuse)
	rc.gaugeMetrics["HeapObjects"] = float64(ms.HeapObjects)
	rc.gaugeMetrics["HeapReleased"] = float64(ms.HeapReleased)
	rc.gaugeMetrics["HeapSys"] = float64(ms.HeapSys)
	rc.gaugeMetrics["LastGC"] = float64(ms.LastGC)
	rc.gaugeMetrics["Lookups"] = float64(ms.Lookups)
	rc.gaugeMetrics["MCacheInuse"] = float64(ms.MCacheInuse)
	rc.gaugeMetrics["MCacheSys"] = float64(ms.MCacheSys)
	rc.gaugeMetrics["MSpanInuse"] = float64(ms.MSpanInuse)
	rc.gaugeMetrics["MSpanSys"] = float64(ms.MSpanSys)
	rc.gaugeMetrics["Mallocs"] = float64(ms.Mallocs)
	rc.gaugeMetrics["NextGC"] = float64(ms.NextGC)
	rc.gaugeMetrics["NumForcedGC"] = float64(ms.NumForcedGC)
	rc.gaugeMetrics["NumGC"] = float64(ms.NumGC)
	rc.gaugeMetrics["OtherSys"] = float64(ms.OtherSys)
	rc.gaugeMetrics["PauseTotalNs"] = float64(ms.PauseTotalNs)
	rc.gaugeMetrics["StackInuse"] = float64(ms.StackInuse)
	rc.gaugeMetrics["StackSys"] = float64(ms.StackSys)
	rc.gaugeMetrics["Sys"] = float64(ms.Sys)
	rc.gaugeMetrics["TotalAlloc"] = float64(ms.TotalAlloc)
	rc.gaugeMetrics["RandomValue"] = rand.Float64()

	rc.counterMetrics["PollCount"] += 1
}

func (rc *RuntimeCollector) Snapshot() []models.Metrics {
	var metricsToSend []models.Metrics

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	for metricName, value := range rc.gaugeMetrics {
		metricsToSend = append(metricsToSend, models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		})
	}
	for metricName, value := range rc.counterMetrics {
		metricsToSend = append(metricsToSend, models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		})
	}

	return metricsToSend
}
