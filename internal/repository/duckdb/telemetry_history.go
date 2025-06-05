package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tallycat/tallycat/internal/schema"
)

type TelemetryHistoryRepository struct {
	pool *ConnectionPool
}

func NewTelemetryHistoryRepository(pool *ConnectionPool) *TelemetryHistoryRepository {
	return &TelemetryHistoryRepository{
		pool: pool,
	}
}

func (r *TelemetryHistoryRepository) InsertTelemetryHistory(ctx context.Context, h *schema.TelemetryHistory) error {
	db := r.pool.GetConnection()
	query := `
		INSERT INTO telemetry_history (
			schema_key, version, timestamp, author, summary, status, snapshot, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	row := db.QueryRowContext(ctx, query,
		h.SchemaKey,
		h.Version,
		h.Timestamp,
		h.Author,
		h.Summary,
		h.Status,
		h.Snapshot,
		time.Now(),
		time.Now(),
	)
	return row.Scan(&h.Id, &h.CreatedAt, &h.UpdatedAt)
}

func (r *TelemetryHistoryRepository) ListTelemetryHistory(ctx context.Context, telemetryID string, page, pageSize int) ([]schema.TelemetryHistory, int, error) {
	db := r.pool.GetConnection()

	// Get total count
	countQuery := `SELECT COUNT(*) FROM telemetry_history WHERE schema_key = ?`
	total := 0
	if err := db.QueryRowContext(ctx, countQuery, telemetryID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count telemetry_history: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, schema_key, version, timestamp, author, summary, status, snapshot, created_at, updated_at
		FROM telemetry_history
		WHERE schema_key = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	rows, err := db.QueryContext(ctx, query, telemetryID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query telemetry_history: %w", err)
	}
	defer rows.Close()

	histories := []schema.TelemetryHistory{}
	for rows.Next() {
		var h schema.TelemetryHistory
		var author sql.NullString
		if err := rows.Scan(
			&h.Id,
			&h.SchemaKey,
			&h.Version,
			&h.Timestamp,
			&author,
			&h.Summary,
			&h.Status,
			&h.Snapshot,
			&h.CreatedAt,
			&h.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan telemetry_history row: %w", err)
		}
		if author.Valid {
			h.Author = &author.String
		} else {
			h.Author = nil
		}
		histories = append(histories, h)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating telemetry_history rows: %w", err)
	}
	return histories, total, nil
}
