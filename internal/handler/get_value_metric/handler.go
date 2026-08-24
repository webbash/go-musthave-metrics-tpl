// Package get_value_metric implements the JSON endpoint for reading one metric.
package get_value_metric

import (
	"bytes"
	"encoding/json"
	"net/http"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

// Handler serves the JSON metric value endpoint.
type Handler struct {
	service metricsService
}

// NewHandler creates a handler for the JSON metric value endpoint.
func NewHandler(service metricsService) *Handler {
	return &Handler{
		service: service,
	}
}

// ServeHTTP handles POST /value requests containing a metric descriptor in JSON.
func (h Handler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var metric models.Metrics
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if metric.MType == "" || metric.ID == "" {
		http.Error(res, "MType and Value are required", http.StatusNotFound)
		return
	}

	metric, err = h.service.GetMetric(r.Context(), metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
