package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
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
	return &Agent{
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		httpClient:     httpClient,
		basicURL:       basicURL,
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

func (a *Agent) SendMetrics() {
	for metricName, value := range a.gaugeMetrics {
		url := fmt.Sprintf("http://%s/update/%s/%s/%v", a.basicURL, models.Gauge, metricName, value)
		a.sendPostRequest(url)
	}

	for metricName, value := range a.counterMetrics {
		url := fmt.Sprintf("%s/update/%s/%s/%v", a.basicURL, models.Counter, metricName, value)
		a.sendPostRequest(url)
	}
}

func (a *Agent) sendPostRequest(url string) {
	request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Println(fmt.Errorf("failed to create request: %v", err))
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := a.httpClient.Do(request)

	if err != nil {
		log.Println(fmt.Errorf("failed to send request: %v", err))
	}

	err = response.Body.Close()
	if err != nil {
		log.Println(fmt.Errorf("failed to close response body: %v", err))
	}
}
