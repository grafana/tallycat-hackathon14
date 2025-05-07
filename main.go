/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"database/sql"
	"log/slog"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/tallycat/tallycat/cmd"
	"github.com/tallycat/tallycat/internal/repository/duckdb/migrator"
)

func main() {
	// Initialize database connection
	db, err := sql.Open("duckdb", "tallycat.db")
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return
	}
	defer db.Close()

	// Run migrations
	if err := migrator.ApplyMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		return
	}

	cmd.Execute()
}
