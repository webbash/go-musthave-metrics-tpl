// Package update implements the legacy endpoint for updating one metric.
package update

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

// Handler serves the legacy metric update endpoint.
type Handler struct {
	service metricsService
	subject *audit.Subject
	logger  *zap.SugaredLogger
}

// NewHandler creates a handler for the legacy metric update endpoint.
func NewHandler(service metricsService, subject *audit.Subject, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		subject: subject,
		logger:  logger,
	}
}

// ServeHTTP handles POST /update/{metricType}/{metricName}/{metricValue} requests.
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

	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ipAddress = r.RemoteAddr
	}

	event := audit.Event{
		TS:        time.Now().Unix(),
		Metrics:   []string{mName},
		IPAddress: ipAddress,
	}
	h.subject.Notify(event)

	res.WriteHeader(http.StatusOK)
}
