package migrator

import (
	"database/sql"
	"io/fs"
	"log/slog"
	"testing"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
	"github.com/tallycat/tallycat/internal/repository/duckdb"
)

func TestListEmbeddedFiles(t *testing.T) {
	// List all embedded files
	err := fs.WalkDir(duckdb.EmbeddedMigrations, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		t.Logf("Found embedded file: %s", path)
		return nil
	})
	require.NoError(t, err)
}

func TestLoadMigrations(t *testing.T) {
	// Create a test database
	db, err := sql.Open("duckdb", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create migrator
	m := New(db, slog.Default())

	// Load migrations
	migrations, err := m.LoadMigrations(duckdb.EmbeddedMigrations)
	require.NoError(t, err)
	require.NotEmpty(t, migrations)

	// Verify migration content
	for _, migration := range migrations {
		t.Logf("Migration %d: %s", migration.Version, migration.Name)
		t.Logf("Up SQL length: %d", len(migration.UpSQL))
		t.Logf("Down SQL length: %d", len(migration.DownSQL))

		require.NotEmpty(t, migration.UpSQL, "Up SQL should not be empty")
		require.NotEmpty(t, migration.DownSQL, "Down SQL should not be empty")
	}
}
