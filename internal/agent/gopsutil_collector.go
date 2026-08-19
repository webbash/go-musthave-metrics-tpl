package agent

import (
	"fmt"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

// GopsutilCollector stores the latest host memory and CPU metrics.
type GopsutilCollector struct {
	gaugeMetrics map[string]float64
	mu           sync.RWMutex
}

// NewGopsutilCollector creates a collector for memory and CPU metrics.
func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{
		gaugeMetrics: make(map[string]float64),
	}
}

// Collect reads the current memory and per-CPU utilization values.
func (g *GopsutilCollector) Collect() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory: %w", err)
	}
	values, err := cpu.Percent(0, true)
	if err != nil {
		return fmt.Errorf("failed to get cpu percents: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.gaugeMetrics["TotalMemory"] = float64(v.Total)
	g.gaugeMetrics["FreeMemory"] = float64(v.Free)

	for i, cpuVal := range values {
		name := fmt.Sprintf("CPUutilization%d", i+1)
		g.gaugeMetrics[name] = cpuVal
	}

	return nil
}

// Snapshot returns the latest system metrics as model values.
func (g *GopsutilCollector) Snapshot() []models.Metrics {
	var metricsToSend []models.Metrics

	g.mu.RLock()
	defer g.mu.RUnlock()

	for metricName, value := range g.gaugeMetrics {
		metricsToSend = append(metricsToSend, models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		})
	}

	return metricsToSend
}
