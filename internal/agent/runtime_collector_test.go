package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var metricsList = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys",
	"LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
	"MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
	"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
	"RandomValue", "PollCount",
}

func TestRuntimeCollector_Collect_NotZero(t *testing.T) {
	collector := NewRuntimeCollector()

	collector.Collect()

	assert.Len(t, collector.gaugeMetrics, len(metricsList)-1) // -1 т.к нам нужно учесть ещё и counter
	assert.Len(t, collector.counterMetrics, 1)

	assert.Equal(t, int64(1), collector.counterMetrics["PollCount"])
}
