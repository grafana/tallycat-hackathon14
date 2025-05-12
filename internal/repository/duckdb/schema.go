package duckdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/repository/query"
)

type SchemaRepository struct {
	pool   *ConnectionPool
	logger *slog.Logger
}

func NewSchemaRepository(pool *ConnectionPool, logger *slog.Logger) *SchemaRepository {
	return &SchemaRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *SchemaRepository) RegisterSchema(ctx context.Context, schema *repository.Schema) error {
	// Start a transaction
	tx, err := r.pool.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or update schema_core
	coreQuery := `
		INSERT INTO schema_core (
			schema_id, signal_type, signal_key, scope_name, scope_version,
			schema_url, seen_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (schema_id) DO UPDATE SET
			seen_count = schema_core.seen_count + excluded.seen_count,
			updated_at = excluded.updated_at
		WHERE excluded.updated_at > schema_core.updated_at;
	`

	_, err = tx.ExecContext(ctx, coreQuery,
		schema.SchemaID,
		schema.SignalType,
		schema.SignalKey,
		schema.ScopeName,
		schema.ScopeVersion,
		schema.SchemaURL,
		schema.SeenCount,
		schema.CreatedAt,
		schema.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert/update schema_core: %w", err)
	}

	// Insert or update schema_details
	fieldTypesJSON, err := json.Marshal(schema.FieldTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal field types: %w", err)
	}

	fieldSourcesJSON, err := json.Marshal(schema.FieldSources)
	if err != nil {
		return fmt.Errorf("failed to marshal field sources: %w", err)
	}

	fieldCardinalityJSON, err := json.Marshal(schema.FieldCardinality)
	if err != nil {
		return fmt.Errorf("failed to marshal field cardinality: %w", err)
	}

	fieldNamesArray := buildDuckDBStringArray(schema.FieldNames)

	detailsQuery := `
		INSERT INTO schema_details (
			schema_id, metric_type, unit, field_names,
			field_types, field_sources, field_cardinality
		) VALUES (?, ?, ?, ` + fieldNamesArray + `, ?, ?, ?)
		ON CONFLICT (schema_id) DO UPDATE SET
			metric_type = excluded.metric_type,
			unit = excluded.unit,
			field_names = excluded.field_names,
			field_types = excluded.field_types,
			field_sources = excluded.field_sources,
			field_cardinality = excluded.field_cardinality;
	`

	_, err = tx.ExecContext(ctx, detailsQuery,
		schema.SchemaID,
		schema.MetricType,
		schema.Unit,
		fieldTypesJSON,
		fieldSourcesJSON,
		fieldCardinalityJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert/update schema_details: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListSchemas returns a paginated, filtered, and searched list of schemas
func (r *SchemaRepository) ListSchemas(ctx context.Context, params query.ListQueryParams) ([]*repository.Schema, int, error) {
	// Build WHERE clause
	var args []interface{}
	where := ""
	if params.FilterType != "" && params.FilterType != "all" {
		where += " AND c.signal_type = ?"
		args = append(args, params.FilterType)
	}
	if params.Search != "" {
		where += " AND (c.schema_id LIKE ? OR c.signal_key LIKE ? OR c.scope_name LIKE ? OR d.metric_type LIKE ? OR d.unit LIKE ?)"
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Count total
	countQuery := `
		SELECT COUNT(DISTINCT (c.signal_type, c.signal_key))
		FROM schema_core c
		JOIN schema_details d ON c.schema_id = d.schema_id
		WHERE 1=1` + where
	db := r.pool.GetConnection()

	// Use context timeout for count query
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var total int
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count schemas: %w", err)
	}

	// If no results, return early
	if total == 0 {
		return []*repository.Schema{}, 0, nil
	}

	// Build main query with grouping
	queryStr := `
		WITH latest_schemas AS (
			SELECT 
				c.schema_id,
				c.signal_type,
				c.signal_key,
				c.scope_name,
				c.scope_version,
				d.metric_type,
				d.unit,
				d.field_names,
				d.field_types,
				d.field_sources,
				d.field_cardinality,
				c.seen_count,
				c.created_at,
				c.updated_at,
				COUNT(*) OVER (PARTITION BY c.signal_type, c.signal_key) as version_count,
				ROW_NUMBER() OVER (
					PARTITION BY c.signal_type, c.signal_key 
					ORDER BY c.updated_at DESC
				) as rn
			FROM schema_core c
			JOIN schema_details d ON c.schema_id = d.schema_id
			WHERE 1=1` + where + `
		)
		SELECT 
			schema_id, signal_type, signal_key, scope_name, scope_version,
			metric_type, unit, field_names, field_types,
			field_sources, field_cardinality, seen_count,
			created_at, updated_at, version_count
		FROM latest_schemas
		WHERE rn = 1
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?`

	// Add pagination parameters
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	// Use context timeout for main query
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	// Scan results
	var schemas []*repository.Schema
	for rows.Next() {
		var schema repository.Schema
		var fieldNamesRaw []interface{}
		var metricType, unit sql.NullString
		var fieldTypesJSON, fieldSourcesJSON, fieldCardinalityJSON string
		var versionCount int

		if err := rows.Scan(
			&schema.SchemaID,
			&schema.SignalType,
			&schema.SignalKey,
			&schema.ScopeName,
			&schema.ScopeVersion,
			&metricType,
			&unit,
			&fieldNamesRaw,
			&fieldTypesJSON,
			&fieldSourcesJSON,
			&fieldCardinalityJSON,
			&schema.SeenCount,
			&schema.CreatedAt,
			&schema.UpdatedAt,
			&versionCount,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan schema row: %w", err)
		}

		// Convert field names
		fieldNames := make([]string, len(fieldNamesRaw))
		for i, v := range fieldNamesRaw {
			if s, ok := v.(string); ok {
				fieldNames[i] = s
			} else {
				fieldNames[i] = ""
			}
		}
		schema.FieldNames = fieldNames

		// Handle nullable fields
		if metricType.Valid {
			schema.MetricType = &metricType.String
		}
		if unit.Valid {
			schema.Unit = &unit.String
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(fieldTypesJSON), &schema.FieldTypes); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal field types: %w", err)
		}
		if err := json.Unmarshal([]byte(fieldSourcesJSON), &schema.FieldSources); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal field sources: %w", err)
		}
		if err := json.Unmarshal([]byte(fieldCardinalityJSON), &schema.FieldCardinality); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal field cardinality: %w", err)
		}

		// Store version count in a custom field
		schema.SchemaVersionCount = versionCount

		schemas = append(schemas, &schema)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, total, nil
}
