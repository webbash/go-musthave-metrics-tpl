package ping_db

import (
	"fmt"
	"net/http"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/config/db"
)

type Handler struct {
	connector db.Connector
}

func NewHandler(connector db.Connector) *Handler {
	return &Handler{
		connector: connector,
	}
}

func (h Handler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	db, err := h.connector.Connect()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	err = db.Ping()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf("PONG")))
}
