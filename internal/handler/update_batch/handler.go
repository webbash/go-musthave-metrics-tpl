// Package update_batch implements the JSON endpoint for batch metric updates.
package update_batch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
)

// Handler serves the batch metric update endpoint.
type Handler struct {
	service metricsService
	logger  *zap.SugaredLogger
	subject *audit.Subject
}

// NewHandler creates a handler for the batch metric update endpoint.
func NewHandler(service metricsService, logger *zap.SugaredLogger, subject *audit.Subject) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		subject: subject,
	}
}

// ServeHTTP handles POST /updates requests containing a JSON metric slice.
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

	if err := h.service.UpdateMany(r.Context(), metrics); err != nil {
		if errors.Is(err, service.ErrUnknownMetricType) {
			http.Error(res, "unknown metric type", http.StatusNotImplemented)
			return
		}

		if errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(res, "error parse metric value", http.StatusBadRequest)
			return
		}

		h.logger.Errorw("error updating metrics", "err", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ipAddress = r.RemoteAddr
	}

	metricNames := make([]string, len(metrics))
	for i, metric := range metrics {
		metricNames[i] = metric.ID
	}

	event := audit.Event{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: ipAddress,
	}
	h.subject.Notify(event)

	res.WriteHeader(http.StatusOK)
}
