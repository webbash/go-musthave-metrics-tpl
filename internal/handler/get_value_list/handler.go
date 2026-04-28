package get_value_list

import (
	"fmt"
	"net/http"
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
	res.Header().Set("Content-Type", "text/html")

	html := "<html><body><h1>Metrics List</h1><ul>"
	gauges, counters := h.service.GetAll(r.Context())

	for name, value := range gauges {
		html += fmt.Sprintf("<li>%s: %g</li>", name, value)
	}

	for name, value := range counters {
		html += fmt.Sprintf("<li>%s: %d</li>", name, value)
	}

	html += "</ul></body></html>"

	res.Write([]byte(html))
}
