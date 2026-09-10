// Package db provides database connection helpers.
package db

import "database/sql"

type Connector interface {
	Connect() (*sql.DB, error)
}
