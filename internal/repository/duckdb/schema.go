package duckdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

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

	query := `
		INSERT INTO otel_schema_catalog (
			schema_id, signal_type, metric_type, unit,
			scope_name, scope_version, field_names, field_types,
			field_sources, field_cardinality, seen_count, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ` + fieldNamesArray + `, ?, ?, ?, ?, ?)
		ON CONFLICT (schema_id) DO UPDATE SET
			seen_count = otel_schema_catalog.seen_count + excluded.seen_count,
			updated_at = excluded.updated_at
		WHERE excluded.updated_at > otel_schema_catalog.updated_at;
	`

	db := r.pool.GetConnection()
	_, err = db.ExecContext(ctx, query,
		schema.SchemaID,
		schema.SignalType,
		schema.MetricType,
		schema.Unit,
		schema.ScopeName,
		schema.ScopeVersion,
		fieldTypesJSON,
		fieldSourcesJSON,
		fieldCardinalityJSON,
		schema.SeenCount,
		schema.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to register schema: %w", err)
	}

	return nil
}

func (r *SchemaRepository) GetSchema(ctx context.Context, schemaID string) (*repository.Schema, error) {
	query := `
		SELECT 
			schema_id, signal_type, metric_type, unit, service_name,
			scope_name, scope_version, field_names, field_types,
			field_sources, seen_count, created_at, updated_at
		FROM otel_schema_catalog
		WHERE schema_id = ?
	`

	db := r.pool.GetConnection()
	row := db.QueryRowContext(ctx, query, schemaID)

	var schema repository.Schema
	var fieldNamesJSON, fieldTypesJSON, fieldSourcesJSON []byte
	var metricType, unit sql.NullString

	err := row.Scan(
		&schema.SchemaID,
		&schema.SignalType,
		&metricType,
		&unit,
		&schema.ScopeName,
		&schema.ScopeVersion,
		&fieldNamesJSON,
		&fieldTypesJSON,
		&fieldSourcesJSON,
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

	return &schema, nil
}
