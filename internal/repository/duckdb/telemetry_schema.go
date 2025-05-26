package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/tallycat/tallycat/internal/repository/query"
	"github.com/tallycat/tallycat/internal/schema"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TelemetrySchemaRepository struct {
	pool *ConnectionPool
}

func NewTelemetrySchemaRepository(pool *ConnectionPool) *TelemetrySchemaRepository {
	return &TelemetrySchemaRepository{
		pool: pool,
	}
}

func (r *TelemetrySchemaRepository) RegisterTelemetrySchemas(ctx context.Context, schemas []schema.Telemetry) error {
	tx, err := r.pool.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	schemaStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO telemetry_schemas (
			schema_id, schema_key, schema_version, schema_url, 
			signal_type, metric_type, temporality, unit, 
			brief, note, protocol, seen_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (schema_id) DO UPDATE SET
			seen_count = telemetry_schemas.seen_count + excluded.seen_count,
			updated_at = excluded.updated_at
		WHERE excluded.updated_at > telemetry_schemas.updated_at;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare schema insert statement: %w", err)
	}
	defer schemaStmt.Close()

	attrStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO schema_attributes (
			schema_id, name, type, source
		) VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare attribute insert statement: %w", err)
	}
	defer attrStmt.Close()

	producerStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO schema_producers (
			schema_id, producer_id, name, namespace, version, instance_id,
			first_seen, last_seen
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (schema_id, producer_id) DO UPDATE SET
			last_seen = excluded.last_seen
		WHERE excluded.last_seen > schema_producers.last_seen;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare producer insert statement: %w", err)
	}
	defer producerStmt.Close()

	for _, schema := range schemas {
		_, err = schemaStmt.ExecContext(ctx,
			schema.SchemaID,
			schema.SchemaKey,
			schema.SchemaVersion,
			schema.SchemaURL,
			schema.TelemetryType,
			schema.MetricType,
			schema.MetricTemporality,
			schema.MetricUnit,
			schema.Brief,
			schema.Note,
			schema.Protocol,
			schema.SeenCount,
			schema.CreatedAt,
			schema.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert schema: %w", err)
		}

		for _, attr := range schema.Attributes {
			_, err = attrStmt.ExecContext(ctx,
				schema.SchemaID,
				attr.Name,
				attr.Type,
				attr.Source,
			)
			if err != nil {
				return fmt.Errorf("failed to insert attribute for schema %v: %w", schema.SchemaID, err)
			}
		}

		// Insert producers
		for _, producer := range schema.Producers {
			_, err = producerStmt.ExecContext(ctx,
				schema.SchemaID,
				producer.ProducerID(),
				producer.Name,
				producer.Namespace,
				producer.Version,
				producer.InstanceID,
				producer.FirstSeen,
				producer.LastSeen,
			)
			if err != nil {
				return fmt.Errorf("failed to insert producer for schema %v: %w", schema.SchemaID, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Debug(
		"successfully registered telemetry schemas",
		"schema_count", len(schemas),
		"attribute_count", len(schemas)*len(schemas[0].Attributes),
	)
	return nil
}

func (r *TelemetrySchemaRepository) ListSchemas(ctx context.Context, params query.ListQueryParams) ([]schema.Telemetry, int, error) {
	var args []any
	where := ""

	if params.FilterType != "" && params.FilterType != "all" {
		where += " AND t.signal_type = ?"
		args = append(args, cases.Title(language.English).String(params.FilterType))
	}

	if params.Search != "" {
		where += " AND (t.schema_id LIKE ? OR t.schema_key LIKE ? OR t.metric_type LIKE ? OR t.unit LIKE ?)"
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	countQuery := `
		SELECT COUNT(DISTINCT (t.signal_type, t.schema_key))
		FROM telemetry_schemas t
		WHERE 1=1` + where

	db := r.pool.GetConnection()

	// TODO: Allow this to be configurable
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	total := 0
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count schemas: %w", err)
	}

	if total == 0 {
		return []schema.Telemetry{}, 0, nil
	}

	query := `
		WITH latest_schemas AS (
			SELECT 
				t.schema_id,
				t.schema_version,
				t.schema_url,
				t.signal_type,
				t.schema_key,
				t.unit,
				t.metric_type,
				t.temporality,
				t.brief,
				t.note,
				t.protocol,
				t.seen_count,
				t.created_at,
				t.updated_at,
				COUNT(*) OVER (PARTITION BY t.signal_type, t.schema_key) as version_count,
				ROW_NUMBER() OVER (
					PARTITION BY t.signal_type, t.schema_key 
					ORDER BY t.updated_at DESC
				) as rn
			FROM telemetry_schemas t
			WHERE 1=1` + where + `
		)
		SELECT 
			schema_id, schema_version, schema_url, signal_type,
			schema_key, unit, metric_type, temporality,
			brief, note, protocol, seen_count,
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

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	// Scan results
	var schemas []schema.Telemetry
	for rows.Next() {
		var schema schema.Telemetry
		var versionCount int

		if err := rows.Scan(
			&schema.SchemaID,
			&schema.SchemaVersion,
			&schema.SchemaURL,
			&schema.TelemetryType,
			&schema.SchemaKey,
			&schema.MetricUnit,
			&schema.MetricType,
			&schema.MetricTemporality,
			&schema.Brief,
			&schema.Note,
			&schema.Protocol,
			&schema.SeenCount,
			&schema.CreatedAt,
			&schema.UpdatedAt,
			&versionCount,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan schema row: %w", err)
		}

		schemas = append(schemas, schema)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, total, nil
}

func (r *TelemetrySchemaRepository) GetSchemaByKey(ctx context.Context, schemaKey string) (*schema.Telemetry, error) {
	queryStr := `
		WITH latest_schema
			AS (SELECT 	t.schema_id,
						t.schema_version,
						t.schema_url,
						t.signal_type,
						t.schema_key,
						t.unit,
						t.metric_type,
						t.temporality,
						t.brief,
						t.note,
						t.protocol,
						t.seen_count,
						t.created_at,
						t.updated_at,
						Count(*)
						OVER (
							partition BY t.signal_type, t.schema_key) AS version_count,
						Row_number()
						OVER (
							partition BY t.signal_type, t.schema_key
							ORDER BY t.updated_at DESC )              AS rn
				FROM   telemetry_schemas t
				WHERE  t.schema_key = ?)
		SELECT schema_id,
			schema_version,
			schema_url,
			signal_type,
			schema_key,
			unit,
			metric_type,
			temporality,
			brief,
			note,
			protocol,
			seen_count,
			created_at,
			updated_at,
			version_count
		FROM   latest_schema
		WHERE  rn = 1 `

	db := r.pool.GetConnection()

	// Use context timeout for query
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var s schema.Telemetry
	var versionCount int

	err := db.QueryRowContext(ctx, queryStr, schemaKey).Scan(
		&s.SchemaID,
		&s.SchemaVersion,
		&s.SchemaURL,
		&s.TelemetryType,
		&s.SchemaKey,
		&s.MetricUnit,
		&s.MetricType,
		&s.MetricTemporality,
		&s.Brief,
		&s.Note,
		&s.Protocol,
		&s.SeenCount,
		&s.CreatedAt,
		&s.UpdatedAt,
		&versionCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query schema: %w", err)
	}

	// Get attributes for this schema
	attrQuery := `
		SELECT DISTINCT name, type, source
		FROM schema_attributes
		WHERE schema_id = ?
		ORDER BY name`

	rows, err := db.QueryContext(ctx, attrQuery, s.SchemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema attributes: %w", err)
	}
	defer rows.Close()

	var attributes []schema.Attribute
	for rows.Next() {
		var attr schema.Attribute
		if err := rows.Scan(&attr.Name, &attr.Type, &attr.Source); err != nil {
			return nil, fmt.Errorf("failed to scan attribute row: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attribute rows: %w", err)
	}

	s.Attributes = attributes

	// Get producers for this schema
	producerQuery := `
		SELECT producer_id, name, namespace, version, instance_id,
			   first_seen, last_seen
		FROM schema_producers
		INNER JOIN telemetry_schemas ON schema_producers.schema_id = telemetry_schemas.schema_id
		WHERE schema_key = ?`

	rows, err = db.QueryContext(ctx, producerQuery, s.SchemaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema producers: %w", err)
	}
	defer rows.Close()

	s.Producers = make(map[string]*schema.Producer)
	for rows.Next() {
		var producer schema.Producer
		var producerID string
		if err := rows.Scan(
			&producerID,
			&producer.Name,
			&producer.Namespace,
			&producer.Version,
			&producer.InstanceID,
			&producer.FirstSeen,
			&producer.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan producer row: %w", err)
		}

		if _, ok := s.Producers[producerID]; !ok {
			s.Producers[producerID] = &producer
		} else {
			slog.Warn("producer already exists", "producer_id", producerID)
		}

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating producer rows: %w", err)
	}

	return &s, nil
}
