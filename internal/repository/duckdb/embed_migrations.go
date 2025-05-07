package duckdb

import (
	"embed"
)

// EmbeddedMigrations contains all SQL migration files embedded in the binary
//
//go:embed migrations/*.sql
var EmbeddedMigrations embed.FS
