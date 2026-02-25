package update

import (
	"net/http"
	"strconv"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type Handler struct {
	service service
}

func NewHandler(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	mType := req.PathValue("metricType")
	mName := req.PathValue("metricName")
	mValue := req.PathValue("metricValue")

	switch mType {
	case models.Counter:
		intVal, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		h.service.IncrementCounter(mName, intVal)
	case models.Gauge:
		floatVal, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		h.service.UpdateGauge(mName, floatVal)
	default:
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}
