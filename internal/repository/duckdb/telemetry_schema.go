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
			profile_sample_aggregation_temporality, profile_sample_unit,
			note, protocol, seen_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			schema.ProfileSampleAggregationTemporality,
			schema.ProfileSampleUnit,
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

		// Insert entities
		for _, entity := range schema.Entities {
			// First insert the entity itself
			_, err = tx.ExecContext(ctx, `
				INSERT INTO telemetry_entities (entity_id, entity_type, first_seen, last_seen)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (entity_id) DO UPDATE SET
					last_seen = excluded.last_seen
				WHERE excluded.last_seen > telemetry_entities.last_seen
			`, entity.ID, entity.Type, entity.FirstSeen, entity.LastSeen)
			if err != nil {
				return fmt.Errorf("failed to insert entity: %w", err)
			}

			// Insert entity attributes
			for attrName, attrValue := range entity.Attributes {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO entity_attributes (entity_id, name, value, type)
					VALUES (?, ?, ?, ?)
				`, entity.ID, attrName, fmt.Sprintf("%v", attrValue), "string")
				if err != nil {
					return fmt.Errorf("failed to insert entity attribute: %w", err)
				}
			}

			// Link schema to entity
			_, err = tx.ExecContext(ctx, `
				INSERT INTO schema_entities (schema_id, entity_id)
				VALUES (?, ?)
				ON CONFLICT (schema_id, entity_id) DO NOTHING
			`, schema.SchemaID, entity.ID)
			if err != nil {
				return fmt.Errorf("failed to link schema to entity: %w", err)
			}
		}

		// Insert scope if present
		if schema.Scope != nil {
			scope := schema.Scope
			// First insert the scope itself
			_, err = tx.ExecContext(ctx, `
				INSERT INTO telemetry_scopes (scope_id, name, version, schema_url, first_seen, last_seen)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT (scope_id) DO UPDATE SET
					last_seen = excluded.last_seen
				WHERE excluded.last_seen > telemetry_scopes.last_seen
			`, scope.ID, scope.Name, scope.Version, scope.SchemaURL, scope.FirstSeen, scope.LastSeen)
			if err != nil {
				return fmt.Errorf("failed to insert scope: %w", err)
			}

			// Insert scope attributes
			for attrName, attrValue := range scope.Attributes {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO scope_attributes (scope_id, name, value, type)
					VALUES (?, ?, ?, ?)
				`, scope.ID, attrName, fmt.Sprintf("%v", attrValue), "string")
				if err != nil {
					return fmt.Errorf("failed to insert scope attribute: %w", err)
				}
			}

			// Link schema to scope
			_, err = tx.ExecContext(ctx, `
				INSERT INTO schema_scopes (schema_id, scope_id)
				VALUES (?, ?)
				ON CONFLICT (schema_id, scope_id) DO NOTHING
			`, schema.SchemaID, scope.ID)
			if err != nil {
				return fmt.Errorf("failed to link schema to scope: %w", err)
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
				-- Profile fields
				t.profile_sample_aggregation_temporality,
				t.profile_sample_unit,
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
			profile_sample_aggregation_temporality, profile_sample_unit,
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
			&schema.ProfileSampleAggregationTemporality,
			&schema.ProfileSampleUnit,
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

	// Populate entities for each telemetry schema
	for i := range schemas {
		// Get entities for this schema
		entityQuery := `
			SELECT te.entity_id, te.entity_type, te.first_seen, te.last_seen
			FROM telemetry_entities te
			INNER JOIN schema_entities se ON te.entity_id = se.entity_id
			WHERE se.schema_id = ?`

		entityRows, err := db.QueryContext(ctx, entityQuery, schemas[i].SchemaID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query schema entities: %w", err)
		}

		schemas[i].Entities = make(map[string]*schema.Entity)
		for entityRows.Next() {
			var entity schema.Entity
			if err := entityRows.Scan(
				&entity.ID,
				&entity.Type,
				&entity.FirstSeen,
				&entity.LastSeen,
			); err != nil {
				entityRows.Close()
				return nil, 0, fmt.Errorf("failed to scan entity row: %w", err)
			}

			// Get entity attributes
			attrQuery := `
				SELECT name, value, type
				FROM entity_attributes
				WHERE entity_id = ?`

			attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
			if err != nil {
				entityRows.Close()
				return nil, 0, fmt.Errorf("failed to query entity attributes: %w", err)
			}

			entity.Attributes = make(map[string]interface{})
			for attrRows.Next() {
				var name, value, attrType string
				if err := attrRows.Scan(&name, &value, &attrType); err != nil {
					attrRows.Close()
					entityRows.Close()
					return nil, 0, fmt.Errorf("failed to scan entity attribute: %w", err)
				}
				entity.Attributes[name] = value
			}
			attrRows.Close()

			schemas[i].Entities[entity.ID] = &entity
		}
		entityRows.Close()

		if err := entityRows.Err(); err != nil {
			return nil, 0, fmt.Errorf("error iterating entity rows: %w", err)
		}

		// Get scope for this schema
		scopeQuery := `
			SELECT ts.scope_id, ts.name, ts.version, ts.schema_url, ts.first_seen, ts.last_seen
			FROM telemetry_scopes ts
			INNER JOIN schema_scopes ss ON ts.scope_id = ss.scope_id
			WHERE ss.schema_id = ?`

		scopeRows, err := db.QueryContext(ctx, scopeQuery, schemas[i].SchemaID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query schema scope: %w", err)
		}

		// Each schema should have at most one scope
		if scopeRows.Next() {
			var scope schema.Scope
			if err := scopeRows.Scan(
				&scope.ID,
				&scope.Name,
				&scope.Version,
				&scope.SchemaURL,
				&scope.FirstSeen,
				&scope.LastSeen,
			); err != nil {
				scopeRows.Close()
				return nil, 0, fmt.Errorf("failed to scan scope row: %w", err)
			}

			// Get scope attributes
			scopeAttrQuery := `
				SELECT name, value, type
				FROM scope_attributes
				WHERE scope_id = ?`

			scopeAttrRows, err := db.QueryContext(ctx, scopeAttrQuery, scope.ID)
			if err != nil {
				scopeRows.Close()
				return nil, 0, fmt.Errorf("failed to query scope attributes: %w", err)
			}

			scope.Attributes = make(map[string]interface{})
			for scopeAttrRows.Next() {
				var name, value, attrType string
				if err := scopeAttrRows.Scan(&name, &value, &attrType); err != nil {
					scopeAttrRows.Close()
					scopeRows.Close()
					return nil, 0, fmt.Errorf("failed to scan scope attribute: %w", err)
				}
				scope.Attributes[name] = value
			}
			scopeAttrRows.Close()

			schemas[i].Scope = &scope
		}
		scopeRows.Close()

		if err := scopeRows.Err(); err != nil {
			return nil, 0, fmt.Errorf("error iterating scope rows: %w", err)
		}
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
						-- Profile fields
						t.profile_sample_aggregation_temporality,
						t.profile_sample_unit,
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
			-- Profile fields
			profile_sample_aggregation_temporality,
			profile_sample_unit,
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
		&s.ProfileSampleAggregationTemporality,
		&s.ProfileSampleUnit,
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

	// Get entities for this schema
	entityQuery := `
		SELECT te.entity_id, te.entity_type, te.first_seen, te.last_seen
		FROM telemetry_entities te
		INNER JOIN schema_entities se ON te.entity_id = se.entity_id
		INNER JOIN telemetry_schemas ts ON se.schema_id = ts.schema_id
		WHERE ts.schema_key = ?`

	rows, err = db.QueryContext(ctx, entityQuery, s.SchemaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema entities: %w", err)
	}
	defer rows.Close()

	s.Entities = make(map[string]*schema.Entity)
	for rows.Next() {
		var entity schema.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.Type,
			&entity.FirstSeen,
			&entity.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entity row: %w", err)
		}

		// Get entity attributes
		attrQuery := `
			SELECT name, value, type
			FROM entity_attributes
			WHERE entity_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query entity attributes: %w", err)
		}

		entity.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, fmt.Errorf("failed to scan entity attribute: %w", err)
			}
			entity.Attributes[name] = value
		}
		attrRows.Close()

		if _, ok := s.Entities[entity.ID]; !ok {
			s.Entities[entity.ID] = &entity
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity rows: %w", err)
	}

	// Get scope for this schema
	scopeQuery := `
		SELECT ts.scope_id, ts.name, ts.version, ts.schema_url, ts.first_seen, ts.last_seen
		FROM telemetry_scopes ts
		INNER JOIN schema_scopes ss ON ts.scope_id = ss.scope_id
		INNER JOIN telemetry_schemas tschema ON ss.schema_id = tschema.schema_id
		WHERE tschema.schema_key = ?`

	scopeRows, err := db.QueryContext(ctx, scopeQuery, s.SchemaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema scope: %w", err)
	}
	defer scopeRows.Close()

	// Each schema should have at most one scope
	if scopeRows.Next() {
		var scope schema.Scope
		if err := scopeRows.Scan(
			&scope.ID,
			&scope.Name,
			&scope.Version,
			&scope.SchemaURL,
			&scope.FirstSeen,
			&scope.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan scope row: %w", err)
		}

		// Get scope attributes
		scopeAttrQuery := `
			SELECT name, value, type
			FROM scope_attributes
			WHERE scope_id = ?`

		scopeAttrRows, err := db.QueryContext(ctx, scopeAttrQuery, scope.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query scope attributes: %w", err)
		}

		scope.Attributes = make(map[string]interface{})
		for scopeAttrRows.Next() {
			var name, value, attrType string
			if err := scopeAttrRows.Scan(&name, &value, &attrType); err != nil {
				scopeAttrRows.Close()
				return nil, fmt.Errorf("failed to scan scope attribute: %w", err)
			}
			scope.Attributes[name] = value
		}
		scopeAttrRows.Close()

		s.Scope = &scope
	}

	if err := scopeRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scope rows: %w", err)
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
			COUNT(DISTINCT se.entity_id) AS entity_count,
			MAX(te.last_seen) AS last_seen
		FROM telemetry_schemas t
		LEFT JOIN schema_versions sv ON t.schema_id = sv.schema_id
		LEFT JOIN schema_entities se ON t.schema_id = se.schema_id
		LEFT JOIN telemetry_entities te ON se.entity_id = te.entity_id
		WHERE 1=1` + where + `
		GROUP BY t.schema_id, sv.version
		ORDER BY MAX(te.last_seen) DESC NULLS LAST
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
		if err := rows.Scan(&row.SchemaId, &row.Version, &row.EntityCount, &lastSeen); err != nil {
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
			COUNT(DISTINCT se.entity_id) AS entity_count,
			MAX(te.last_seen) AS last_seen
		FROM telemetry_schemas t
		LEFT JOIN schema_versions sv ON t.schema_id = sv.schema_id
		LEFT JOIN schema_entities se ON t.schema_id = se.schema_id
		LEFT JOIN telemetry_entities te ON se.entity_id = te.entity_id
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
		&s.EntityCount,
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

	// Get entities for this schema
	entityQuery := `
		SELECT te.entity_id, te.entity_type, te.first_seen, te.last_seen
		FROM telemetry_entities te
		INNER JOIN schema_entities se ON te.entity_id = se.entity_id
		WHERE se.schema_id = ?`

	rows, err = db.QueryContext(ctx, entityQuery, schemaId)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema entities: %w", err)
	}
	defer rows.Close()

	s.Entities = make(map[string]*schema.Entity)
	for rows.Next() {
		var entity schema.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.Type,
			&entity.FirstSeen,
			&entity.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entity row: %w", err)
		}

		// Get entity attributes
		attrQuery := `
			SELECT name, value, type
			FROM entity_attributes
			WHERE entity_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query entity attributes: %w", err)
		}

		entity.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, fmt.Errorf("failed to scan entity attribute: %w", err)
			}
			entity.Attributes[name] = value
		}
		attrRows.Close()

		s.Entities[entity.ID] = &entity
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity rows: %w", err)
	}

	return &s, nil
}

