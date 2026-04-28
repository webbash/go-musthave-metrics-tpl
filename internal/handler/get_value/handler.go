package get_value

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
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
	mType := chi.URLParam(r, "metricType")
	mName := chi.URLParam(r, "metricName")

	value, err := h.service.Get(r.Context(), mType, mName)
	if err != nil {
		if errors.Is(err, service.ErrMetricNotFound) {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}

		http.Error(res, "unknown metric type", http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(value))
}
