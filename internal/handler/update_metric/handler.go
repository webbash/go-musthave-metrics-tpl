package update_metric

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/audit"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	service metricsService
	subject *audit.Subject
	logger  *zap.SugaredLogger
}

func NewHandler(service metricsService, subject *audit.Subject, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		subject: subject,
		logger:  logger,
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

	metric, err = h.service.UpdateMetric(r.Context(), metric)
	if err != nil {
		if errors.Is(err, service.ErrUnknownMetricType) || errors.Is(err, service.ErrInvalidMetricValue) {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ipAddress = r.RemoteAddr
	}

	event := audit.Event{
		TS:        time.Now().Unix(),
		Metrics:   []string{metric.MType},
		IPAddress: ipAddress,
	}
	err = h.subject.Notify(event)
	if err != nil {
		h.logger.Errorw("failed to notify subject", "err", err)
	}

	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
