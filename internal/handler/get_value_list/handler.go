package get_value_list

import (
	"fmt"
	"net/http"
	"strings"
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

	html := buildHTMLBuilder(gauges, counters)

	res.Write([]byte(html))
}

func buildHTMLConcat(
	gauges map[string]float64,
	counters map[string]int64,
) string {
	html := "<html><body><h1>Metrics List</h1><ul>"

	for name, value := range gauges {
		html += fmt.Sprintf("<li>%s: %g</li>", name, value)
	}

	for name, value := range counters {
		html += fmt.Sprintf("<li>%s: %d</li>", name, value)
	}

	html += "</ul></body></html>"
	return html
}

func buildHTMLBuilder(
	gauges map[string]float64,
	counters map[string]int64,
) string {
	w := strings.Builder{}

	w.WriteString("<html><body><h1>Metrics List</h1><ul>")

	for name, value := range gauges {
		w.WriteString(fmt.Sprintf("<li>%s: %g</li>", name, value))
	}

	for name, value := range counters {
		w.WriteString(fmt.Sprintf("<li>%s: %d</li>", name, value))
	}

	w.WriteString("</ul></body></html>")

	return w.String()
}
