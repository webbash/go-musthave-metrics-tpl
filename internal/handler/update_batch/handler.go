package update_batch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

type Handler struct {
	service MetricsService
}

func NewHandler(service MetricsService) *Handler {
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

	var metrics []models.Metrics
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateBatch(r.Context(), metrics); err != nil {
		if errors.Is(err, service.ErrUnknownMetricType) {
			http.Error(res, "unknown metric type", http.StatusNotImplemented)
			return
		}

		if errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(res, "error parse metric value", http.StatusBadRequest)
			return
		}

		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.service.UpdateBatch(r.Context(), metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}