func (r *TelemetrySchemaRepository) ListTelemetriesByEntity(ctx context.Context, entityType string) ([]schema.Telemetry, error) {
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
				-- Profile fields
				t.profile_sample_aggregation_temporality,
				t.profile_sample_unit,
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
			INNER JOIN schema_entities se ON t.schema_id = se.schema_id
			INNER JOIN telemetry_entities te ON se.entity_id = te.entity_id
			WHERE te.entity_type = ?
		)
		SELECT 
			schema_id, schema_version, schema_url, signal_type, schema_key,
			unit, metric_type, temporality, brief,
			log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
			span_kind, span_name, span_id, span_trace_id,
			profile_sample_aggregation_temporality, profile_sample_unit,
			note, protocol, seen_count,
			created_at, updated_at
		FROM latest_schemas
		WHERE rn = 1
		ORDER BY updated_at DESC`

	db := r.pool.GetConnection()

	rows, err := db.QueryContext(ctx, query, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetries by entity: %w", err)
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
			&t.ProfileSampleAggregationTemporality,
			&t.ProfileSampleUnit,
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

	// For each telemetry, get its attributes and entities
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

		telemetries[i].Attributes = attributes

		// Get entities
		entityQuery := `
			SELECT te.entity_id, te.entity_type, te.first_seen, te.last_seen
			FROM telemetry_entities te
			INNER JOIN schema_entities se ON te.entity_id = se.entity_id
			WHERE se.schema_id = ?`

		entityRows, err := db.QueryContext(ctx, entityQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema entities: %w", err)
		}

		telemetries[i].Entities = make(map[string]*schema.Entity)
		for entityRows.Next() {
			var entity schema.Entity
			if err := entityRows.Scan(
				&entity.ID,
				&entity.Type,
				&entity.FirstSeen,
				&entity.LastSeen,
			); err != nil {
				entityRows.Close()
				return nil, fmt.Errorf("failed to scan entity row: %w", err)
			}

			// Get entity attributes
			attrQuery := `
				SELECT name, value, type
				FROM entity_attributes
				WHERE entity_id = ?`

			attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
			if err != nil {
				entityRows.Close()
				return nil, fmt.Errorf("failed to query entity attributes: %w", err)
			}

			entity.Attributes = make(map[string]interface{})
			for attrRows.Next() {
				var name, value, attrType string
				if err := attrRows.Scan(&name, &value, &attrType); err != nil {
					attrRows.Close()
					entityRows.Close()
					return nil, fmt.Errorf("failed to scan entity attribute: %w", err)
				}
				entity.Attributes[name] = value
			}
			attrRows.Close()

			telemetries[i].Entities[entity.ID] = &entity
		}
		entityRows.Close()

		if err := entityRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating entity rows: %w", err)
		}

		// Get scope for this schema
		scopeQuery := `
			SELECT ts.scope_id, ts.name, ts.version, ts.schema_url, ts.first_seen, ts.last_seen
			FROM telemetry_scopes ts
			INNER JOIN schema_scopes ss ON ts.scope_id = ss.scope_id
			WHERE ss.schema_id = ?`

		scopeRows, err := db.QueryContext(ctx, scopeQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema scope: %w", err)
		}

		// Each schema should have at most one scope
		if scopeRows.Next() {
			var scope schema.Scope
			if err := scopeRows.Scan(
				&scope.ID,
				&scope.Name,
				&scope.Version,
				&scope.SchemaURL,
				&scope.FirstSeen,
				&scope.LastSeen,
			); err != nil {
				scopeRows.Close()
				return nil, fmt.Errorf("failed to scan scope row: %w", err)
			}

			// Get scope attributes
			scopeAttrQuery := `
				SELECT name, value, type
				FROM scope_attributes
				WHERE scope_id = ?`

			scopeAttrRows, err := db.QueryContext(ctx, scopeAttrQuery, scope.ID)
			if err != nil {
				scopeRows.Close()
				return nil, fmt.Errorf("failed to query scope attributes: %w", err)
			}

			scope.Attributes = make(map[string]interface{})
			for scopeAttrRows.Next() {
				var name, value, attrType string
				if err := scopeAttrRows.Scan(&name, &value, &attrType); err != nil {
					scopeAttrRows.Close()
					scopeRows.Close()
					return nil, fmt.Errorf("failed to scan scope attribute: %w", err)
				}
				scope.Attributes[name] = value
			}
			scopeAttrRows.Close()

			telemetries[i].Scope = &scope
		}
		scopeRows.Close()

		if err := scopeRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating scope rows: %w", err)
		}
	}

	return telemetries, nil
}

func (r *TelemetrySchemaRepository) ListTelemetriesByScope(ctx context.Context, scopeName string) ([]schema.Telemetry, error) {
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
				-- Profile fields
				t.profile_sample_aggregation_temporality,
				t.profile_sample_unit,
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
			INNER JOIN schema_scopes ss ON t.schema_id = ss.schema_id
			INNER JOIN telemetry_scopes ts ON ss.scope_id = ts.scope_id
			WHERE ts.name = ?
		)
		SELECT 
			schema_id, schema_version, schema_url, signal_type, schema_key,
			unit, metric_type, temporality, brief,
			log_severity_number, log_severity_text, log_body, log_flags, log_trace_id, log_span_id, log_event_name, log_dropped_attributes_count,
			span_kind, span_name, span_id, span_trace_id,
			profile_sample_aggregation_temporality, profile_sample_unit,
			note, protocol, seen_count,
			created_at, updated_at
		FROM latest_schemas
		WHERE rn = 1
		ORDER BY updated_at DESC`

	db := r.pool.GetConnection()

	rows, err := db.QueryContext(ctx, query, scopeName)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetries by scope: %w", err)
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
			&t.ProfileSampleAggregationTemporality,
			&t.ProfileSampleUnit,
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

	// For each telemetry, get its attributes and entities
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

		telemetries[i].Attributes = attributes

		// Get entities
		entityQuery := `
			SELECT te.entity_id, te.entity_type, te.first_seen, te.last_seen
			FROM telemetry_entities te
			INNER JOIN schema_entities se ON te.entity_id = se.entity_id
			WHERE se.schema_id = ?`

		entityRows, err := db.QueryContext(ctx, entityQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema entities: %w", err)
		}

		telemetries[i].Entities = make(map[string]*schema.Entity)
		for entityRows.Next() {
			var entity schema.Entity
			if err := entityRows.Scan(
				&entity.ID,
				&entity.Type,
				&entity.FirstSeen,
				&entity.LastSeen,
			); err != nil {
				entityRows.Close()
				return nil, fmt.Errorf("failed to scan entity row: %w", err)
			}

			// Get entity attributes
			attrQuery := `
				SELECT name, value, type
				FROM entity_attributes
				WHERE entity_id = ?`

			attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
			if err != nil {
				entityRows.Close()
				return nil, fmt.Errorf("failed to query entity attributes: %w", err)
			}

			entity.Attributes = make(map[string]interface{})
			for attrRows.Next() {
				var name, value, attrType string
				if err := attrRows.Scan(&name, &value, &attrType); err != nil {
					attrRows.Close()
					entityRows.Close()
					return nil, fmt.Errorf("failed to scan entity attribute: %w", err)
				}
				entity.Attributes[name] = value
			}
			attrRows.Close()

			telemetries[i].Entities[entity.ID] = &entity
		}
		entityRows.Close()

		if err := entityRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating entity rows: %w", err)
		}

		// Get scope for this schema
		scopeQuery := `
			SELECT ts.scope_id, ts.name, ts.version, ts.schema_url, ts.first_seen, ts.last_seen
			FROM telemetry_scopes ts
			INNER JOIN schema_scopes ss ON ts.scope_id = ss.scope_id
			WHERE ss.schema_id = ?`

		scopeRows, err := db.QueryContext(ctx, scopeQuery, telemetries[i].SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to query schema scope: %w", err)
		}

		// Each schema should have at most one scope
		if scopeRows.Next() {
			var scope schema.Scope
			if err := scopeRows.Scan(
				&scope.ID,
				&scope.Name,
				&scope.Version,
				&scope.SchemaURL,
				&scope.FirstSeen,
				&scope.LastSeen,
			); err != nil {
				scopeRows.Close()
				return nil, fmt.Errorf("failed to scan scope row: %w", err)
			}

			// Get scope attributes
			scopeAttrQuery := `
				SELECT name, value, type
				FROM scope_attributes
				WHERE scope_id = ?`

			scopeAttrRows, err := db.QueryContext(ctx, scopeAttrQuery, scope.ID)
			if err != nil {
				scopeRows.Close()
				return nil, fmt.Errorf("failed to query scope attributes: %w", err)
			}

			scope.Attributes = make(map[string]interface{})
			for scopeAttrRows.Next() {
				var name, value, attrType string
				if err := scopeAttrRows.Scan(&name, &value, &attrType); err != nil {
					scopeAttrRows.Close()
					scopeRows.Close()
					return nil, fmt.Errorf("failed to scan scope attribute: %w", err)
				}
				scope.Attributes[name] = value
			}
			scopeAttrRows.Close()

			telemetries[i].Scope = &scope
		}
		scopeRows.Close()

		if err := scopeRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating scope rows: %w", err)
		}
	}

	return telemetries, nil
}

