package ping_db

import (
	"database/sql"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewHandler(db *sql.DB, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}

func (h Handler) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	err := h.db.PingContext(r.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorw("ping error", "err", err)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf("PONG")))
}
