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
			schema_id, schema_key, schema_version, schema_url, signal_type, 
			metric_type, temporality, unit, brief, 
			log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
			span_kind, span_name, span_id, span_trace_id,
			note, protocol, seen_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			schema.LogSeverityNumber,
			schema.LogSeverityText,
			schema.LogBody,
			schema.LogFlags,
			schema.LogTraceID,
			schema.LogSpanID,
			schema.LogEventName,
			schema.LogDroppedAttributesCount,
			schema.SpanKind,
			schema.SpanName,
			schema.SpanID,
			schema.SpanTraceID,
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

	attributeCount := 0
	if len(schemas) > 0 {
		for _, schema := range schemas {
			attributeCount += len(schema.Attributes)
		}
	}

	slog.Debug(
		"successfully registered telemetry schemas",
		"schema_count", len(schemas),
		"attribute_count", attributeCount,
	)
	return nil
}

func (r *TelemetrySchemaRepository) ListTelemetries(ctx context.Context, params query.ListQueryParams) ([]schema.Telemetry, int, error) {
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
				-- Metric fields
				t.unit,
				t.metric_type,
				t.temporality,
				t.brief,
				-- Log fields
				t.log_severity_number,
				t.log_severity_text,
				t.log_body,
				t.log_flags,
				t.log_trace_id,
				t.log_span_id,
				t.log_event_name,
				t.log_dropped_attributes_count,
				-- Span fields
				t.span_kind,
				t.span_name,
				t.span_id,
				t.span_trace_id,
				-- Common fields
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
			schema_id, schema_version, schema_url, signal_type, schema_key, 
			unit, metric_type, temporality, brief,
			log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
			span_kind, span_name, span_id, span_trace_id,
			note, protocol, seen_count,
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
			&schema.LogSeverityNumber,
			&schema.LogSeverityText,
			&schema.LogBody,
			&schema.LogFlags,
			&schema.LogTraceID,
			&schema.LogSpanID,
			&schema.LogEventName,
			&schema.LogDroppedAttributesCount,
			&schema.SpanKind,
			&schema.SpanName,
			&schema.SpanID,
			&schema.SpanTraceID,
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

func (r *TelemetrySchemaRepository) GetTelemetry(ctx context.Context, schemaKey string) (*schema.Telemetry, error) {
	queryStr := `
		WITH latest_schema
			AS (SELECT 	t.schema_id,
						t.schema_version,
						t.schema_url,
						t.signal_type,
						t.schema_key,
						-- Metric fields
						t.unit,
						t.metric_type,
						t.temporality,
						t.brief,
						-- Log fields
						t.log_severity_number,
						t.log_severity_text,
						t.log_body,
						t.log_flags,
						t.log_trace_id,
						t.log_span_id,
						t.log_event_name,
						t.log_dropped_attributes_count,
						-- Span fields
						t.span_kind,
						t.span_name,
						t.span_id,
						t.span_trace_id,
						-- Common fields
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
			-- Metric fields
			unit,
			metric_type,
			temporality,
			brief,
			-- Log fields
			log_severity_number,
			log_severity_text,
			log_body,
			log_flags,
			log_trace_id,
			log_span_id,
			log_event_name,
			log_dropped_attributes_count,
			-- Span fields
			span_kind,
			span_name,
			span_id,
			span_trace_id,
			-- Common fields
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
		&s.LogSeverityNumber,
		&s.LogSeverityText,
		&s.LogBody,
		&s.LogFlags,
		&s.LogTraceID,
		&s.LogSpanID,
		&s.LogEventName,
		&s.LogDroppedAttributesCount,
		&s.SpanKind,
		&s.SpanName,
		&s.SpanID,
		&s.SpanTraceID,
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

func (r *TelemetrySchemaRepository) AssignTelemetrySchemaVersion(ctx context.Context, assgiment schema.SchemaAssignment) error {
	tx, err := r.pool.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO schema_versions (schema_id, version, reason, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (schema_id) DO UPDATE SET
			version = excluded.version,
			reason = excluded.reason,
			updated_at = excluded.updated_at
		WHERE excluded.updated_at > schema_versions.updated_at;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		assgiment.SchemaId,
		assgiment.Version,
		assgiment.Reason,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TelemetrySchemaRepository) ListTelemetrySchemas(ctx context.Context, schemaKey string, params query.ListQueryParams) ([]schema.TelemetrySchema, int, error) {
	var args []any
	where := " AND t.schema_key = ?"
	args = append(args, schemaKey)

	if params.FilterType != "" && params.FilterType != "all" {
		where += " AND t.signal_type = ?"
		args = append(args, cases.Title(language.English).String(params.FilterType))
	}

	if params.Search != "" {
		where += " AND (t.schema_id LIKE ? OR t.schema_key LIKE ? OR t.metric_type LIKE ? OR t.unit LIKE ?)"
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	db := r.pool.GetConnection()

	countQuery := `
		SELECT COUNT(*)
		FROM telemetry_schemas t
		WHERE 1=1` + where

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	total := 0
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count schema assignments: %w", err)
	}

	if total == 0 {
		return []schema.TelemetrySchema{}, 0, nil
	}

	query := `
		SELECT
			t.schema_id,
			COALESCE(sv.version, 'Unassigned') AS version,
			COUNT(DISTINCT sp.producer_id) AS producer_count,
			MAX(sp.last_seen) AS last_seen
		FROM telemetry_schemas t
		LEFT JOIN schema_versions sv ON t.schema_id = sv.schema_id
		LEFT JOIN schema_producers sp ON t.schema_id = sp.schema_id
		WHERE 1=1` + where + `
		GROUP BY t.schema_id, sv.version
		ORDER BY MAX(sp.last_seen) DESC NULLS LAST
		LIMIT ? OFFSET ?`

	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query schema assignments: %w", err)
	}
	defer rows.Close()

	var assignments []schema.TelemetrySchema
	for rows.Next() {
		var row schema.TelemetrySchema
		var lastSeen sql.NullTime
		if err := rows.Scan(&row.SchemaId, &row.Version, &row.ProducerCount, &lastSeen); err != nil {
			return nil, 0, fmt.Errorf("failed to scan schema assignment row: %w", err)
		}
		if lastSeen.Valid {
			row.LastSeen = &lastSeen.Time
		} else {
			row.LastSeen = nil
		}
		assignments = append(assignments, row)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating schema assignment rows: %w", err)
	}

	return assignments, total, nil
}

func (r *TelemetrySchemaRepository) GetTelemetrySchema(ctx context.Context, schemaId string) (*schema.TelemetrySchema, error) {
	query := `
		SELECT
			t.schema_id,
			COALESCE(sv.version, 'Unassigned') AS version,
			COUNT(DISTINCT sp.producer_id) AS producer_count,
			MAX(sp.last_seen) AS last_seen
		FROM telemetry_schemas t
		LEFT JOIN schema_versions sv ON t.schema_id = sv.schema_id
		LEFT JOIN schema_producers sp ON t.schema_id = sp.schema_id
		WHERE t.schema_id = ?
		GROUP BY t.schema_id, sv.version`

	db := r.pool.GetConnection()

	// Use context timeout for query
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var s schema.TelemetrySchema
	var lastSeen sql.NullTime

	err := db.QueryRowContext(ctx, query, schemaId).Scan(
		&s.SchemaId,
		&s.Version,
		&s.ProducerCount,
		&lastSeen,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query schema: %w", err)
	}

	if lastSeen.Valid {
		s.LastSeen = &lastSeen.Time
	}

	// Get attributes for this schema
	attrQuery := `
		SELECT DISTINCT name, type, source
		FROM schema_attributes
		WHERE schema_id = ?
		ORDER BY name`

	rows, err := db.QueryContext(ctx, attrQuery, schemaId)
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
		WHERE schema_id = ?`

	rows, err = db.QueryContext(ctx, producerQuery, schemaId)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema producers: %w", err)
	}
	defer rows.Close()

	var producers []schema.Producer
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
		producers = append(producers, producer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating producer rows: %w", err)
	}

	s.Producers = producers

	return &s, nil
}

