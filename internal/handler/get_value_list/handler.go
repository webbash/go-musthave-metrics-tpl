package get_value_list

import (
	"fmt"
	"net/http"
)

type Handler struct {
	repository metricsRepository
}

func NewHandler(repository metricsRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (h Handler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	res.Header().Set("Content-Type", "text/html")

	html := "<html><body><h1>Metrics List</h1><ul>"
	gauges, err := h.repository.GetAllGauges(r.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	counters, err := h.repository.GetAllCounters(r.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	for name, value := range gauges {
		html += fmt.Sprintf("<li>%s: %g</li>", name, value)
	}

	for name, value := range counters {
		html += fmt.Sprintf("<li>%s: %d</li>", name, value)
	}

	html += "</ul></body></html>"

	res.Write([]byte(html))
}
