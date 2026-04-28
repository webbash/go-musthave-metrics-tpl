package update

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
	mValue := chi.URLParam(r, "metricValue")

	if err := h.service.Update(r.Context(), mType, mName, mValue); err != nil {
		if errors.Is(err, service.ErrUnknownMetricType) {
			http.Error(res, "unknown metric type", http.StatusNotImplemented)
			return
		}

		if errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(res, "error parse metric value", http.StatusBadRequest)
			return
		}

		http.Error(res, "unknown metric type", http.StatusNotImplemented)
		return
	}

	res.WriteHeader(http.StatusOK)
}
