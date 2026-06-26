package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGConnector struct {
	DSN string
}

func NewPGConnector(dsn string) *PGConnector {
	return &PGConnector{DSN: dsn}
}

func (c *PGConnector) Connect() (*sql.DB, error) {
	return sql.Open("pgx", c.DSN)
}