func (r *TelemetrySchemaRepository) ListTelemetriesByProducer(ctx context.Context, producerName, producerVersion string) ([]schema.Telemetry, error) {
	// Handle empty version by checking for NULL or empty string in database
	var query string
	var args []interface{}

	if producerVersion == "" {
		query = `
			WITH latest_schemas AS (
				SELECT 
					t.schema_id,
					t.schema_version,
					t.schema_url,
					t.signal_type,
					t.schema_key,
					-- Metric fields
					t.unit,
					t.metric_type,
					t.temporality,
					t.brief,
					-- Log fields
					t.log_severity_number,
					t.log_severity_text,
					t.log_body,
					t.log_flags,
					t.log_trace_id,
					t.log_span_id,
					t.log_event_name,
					t.log_dropped_attributes_count,
					-- Span fields
					t.span_kind,
					t.span_name,
					t.span_id,
					t.span_trace_id,
					-- Common fields
					t.note,
					t.protocol,
					t.seen_count,
					t.created_at,
					t.updated_at,
					ROW_NUMBER() OVER (
						PARTITION BY t.signal_type, t.schema_key 
						ORDER BY t.updated_at DESC
					) as rn
				FROM telemetry_schemas t
				INNER JOIN schema_producers sp ON t.schema_id = sp.schema_id
				WHERE sp.name = ? AND (sp.version IS NULL OR sp.version = '')
			)
			SELECT 
				schema_id, schema_version, schema_url, signal_type, schema_key,
				unit, metric_type, temporality, brief,
				log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
				span_kind, span_name, span_id, span_trace_id,
				note, protocol, seen_count,
				created_at, updated_at
			FROM latest_schemas
			WHERE rn = 1
			ORDER BY updated_at DESC`
		args = []interface{}{producerName}
	} else {
		query = `
			WITH latest_schemas AS (
				SELECT 
					t.schema_id,
					t.schema_version,
					t.schema_url,
					t.signal_type,
					t.schema_key,
					-- Metric fields
					t.unit,
					t.metric_type,
					t.temporality,
					t.brief,
					-- Log fields
					t.log_severity_number,
					t.log_severity_text,
					t.log_body,
					t.log_flags,
					t.log_trace_id,
					t.log_span_id,
					t.log_event_name,
					t.log_dropped_attributes_count,
					-- Span fields
					t.span_kind,
					t.span_name,
					t.span_id,
					t.span_trace_id,
					-- Common fields
					t.note,
					t.protocol,
					t.seen_count,
					t.created_at,
					t.updated_at,
					ROW_NUMBER() OVER (
						PARTITION BY t.signal_type, t.schema_key 
						ORDER BY t.updated_at DESC
					) as rn
				FROM telemetry_schemas t
				INNER JOIN schema_producers sp ON t.schema_id = sp.schema_id
				WHERE sp.name = ? AND sp.version = ?
			)
			SELECT 
				schema_id, schema_version, schema_url, signal_type, schema_key,
				unit, metric_type, temporality, brief,
				log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
				span_kind, span_name, span_id, span_trace_id,
				note, protocol, seen_count,
				created_at, updated_at
			FROM latest_schemas
			WHERE rn = 1
			ORDER BY updated_at DESC`
		args = []interface{}{producerName, producerVersion}
	}

	db := r.pool.GetConnection()

	// Use context timeout for query
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetries by producer: %w", err)
	}
	defer rows.Close()

	var telemetries []schema.Telemetry
	for rows.Next() {
		var t schema.Telemetry

		err := rows.Scan(
			&t.SchemaID,
			&t.SchemaVersion,
			&t.SchemaURL,
			&t.TelemetryType,
			&t.SchemaKey,
			&t.MetricUnit,
			&t.MetricType,
			&t.MetricTemporality,
			&t.Brief,
			&t.LogSeverityNumber,
			&t.LogSeverityText,
			&t.LogBody,
			&t.LogFlags,
			&t.LogTraceID,
			&t.LogSpanID,
			&t.LogEventName,
			&t.LogDroppedAttributesCount,
			&t.SpanKind,
			&t.SpanName,
			&t.SpanID,
			&t.SpanTraceID,
			&t.Note,
			&t.Protocol,
			&t.SeenCount,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry row: %w", err)
		}

		telemetries = append(telemetries, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	// For each telemetry, get its attributes and producers
	for i := range telemetries {
		// Get attributes
		attrQuery := `
			SELECT DISTINCT name, type, source
			FROM schema_attributes
			WHERE schema_id = ?
			ORDER BY name`

		attrRows, err := db.QueryContext(ctx, attrQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema attributes: %w", err)
		}

		var attributes []schema.Attribute
		for attrRows.Next() {
			var attr schema.Attribute
			if err := attrRows.Scan(&attr.Name, &attr.Type, &attr.Source); err != nil {
				attrRows.Close()
				return nil, fmt.Errorf("failed to scan attribute row: %w", err)
			}
			attributes = append(attributes, attr)
		}
		attrRows.Close()

		if err := attrRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating attribute rows: %w", err)
		}

		telemetries[i].Attributes = attributes

		// Get producers
		producerQuery := `
			SELECT producer_id, name, namespace, version, instance_id,
				   first_seen, last_seen
			FROM schema_producers
			WHERE schema_id = ?`

		prodRows, err := db.QueryContext(ctx, producerQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema producers: %w", err)
		}

		telemetries[i].Producers = make(map[string]*schema.Producer)
		for prodRows.Next() {
			var producer schema.Producer
			var producerID string
			if err := prodRows.Scan(
				&producerID,
				&producer.Name,
				&producer.Namespace,
				&producer.Version,
				&producer.InstanceID,
				&producer.FirstSeen,
				&producer.LastSeen,
			); err != nil {
				prodRows.Close()
				return nil, fmt.Errorf("failed to scan producer row: %w", err)
			}

			telemetries[i].Producers[producerID] = &producer
		}
		prodRows.Close()

		if err := prodRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating producer rows: %w", err)
		}
	}

	return telemetries, nil
}

func (r *TelemetrySchemaRepository) Pool() *ConnectionPool {
	return r.pool
}
