package repository

import (
	"database/sql"
)

type ConnectionProvider interface {
	GetConnection() *sql.DB
	Close() error
	HealthCheck() error
}
