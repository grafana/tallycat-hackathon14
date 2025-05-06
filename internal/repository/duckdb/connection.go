package duckdb

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/tallycat/tallycat/internal/repository"
)

type Config struct {
	DatabasePath    string        // Path to DuckDB database file
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

type ConnectionPool struct {
	db     *sql.DB
	config *Config
	logger *slog.Logger
}

func NewConnectionPool(config *Config, logger *slog.Logger) (repository.ConnectionProvider, error) {
	db, err := sql.Open("duckdb", config.DatabasePath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	return &ConnectionPool{
		db:     db,
		config: config,
		logger: logger,
	}, nil
}

func (p *ConnectionPool) Close() error {
	return p.db.Close()
}

func (p *ConnectionPool) GetConnection() *sql.DB {
	return p.db
}

func (p *ConnectionPool) HealthCheck() error {
	return p.db.Ping()
}

func (p *ConnectionPool) ValidateConnection() error {
	if p.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return p.db.Ping()
}

type ConnectionStats struct {
	OpenConnections int
	InUse           int
	Idle            int
	WaitCount       int64
	WaitDuration    time.Duration
}

func (p *ConnectionPool) GetStats() ConnectionStats {
	stats := p.db.Stats()
	return ConnectionStats{
		OpenConnections: stats.OpenConnections,
		InUse:           stats.InUse,
		Idle:            stats.Idle,
		WaitCount:       stats.WaitCount,
		WaitDuration:    stats.WaitDuration,
	}
}
