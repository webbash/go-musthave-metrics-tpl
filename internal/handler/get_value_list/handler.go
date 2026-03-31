package get_value_list

import (
	"fmt"
	"net/http"
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
	res.Header().Set("Content-Type", "text/html")

	html := "<html><body><h1>Metrics List</h1><ul>"

	for name, value := range h.storage.GetAllGauges() {
		html += fmt.Sprintf("<li>%s: %g</li>", name, value)
	}

	for name, value := range h.storage.GetAllCounters() {
		html += fmt.Sprintf("<li>%s: %d</li>", name, value)
	}

	html += "</ul></body></html>"

	res.Write([]byte(html))
}