func (r *TelemetrySchemaRepository) ListEntities(ctx context.Context, params query.ListQueryParams) ([]schema.Entity, int, error) {
	var args []any
	where := ""

	if params.Search != "" {
		where += " AND (te.entity_type LIKE ? OR te.entity_id LIKE ?)"
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	db := r.pool.GetConnection()

	countQuery := `
		SELECT COUNT(DISTINCT te.entity_id)
		FROM telemetry_entities te
		WHERE 1=1` + where

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	total := 0
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}

	if total == 0 {
		return []schema.Entity{}, 0, nil
	}

	query := `
		SELECT 
			te.entity_id,
			te.entity_type,
			te.first_seen,
			te.last_seen
		FROM telemetry_entities te
		WHERE 1=1` + where + `
		ORDER BY te.last_seen DESC
		LIMIT ? OFFSET ?`

	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	var entities []schema.Entity
	for rows.Next() {
		var entity schema.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.Type,
			&entity.FirstSeen,
			&entity.LastSeen,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan entity row: %w", err)
		}

		// Get entity attributes
		attrQuery := `
			SELECT name, value, type
			FROM entity_attributes
			WHERE entity_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query entity attributes: %w", err)
		}

		entity.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, 0, fmt.Errorf("failed to scan entity attribute: %w", err)
			}
			entity.Attributes[name] = value
		}
		attrRows.Close()

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating entity rows: %w", err)
	}

	return entities, total, nil
}

func (r *TelemetrySchemaRepository) ListEntitiesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Entity, error) {
	db := r.pool.GetConnection()

	query := `
		SELECT 
			entities.entity_id,
			entities.entity_type,
			entities.first_seen,
			entities.last_seen
		FROM telemetry_entities entities
		WHERE EXISTS (
			SELECT 1 FROM schema_entities se 
			INNER JOIN telemetry_schemas ts ON se.schema_id = ts.schema_id 
			WHERE se.entity_id = entities.entity_id 
			AND ts.schema_key = ?
		)
		ORDER BY entities.last_seen DESC`

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, telemetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities for telemetry: %w", err)
	}
	defer rows.Close()

	var entities []schema.Entity
	for rows.Next() {
		var entity schema.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.Type,
			&entity.FirstSeen,
			&entity.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan entity row: %w", err)
		}

		// Get entity attributes
		attrQuery := `
			SELECT name, value, type
			FROM entity_attributes
			WHERE entity_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, entity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query entity attributes: %w", err)
		}

		entity.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, fmt.Errorf("failed to scan entity attribute: %w", err)
			}
			entity.Attributes[name] = value
		}
		attrRows.Close()

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity rows: %w", err)
	}

	return entities, nil
}

