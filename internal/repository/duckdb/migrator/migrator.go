package migrator

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

// New creates a new Migrator instance
func New(db *sql.DB, logger *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// LoadMigrations loads migrations from the embedded filesystem
func (m *Migrator) LoadMigrations(migrationsFS embed.FS) ([]Migration, error) {
	var migrations []Migration
	migrationMap := make(map[int]*Migration)

	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		m.logger.Debug("processing migration file", "path", path)

		// Parse version and name from filename (e.g., 000001_create_users.up.sql)
		parts := strings.Split(d.Name(), "_")
		if len(parts) < 3 {
			return fmt.Errorf("invalid migration filename: %s", d.Name())
		}

		version := 0
		fmt.Sscanf(parts[0], "%d", &version)
		if version == 0 {
			return fmt.Errorf("invalid migration version in filename: %s", d.Name())
		}

		// Read migration file
		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		m.logger.Debug("read migration file", "path", path, "content_length", len(content))

		// Get or create migration
		migration, exists := migrationMap[version]
		if !exists {
			name := strings.Join(parts[1:len(parts)-2], "_")
			migration = &Migration{
				Version: version,
				Name:    name,
			}
			migrationMap[version] = migration
		}

		// Set SQL content based on file type
		if strings.HasSuffix(path, ".up.sql") {
			migration.UpSQL = string(content)
			m.logger.Debug("set up migration SQL", "version", version, "sql_length", len(migration.UpSQL))
		} else if strings.HasSuffix(path, ".down.sql") {
			migration.DownSQL = string(content)
			m.logger.Debug("set down migration SQL", "version", version, "sql_length", len(migration.DownSQL))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert map to slice after all SQL content is set
	for _, migration := range migrationMap {
		migrations = append(migrations, *migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Log loaded migrations
	for _, migration := range migrations {
		m.logger.Debug("loaded migration",
			"version", migration.Version,
			"name", migration.Name,
			"up_sql_length", len(migration.UpSQL),
			"down_sql_length", len(migration.DownSQL))
	}

	return migrations, nil
}

// GetAppliedMigrations returns a map of applied migrations
func (m *Migrator) GetAppliedMigrations() (map[int]Migration, error) {
	rows, err := m.db.Query("SELECT version, name, applied_at FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]Migration)
	for rows.Next() {
		var migration Migration
		if err := rows.Scan(&migration.Version, &migration.Name, &migration.AppliedAt); err != nil {
			return nil, err
		}
		applied[migration.Version] = migration
	}
	return applied, rows.Err()
}

// Migrate applies all pending migrations
func (m *Migrator) Migrate(migrations []Migration) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; exists {
			continue
		}

		m.logger.Info("applying migration",
			"version", migration.Version,
			"name", migration.Name,
			"sql_length", len(migration.UpSQL))

		if len(migration.UpSQL) == 0 {
			return fmt.Errorf("empty SQL for migration %d", migration.Version)
		}

		tx, err := m.db.Begin()
		if err != nil {
			return err
		}

		// Apply migration
		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Record migration
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			migration.Version,
			migration.Name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		m.logger.Info("applied migration", "version", migration.Version, "name", migration.Name)
	}

	return nil
}

// Rollback rolls back the last applied migration
func (m *Migrator) Rollback(migrations []Migration) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Find the last applied migration
	var lastMigration *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if _, exists := applied[migrations[i].Version]; exists {
			lastMigration = &migrations[i]
			break
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	m.logger.Info("rolling back migration", "version", lastMigration.Version, "name", lastMigration.Name)

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	// Apply down migration
	if _, err := tx.Exec(lastMigration.DownSQL); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration %d: %w", lastMigration.Version, err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", lastMigration.Version); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %d: %w", lastMigration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback of migration %d: %w", lastMigration.Version, err)
	}

	m.logger.Info("rolled back migration", "version", lastMigration.Version, "name", lastMigration.Name)
	return nil
}
