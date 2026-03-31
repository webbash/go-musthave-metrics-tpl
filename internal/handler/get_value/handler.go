package get_value

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type Handler struct {
	storage storage
}

func NewHandler(storage storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h Handler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "metricType")
	mName := chi.URLParam(r, "metricName")

	switch mType {
	case models.Counter:
		intVal, ok := h.storage.GetCounter(mName)
		if !ok {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(strconv.FormatInt(intVal, 10)))
		return
	case models.Gauge:
		floatVal, ok := h.storage.GetGauge(mName)
		if !ok {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(strconv.FormatFloat(floatVal, 'f', -1, 64)))
		return
	default:
		http.Error(res, "unknown metric type", http.StatusBadRequest)
		return
	}
}
