package agent

import (
	"net/http"
	"testing"
	"time"

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

func TestAgent_ReadMetrics_NotZero(t *testing.T) {
	agent := NewAgent("http://localhost", 1*time.Second, 1*time.Second, &http.Client{})

	agent.ReadMetrics()

	assert.Len(t, agent.gaugeMetrics, len(metricsList)-1) // -1 т.к нам нужно учесть ещё и counter
	assert.Len(t, agent.counterMetrics, 1)

	assert.Equal(t, int64(1), agent.counterMetrics["PollCount"])
}