func (r *TelemetrySchemaRepository) ListScopes(ctx context.Context, params query.ListQueryParams) ([]schema.Scope, int, error) {
	var args []any
	where := ""

	if params.Search != "" {
		where += " AND (ts.name LIKE ? OR ts.version LIKE ? OR ts.schema_url LIKE ?)"
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	db := r.pool.GetConnection()

	countQuery := `
		SELECT COUNT(DISTINCT ts.scope_id)
		FROM telemetry_scopes ts
		WHERE 1=1` + where

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	total := 0
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count scopes: %w", err)
	}

	if total == 0 {
		return []schema.Scope{}, 0, nil
	}

	query := `
		SELECT 
			ts.scope_id,
			ts.name,
			ts.version,
			ts.schema_url,
			ts.first_seen,
			ts.last_seen
		FROM telemetry_scopes ts
		WHERE 1=1` + where + `
		ORDER BY ts.last_seen DESC
		LIMIT ? OFFSET ?`

	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query scopes: %w", err)
	}
	defer rows.Close()

	var scopes []schema.Scope
	for rows.Next() {
		var scope schema.Scope
		if err := rows.Scan(
			&scope.ID,
			&scope.Name,
			&scope.Version,
			&scope.SchemaURL,
			&scope.FirstSeen,
			&scope.LastSeen,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan scope row: %w", err)
		}

		// Get scope attributes
		attrQuery := `
			SELECT name, value, type
			FROM scope_attributes
			WHERE scope_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, scope.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to query scope attributes: %w", err)
		}

		scope.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, 0, fmt.Errorf("failed to scan scope attribute: %w", err)
			}
			scope.Attributes[name] = value
		}
		attrRows.Close()

		scopes = append(scopes, scope)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating scope rows: %w", err)
	}

	return scopes, total, nil
}

func (r *TelemetrySchemaRepository) ListScopesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Scope, error) {
	db := r.pool.GetConnection()

	query := `
		SELECT 
			scopes.scope_id,
			scopes.name,
			scopes.version,
			scopes.schema_url,
			scopes.first_seen,
			scopes.last_seen
		FROM telemetry_scopes scopes
		WHERE EXISTS (
			SELECT 1 FROM schema_scopes ss 
			INNER JOIN telemetry_schemas ts ON ss.schema_id = ts.schema_id 
			WHERE ss.scope_id = scopes.scope_id 
			AND ts.schema_key = ?
		)
		ORDER BY scopes.last_seen DESC`

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, telemetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query scopes for telemetry: %w", err)
	}
	defer rows.Close()

	var scopes []schema.Scope
	for rows.Next() {
		var scope schema.Scope
		if err := rows.Scan(
			&scope.ID,
			&scope.Name,
			&scope.Version,
			&scope.SchemaURL,
			&scope.FirstSeen,
			&scope.LastSeen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan scope row: %w", err)
		}

		// Get scope attributes
		attrQuery := `
			SELECT name, value, type
			FROM scope_attributes
			WHERE scope_id = ?`

		attrRows, err := db.QueryContext(ctx, attrQuery, scope.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query scope attributes: %w", err)
		}

		scope.Attributes = make(map[string]interface{})
		for attrRows.Next() {
			var name, value, attrType string
			if err := attrRows.Scan(&name, &value, &attrType); err != nil {
				attrRows.Close()
				return nil, fmt.Errorf("failed to scan scope attribute: %w", err)
			}
			scope.Attributes[name] = value
		}
		attrRows.Close()

		scopes = append(scopes, scope)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scope rows: %w", err)
	}

	return scopes, nil
}

func (r *TelemetrySchemaRepository) Pool() *ConnectionPool {
	return r.pool
}
