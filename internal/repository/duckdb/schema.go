package duckdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tallycat/tallycat/internal/repository"
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

func (r *SchemaRepository) CreateSchemaTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS otel_schema_catalog (
			schema_id TEXT PRIMARY KEY,
			signal_type TEXT,
			metric_type TEXT,
			unit TEXT,
			scope_name TEXT,
			scope_version TEXT,
			schema_url TEXT,
			field_names TEXT[],
			field_types JSON,
			field_sources JSON,
			field_cardinality JSON,
			seen_count INTEGER,
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP
		)
	`

	db := r.pool.GetConnection()
	_, err := db.ExecContext(ctx, query)
	return err
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

func (r *SchemaRepository) GetSchema(ctx context.Context, schemaID string) (*repository.Schema, error) {
	query := `
		SELECT 
			c.schema_id, c.signal_type, c.signal_key, c.scope_name, c.scope_version,
			d.metric_type, d.unit, d.field_names, d.field_types,
			d.field_sources, d.field_cardinality, c.seen_count,
			c.created_at, c.updated_at
		FROM schema_core c
		JOIN schema_details d ON c.schema_id = d.schema_id
		WHERE c.schema_id = ?
	`

	db := r.pool.GetConnection()
	row := db.QueryRowContext(ctx, query, schemaID)

	var schema repository.Schema
	var fieldNamesJSON, fieldTypesJSON, fieldSourcesJSON, fieldCardinalityJSON []byte
	var metricType, unit sql.NullString

	err := row.Scan(
		&schema.SchemaID,
		&schema.SignalType,
		&schema.SignalKey,
		&schema.ScopeName,
		&schema.ScopeVersion,
		&metricType,
		&unit,
		&fieldNamesJSON,
		&fieldTypesJSON,
		&fieldSourcesJSON,
		&fieldCardinalityJSON,
		&schema.SeenCount,
		&schema.CreatedAt,
		&schema.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found: %s", schemaID)
		}
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	if metricType.Valid {
		schema.MetricType = &metricType.String
	}
	if unit.Valid {
		schema.Unit = &unit.String
	}

	if err := json.Unmarshal(fieldNamesJSON, &schema.FieldNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field names: %w", err)
	}

	if err := json.Unmarshal(fieldTypesJSON, &schema.FieldTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field types: %w", err)
	}

	if err := json.Unmarshal(fieldSourcesJSON, &schema.FieldSources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field sources: %w", err)
	}

	if err := json.Unmarshal(fieldCardinalityJSON, &schema.FieldCardinality); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field cardinality: %w", err)
	}

	return &schema, nil
}

func (r *SchemaRepository) GetSchemasByScope(ctx context.Context, scopeName, scopeVersion string) ([]*repository.Schema, error) {
	query := `
		SELECT 
			c.schema_id, c.signal_type, c.scope_name, c.scope_version,
			d.metric_type, d.unit, d.field_names, d.field_types,
			d.field_sources, d.field_cardinality, c.seen_count,
			c.created_at, c.updated_at
		FROM schema_core c
		JOIN schema_details d ON c.schema_id = d.schema_id
		WHERE c.scope_name = ? AND c.scope_version = ?
	`

	db := r.pool.GetConnection()
	rows, err := db.QueryContext(ctx, query, scopeName, scopeVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*repository.Schema
	for rows.Next() {
		var schema repository.Schema
		var fieldNamesJSON, fieldTypesJSON, fieldSourcesJSON, fieldCardinalityJSON []byte
		var metricType, unit sql.NullString

		err := rows.Scan(
			&schema.SchemaID,
			&schema.SignalType,
			&schema.ScopeName,
			&schema.ScopeVersion,
			&metricType,
			&unit,
			&fieldNamesJSON,
			&fieldTypesJSON,
			&fieldSourcesJSON,
			&fieldCardinalityJSON,
			&schema.SeenCount,
			&schema.CreatedAt,
			&schema.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}

		if metricType.Valid {
			schema.MetricType = &metricType.String
		}
		if unit.Valid {
			schema.Unit = &unit.String
		}

		if err := json.Unmarshal(fieldNamesJSON, &schema.FieldNames); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field names: %w", err)
		}

		if err := json.Unmarshal(fieldTypesJSON, &schema.FieldTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field types: %w", err)
		}

		if err := json.Unmarshal(fieldSourcesJSON, &schema.FieldSources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field sources: %w", err)
		}

		if err := json.Unmarshal(fieldCardinalityJSON, &schema.FieldCardinality); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field cardinality: %w", err)
		}

		schemas = append(schemas, &schema)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schemas: %w", err)
	}

	return schemas, nil
}

func (r *SchemaRepository) GetSchemasBySignalType(ctx context.Context, signalType string) ([]*repository.Schema, error) {
	query := `
		SELECT 
			c.schema_id, c.signal_type, c.scope_name, c.scope_version,
			d.metric_type, d.unit, d.field_names, d.field_types,
			d.field_sources, d.field_cardinality, c.seen_count,
			c.created_at, c.updated_at
		FROM schema_core c
		JOIN schema_details d ON c.schema_id = d.schema_id
		WHERE c.signal_type = ?
	`

	db := r.pool.GetConnection()
	rows, err := db.QueryContext(ctx, query, signalType)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*repository.Schema
	for rows.Next() {
		var schema repository.Schema
		var fieldNamesJSON, fieldTypesJSON, fieldSourcesJSON, fieldCardinalityJSON []byte
		var metricType, unit sql.NullString

		err := rows.Scan(
			&schema.SchemaID,
			&schema.SignalType,
			&schema.ScopeName,
			&schema.ScopeVersion,
			&metricType,
			&unit,
			&fieldNamesJSON,
			&fieldTypesJSON,
			&fieldSourcesJSON,
			&fieldCardinalityJSON,
			&schema.SeenCount,
			&schema.CreatedAt,
			&schema.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}

		if metricType.Valid {
			schema.MetricType = &metricType.String
		}
		if unit.Valid {
			schema.Unit = &unit.String
		}

		if err := json.Unmarshal(fieldNamesJSON, &schema.FieldNames); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field names: %w", err)
		}

		if err := json.Unmarshal(fieldTypesJSON, &schema.FieldTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field types: %w", err)
		}

		if err := json.Unmarshal(fieldSourcesJSON, &schema.FieldSources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field sources: %w", err)
		}

		if err := json.Unmarshal(fieldCardinalityJSON, &schema.FieldCardinality); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field cardinality: %w", err)
		}

		schemas = append(schemas, &schema)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schemas: %w", err)
	}

	return schemas, nil
}

func (r *SchemaRepository) UpdateSchemaSeenCount(ctx context.Context, schemaID string, count int) error {
	query := `
		UPDATE schema_core
		SET seen_count = seen_count + ?,
			updated_at = ?
		WHERE schema_id = ?
	`

	db := r.pool.GetConnection()
	_, err := db.ExecContext(ctx, query, count, time.Now(), schemaID)
	if err != nil {
		return fmt.Errorf("failed to update schema seen count: %w", err)
	}

	return nil
}

func (r *SchemaRepository) DeleteSchema(ctx context.Context, schemaID string) error {
	// Start a transaction
	tx, err := r.pool.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete from schema_details first (due to foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM schema_details WHERE schema_id = ?", schemaID)
	if err != nil {
		return fmt.Errorf("failed to delete schema details: %w", err)
	}

	// Delete from schema_core
	_, err = tx.ExecContext(ctx, "DELETE FROM schema_core WHERE schema_id = ?", schemaID)
	if err != nil {
		return fmt.Errorf("failed to delete schema core: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateFieldCardinality updates the cardinality information for a field
func (r *SchemaRepository) UpdateFieldCardinality(ctx context.Context, fieldName string, isHighCardinality bool) error {
	// Start a transaction
	tx, err := r.pool.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the field cardinality in schema_details
	query := `
		UPDATE schema_details 
		SET field_cardinality = json_set(field_cardinality, $1, $2),
		    updated_at = CURRENT_TIMESTAMP
		WHERE field_names ? $1
	`

	// Convert boolean to JSON string
	cardinalityJSON := fmt.Sprintf("%t", isHighCardinality)

	_, err = tx.ExecContext(ctx, query, fieldName, cardinalityJSON)
	if err != nil {
		return fmt.Errorf("failed to update field cardinality: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
