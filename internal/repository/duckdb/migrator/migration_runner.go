package migrator

import (
	"database/sql"
	"log/slog"

	"github.com/tallycat/tallycat/internal/repository/duckdb"
)

// ApplyMigrations runs all pending migrations in the database
func ApplyMigrations(db *sql.DB) error {
	logger := slog.Default()
	migrator := New(db, logger)

	// Initialize the migrations table
	if err := migrator.Initialize(); err != nil {
		return err
	}

	// Load migrations from the embedded filesystem
	migrations, err := migrator.LoadMigrations(duckdb.EmbeddedMigrations)
	if err != nil {
		return err
	}

	// Apply all pending migrations
	if err := migrator.Migrate(migrations); err != nil {
		return err
	}

	return nil
}

// RollbackLastMigration rolls back the most recently applied migration
func RollbackLastMigration(db *sql.DB) error {
	logger := slog.Default()
	migrator := New(db, logger)

	// Load migrations from the embedded filesystem
	migrations, err := migrator.LoadMigrations(duckdb.EmbeddedMigrations)
	if err != nil {
		return err
	}

	// Rollback the last migration
	if err := migrator.Rollback(migrations); err != nil {
		return err
	}

	return nil
}
