package db

import "database/sql"

type Connector interface {
	Connect() (*sql.DB, error)
}
