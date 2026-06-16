package update_v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type Handler struct {
	service metricsService
}

func NewHandler(service metricsService) *Handler {
	return &Handler{
		service: service,
	}
}

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
		http.Error(res, "MType and Value are required", http.StatusBadRequest)
		return
	}

	if metric.Value == nil && metric.Delta == nil {
		http.Error(res, "Value or Delta are required", http.StatusBadRequest)
		return
	}

	var value string
	if metric.MType == models.Counter {
		if metric.Value != nil {
			value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		} else if metric.Delta != nil {
			value = strconv.FormatInt(*metric.Delta, 10)
		}
	}

	err = h.service.Update(r.Context(), metric.MType, metric.ID, value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
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
