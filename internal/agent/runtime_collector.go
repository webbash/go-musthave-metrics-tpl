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

func (a *RuntimeCollector) Collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.gaugeMetrics["Alloc"] = float64(ms.Alloc)
	a.gaugeMetrics["BuckHashSys"] = float64(ms.BuckHashSys)
	a.gaugeMetrics["Frees"] = float64(ms.Frees)
	a.gaugeMetrics["GCCPUFraction"] = float64(ms.GCCPUFraction)
	a.gaugeMetrics["GCSys"] = float64(ms.GCSys)
	a.gaugeMetrics["HeapAlloc"] = float64(ms.HeapAlloc)
	a.gaugeMetrics["HeapIdle"] = float64(ms.HeapIdle)
	a.gaugeMetrics["HeapInuse"] = float64(ms.HeapInuse)
	a.gaugeMetrics["HeapObjects"] = float64(ms.HeapObjects)
	a.gaugeMetrics["HeapReleased"] = float64(ms.HeapReleased)
	a.gaugeMetrics["HeapSys"] = float64(ms.HeapSys)
	a.gaugeMetrics["LastGC"] = float64(ms.LastGC)
	a.gaugeMetrics["Lookups"] = float64(ms.Lookups)
	a.gaugeMetrics["MCacheInuse"] = float64(ms.MCacheInuse)
	a.gaugeMetrics["MCacheSys"] = float64(ms.MCacheSys)
	a.gaugeMetrics["MSpanInuse"] = float64(ms.MSpanInuse)
	a.gaugeMetrics["MSpanSys"] = float64(ms.MSpanSys)
	a.gaugeMetrics["Mallocs"] = float64(ms.Mallocs)
	a.gaugeMetrics["NextGC"] = float64(ms.NextGC)
	a.gaugeMetrics["NumForcedGC"] = float64(ms.NumForcedGC)
	a.gaugeMetrics["NumGC"] = float64(ms.NumGC)
	a.gaugeMetrics["OtherSys"] = float64(ms.OtherSys)
	a.gaugeMetrics["PauseTotalNs"] = float64(ms.PauseTotalNs)
	a.gaugeMetrics["StackInuse"] = float64(ms.StackInuse)
	a.gaugeMetrics["StackSys"] = float64(ms.StackSys)
	a.gaugeMetrics["Sys"] = float64(ms.Sys)
	a.gaugeMetrics["TotalAlloc"] = float64(ms.TotalAlloc)
	a.gaugeMetrics["RandomValue"] = rand.Float64()

	a.counterMetrics["PollCount"] += 1
}

func (a *RuntimeCollector) Snapshot() []models.Metrics {
	var metricsToSend []models.Metrics

	a.mu.RLock()
	defer a.mu.RUnlock()

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

	return metricsToSend
}
