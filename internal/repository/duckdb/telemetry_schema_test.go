package duckdb

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
	"github.com/tallycat/tallycat/internal/repository/query"
	"github.com/tallycat/tallycat/internal/schema"
)

func setupTestDB(t *testing.T) *TelemetrySchemaRepository {
	// Create in-memory database
	db, err := sql.Open("duckdb", ":memory:")
	require.NoError(t, err)

	// Create connection pool directly with the database
	pool := &ConnectionPool{
		db:     db,
		config: &Config{},
		logger: slog.Default(),
	}

	// Create tables manually to avoid import cycle
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS telemetry_schemas (
			schema_id TEXT PRIMARY KEY,
			schema_key TEXT,
			schema_version TEXT,
			schema_url TEXT,
			signal_type TEXT,
			-- Metric fields
			metric_type TEXT,
			temporality TEXT,
			unit TEXT,
			brief TEXT,
			-- Log fields
			log_severity_number INTEGER,
			log_severity_text TEXT,
			log_body TEXT,
			log_flags INTEGER,
			log_trace_id TEXT,
			log_span_id TEXT,
			log_event_name TEXT,
			log_dropped_attributes_count INTEGER,
			-- Span fields
			span_kind TEXT,
			span_name TEXT,
			span_id TEXT,
			span_trace_id TEXT,
			-- Profile fields
			profile_sample_aggregation_temporality TEXT,
			profile_sample_unit TEXT,
			-- Common fields
			note TEXT,
			protocol TEXT,
			seen_count INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS schema_attributes (
			schema_id TEXT,
			name TEXT,
			type TEXT,
			source TEXT,
			FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id)
		);

		-- Create telemetry_entities table
		CREATE TABLE IF NOT EXISTS telemetry_entities (
			entity_id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL
		);

		-- Create entity_attributes table
		CREATE TABLE IF NOT EXISTS entity_attributes (
			entity_id TEXT,
			name TEXT,
			value TEXT,
			type TEXT,
			FOREIGN KEY (entity_id) REFERENCES telemetry_entities(entity_id)
		);

		-- Create schema_entities table (many-to-many relationship)
		CREATE TABLE IF NOT EXISTS schema_entities (
			schema_id TEXT,
			entity_id TEXT,
			FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id),
			FOREIGN KEY (entity_id) REFERENCES telemetry_entities(entity_id),
			PRIMARY KEY (schema_id, entity_id)
		);

		-- Create telemetry_scopes table
		CREATE TABLE IF NOT EXISTS telemetry_scopes (
			scope_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			version TEXT,
			schema_url TEXT,
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL
		);

		-- Create scope_attributes table
		CREATE TABLE IF NOT EXISTS scope_attributes (
			scope_id TEXT,
			name TEXT,
			value TEXT,
			type TEXT,
			FOREIGN KEY (scope_id) REFERENCES telemetry_scopes(scope_id)
		);

		-- Create schema_scopes table (many-to-many relationship)
		CREATE TABLE IF NOT EXISTS schema_scopes (
			schema_id TEXT,
			scope_id TEXT,
			FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id),
			FOREIGN KEY (scope_id) REFERENCES telemetry_scopes(scope_id),
			PRIMARY KEY (schema_id, scope_id)
		);
	`)
	require.NoError(t, err)

	return &TelemetrySchemaRepository{
		pool: pool,
	}
}

func TestListTelemetriesByEntity_EntityNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByEntity_EntityWithMetrics(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			Brief:         "HTTP server request duration",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.server.requests",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeSum,
			MetricUnit:    "1",
			Brief:         "HTTP server request count",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 5,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity2": {
					ID:   "entity2",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "1.0.0",
						"service.namespace": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "metric3_schema_id",
			SchemaKey:     "cpu.usage",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "%",
			Brief:         "CPU usage percentage",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity3": {
					ID:   "entity3",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "my-pod",
						"k8s.namespace.name": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get metrics for service entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right metrics for service entity
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "http.server.duration")
	require.Contains(t, schemaKeys, "http.server.requests")
	require.NotContains(t, schemaKeys, "cpu.usage")
}

func TestListTelemetriesByEntity_EntityWithNoMetrics(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert a producer with different name/version
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "test.metric",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "1",
			Brief:         "Test metric",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "2.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Look for a entity that has no metrics
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "k8s")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByEntity_EntityWithLogRecords(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "log1_schema_id",
			SchemaKey:     "Reading CPU usage",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         1,
			LogSeverityText:           "INFO",
			LogBody:                   "Reading CPU usage",
			LogFlags:                  0,
			LogTraceID:                "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:                 "1234567890ABCDEF",
			LogEventName:              "Reading CPU usage",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "log2_schema_id",
			SchemaKey:     "Reading Total Memory",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         1,
			LogSeverityText:           "INFO",
			LogBody:                   "Reading Total Memory",
			LogFlags:                  0,
			LogTraceID:                "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:                 "1234567890ABCDEF",
			LogEventName:              "Reading Total Memory",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 5,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity2": {
					ID:   "entity2",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "2.0.0",
						"service.namespace": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "log3_schema_id",
			SchemaKey:     "HTTP server requested",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         1,
			LogSeverityText:           "INFO",
			LogBody:                   "HTTP server requested",
			LogFlags:                  0,
			LogTraceID:                "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:                 "1234567890ABCDEF",
			LogEventName:              "HTTP server requested",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity3": {
					ID:   "entity3",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "other-pod",
						"k8s.namespace.name": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get logs for service entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right logs
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "Reading CPU usage")
	require.Contains(t, schemaKeys, "Reading Total Memory")
	require.NotContains(t, schemaKeys, "HTTP server requested")
}

func TestListTelemetriesByEntity_EntityWithNoLogRecords(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "log1_schema_id",
			SchemaKey:     "Reading CPU usage",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         1,
			LogSeverityText:           "INFO",
			LogBody:                   "Reading CPU usage",
			LogFlags:                  0,
			LogTraceID:                "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:                 "1234567890ABCDEF",
			LogEventName:              "Reading CPU usage",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get logs for service entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "k8s")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByEntity_EntityWithTraces(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "trace1_schema_id",
			SchemaKey:     "database-operation",
			TelemetryType: schema.TelemetryTypeSpan,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "Server",
			SpanName:    "database-operation",
			SpanID:      "1234567890ABCDEF",
			SpanTraceID: "1234567890ABCDEF1234567890ABCDEF",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "trace2_schema_id",
			SchemaKey:     "http-request",
			TelemetryType: schema.TelemetryTypeSpan,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "Client",
			SpanName:    "http-request",
			SpanID:      "1234567890ABCDEF",
			SpanTraceID: "1234567890ABCDEF1234567890ABCDEF",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 5,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity2": {
					ID:   "entity2",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "2.0.0",
						"service.namespace": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "trace3_schema_id",
			SchemaKey:     "Random operation",
			TelemetryType: schema.TelemetryTypeSpan,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "Server",
			SpanName:    "Random operation",
			SpanID:      "1234567890ABCDEF",
			SpanTraceID: "1234567890ABCDEF1234567890ABCDEF",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity3": {
					ID:   "entity3",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "other-pod",
						"k8s.namespace.name": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get spans for service entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right spans for service entity
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "database-operation")
	require.Contains(t, schemaKeys, "http-request")
	require.NotContains(t, schemaKeys, "Random operation")
}

func TestListTelemetriesByEntity_EntityWithNoTraces(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "trace1_schema_id",
			SchemaKey:     "database-operation",
			TelemetryType: schema.TelemetryTypeSpan,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "Server",
			SpanName:    "database-operation",
			SpanID:      "1234567890ABCDEF",
			SpanTraceID: "1234567890ABCDEF1234567890ABCDEF",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get telemetries for nonexistent entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "k8s")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByEntity_EntityWithProfiles(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "profile1_schema_id",
			SchemaKey:     "cpu.usage",
			TelemetryType: schema.TelemetryTypeProfile,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "Delta",
			ProfileSampleUnit:                   "ms",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "profile2_schema_id",
			SchemaKey:     "memory.usage",
			TelemetryType: schema.TelemetryTypeProfile,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "Delta",
			ProfileSampleUnit:                   "bytes",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 5,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity2": {
					ID:   "entity2",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "2.0.0",
						"service.namespace": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
		{
			SchemaID:      "profile3_schema_id",
			SchemaKey:     "objects_alloc",
			TelemetryType: schema.TelemetryTypeProfile,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "Delta",
			ProfileSampleUnit:                   "objects",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity3": {
					ID:   "entity3",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "other-pod",
						"k8s.namespace.name": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get profiles for service entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right profiles
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "cpu.usage")
	require.Contains(t, schemaKeys, "memory.usage")
	require.NotContains(t, schemaKeys, "objects_alloc")
}

func TestListTelemetriesByEntity_EntityWithNoProfiles(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "profile1_schema_id",
			SchemaKey:     "cpu.usage",
			TelemetryType: schema.TelemetryTypeProfile,
			// Metric fields
			MetricType: schema.MetricTypeEmpty,
			MetricUnit: "",
			Brief:      "",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "Delta",
			ProfileSampleUnit:                   "ms",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get telemetries for nonexistent entity
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "k8s")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByEntity_LatestSchemaVersionOnly(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()

	// Insert multiple versions of the same metric from the same producer
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_v1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			Brief:         "HTTP server request duration v1",
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     10,
			CreatedAt:     now.Add(-2 * time.Hour), // Older
			UpdatedAt:     now.Add(-2 * time.Hour),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: now.Add(-2 * time.Hour),
					LastSeen:  now,
				},
			},
		},
		{
			SchemaID:      "metric1_v2_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			Brief:         "HTTP server request duration v2",
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     15,
			CreatedAt:     now.Add(-1 * time.Hour), // Newer
			UpdatedAt:     now.Add(-1 * time.Hour),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: now.Add(-2 * time.Hour),
					LastSeen:  now,
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Should return only the latest version
	telemetries, err := repo.ListTelemetriesByEntity(ctx, "service")

	require.NoError(t, err)
	require.Len(t, telemetries, 1)
	require.Equal(t, "metric1_v2_schema_id", telemetries[0].SchemaID)
	require.Equal(t, "HTTP server request duration v2", telemetries[0].Brief)
}

func TestListScopes_NoScopes(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	params := query.ListQueryParams{
		Page:     1,
		PageSize: 10,
	}

	scopes, total, err := repo.ListScopes(ctx, params)

	require.NoError(t, err)
	require.Empty(t, scopes)
	require.Equal(t, 0, total)
}

func TestListScopes_WithScopes(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data - minimal telemetries with scopes
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Scope: &schema.Scope{
				ID:         "scope1",
				Name:       "@opentelemetry/instrumentation-http",
				Version:    "0.45.0",
				SchemaURL:  "https://opentelemetry.io/schemas/1.21.0",
				Attributes: map[string]interface{}{"library.language": "javascript"},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.client.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Scope: &schema.Scope{
				ID:         "scope2",
				Name:       "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
				Version:    "0.46.1",
				SchemaURL:  "https://opentelemetry.io/schemas/1.22.0",
				Attributes: map[string]interface{}{"library.language": "go"},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: List all scopes
	params := query.ListQueryParams{
		Page:     1,
		PageSize: 10,
	}

	scopes, total, err := repo.ListScopes(ctx, params)

	require.NoError(t, err)
	require.Len(t, scopes, 2)
	require.Equal(t, 2, total)

	// Verify scope data
	scopeMap := make(map[string]schema.Scope)
	for _, scope := range scopes {
		scopeMap[scope.ID] = scope
	}

	// Check first scope
	scope1, exists := scopeMap["scope1"]
	require.True(t, exists)
	require.Equal(t, "@opentelemetry/instrumentation-http", scope1.Name)
	require.Equal(t, "0.45.0", scope1.Version)
	require.Equal(t, "https://opentelemetry.io/schemas/1.21.0", scope1.SchemaURL)
	require.Equal(t, "javascript", scope1.Attributes["library.language"])

	// Check second scope
	scope2, exists := scopeMap["scope2"]
	require.True(t, exists)
	require.Equal(t, "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp", scope2.Name)
	require.Equal(t, "0.46.1", scope2.Version)
	require.Equal(t, "https://opentelemetry.io/schemas/1.22.0", scope2.SchemaURL)
	require.Equal(t, "go", scope2.Attributes["library.language"])
}

func TestListScopes_WithSearch(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric_search_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Scope: &schema.Scope{
				ID:        "scope_search",
				Name:      "@opentelemetry/instrumentation-http",
				Version:   "0.45.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.21.0",
				FirstSeen: time.Now(),
				LastSeen:  time.Now(),
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Search by name
	params := query.ListQueryParams{
		Page:     1,
		PageSize: 10,
		Search:   "instrumentation-http",
	}

	scopes, total, err := repo.ListScopes(ctx, params)

	require.NoError(t, err)
	require.Len(t, scopes, 1)
	require.Equal(t, 1, total)
	require.Equal(t, "@opentelemetry/instrumentation-http", scopes[0].Name)

	// Test: Search with no matches
	params.Search = "nonexistent"
	scopes, total, err = repo.ListScopes(ctx, params)

	require.NoError(t, err)
	require.Empty(t, scopes)
	require.Equal(t, 0, total)
}

func TestListScopes_ScopeConflictResolution(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()
	later := now.Add(1 * time.Hour)

	// First telemetry with scope
	firstTelemetry := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     now,
			UpdatedAt:     now,
			Scope: &schema.Scope{
				ID:        "scope_conflict",
				Name:      "@opentelemetry/instrumentation-http",
				Version:   "0.45.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.21.0",
				FirstSeen: now,
				LastSeen:  now,
			},
		},
	}

	// Second telemetry with same scope but later timestamp
	secondTelemetry := []schema.Telemetry{
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.client.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     later,
			UpdatedAt:     later,
			Scope: &schema.Scope{
				ID:        "scope_conflict", // Same scope ID
				Name:      "@opentelemetry/instrumentation-http",
				Version:   "0.45.0",
				SchemaURL: "https://opentelemetry.io/schemas/1.21.0",
				FirstSeen: now,
				LastSeen:  later, // Updated last_seen
			},
		},
	}

	// Register first telemetry
	err := repo.RegisterTelemetrySchemas(ctx, firstTelemetry)
	require.NoError(t, err)

	// Register second telemetry (should update last_seen)
	err = repo.RegisterTelemetrySchemas(ctx, secondTelemetry)
	require.NoError(t, err)

	// Verify scope was updated correctly
	params := query.ListQueryParams{
		Page:     1,
		PageSize: 10,
	}

	scopes, total, err := repo.ListScopes(ctx, params)

	require.NoError(t, err)
	require.Len(t, scopes, 1)
	require.Equal(t, 1, total)

	scope := scopes[0]
	require.Equal(t, "scope_conflict", scope.ID)
	require.Equal(t, "@opentelemetry/instrumentation-http", scope.Name)
	// LastSeen should be updated to the later time
	require.True(t, scope.LastSeen.After(scope.FirstSeen) || scope.LastSeen.Equal(later))
}

func TestListTelemetriesByScope_ScopeNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	telemetries, err := repo.ListTelemetriesByScope(ctx, "nonexistent-scope")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByScope_ScopeWithMetrics(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data with scopes
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			Brief:         "HTTP server request duration",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "my-service",
						"service.version":   "1.0.0",
						"service.namespace": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
			Scope: &schema.Scope{
				ID:         "scope1",
				Name:       "@opentelemetry/instrumentation-http",
				Version:    "1.0.0",
				SchemaURL:  "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.server.requests",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeSum,
			MetricUnit:    "1",
			Brief:         "HTTP server request count",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 5,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity2": {
					ID:   "entity2",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name":      "other-service",
						"service.version":   "1.0.0",
						"service.namespace": "other-namespace",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
			Scope: &schema.Scope{
				ID:         "scope2",
				Name:       "@opentelemetry/instrumentation-http",
				Version:    "1.1.0",
				SchemaURL:  "https://opentelemetry.io/schemas/1.1.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
		{
			SchemaID:      "metric3_schema_id",
			SchemaKey:     "cpu.usage",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "%",
			Brief:         "CPU usage percentage",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity3": {
					ID:   "entity3",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "my-pod",
						"k8s.namespace.name": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
			Scope: &schema.Scope{
				ID:         "scope3",
				Name:       "@opentelemetry/instrumentation-system",
				Version:    "2.0.0",
				SchemaURL:  "https://opentelemetry.io/schemas/2.0.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get metrics for http instrumentation scope
	telemetries, err := repo.ListTelemetriesByScope(ctx, "@opentelemetry/instrumentation-http")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right metrics for the scope
	schemaKeys := make([]string, len(telemetries))
	for i, telemetry := range telemetries {
		schemaKeys[i] = telemetry.SchemaKey
		require.Equal(t, schema.TelemetryTypeMetric, telemetry.TelemetryType)
		require.NotNil(t, telemetry.Scope)
		require.Equal(t, "@opentelemetry/instrumentation-http", telemetry.Scope.Name)
		require.NotEmpty(t, telemetry.Entities)
	}

	// Should contain both http metrics
	require.Contains(t, schemaKeys, "http.server.duration")
	require.Contains(t, schemaKeys, "http.server.requests")
	require.NotContains(t, schemaKeys, "cpu.usage")
}

func TestListTelemetriesByScope_ScopeWithNoTelemetries(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data with different scope
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "cpu.usage",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "%",
			Brief:         "CPU usage percentage",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "k8s",
					Attributes: map[string]interface{}{
						"k8s.pod.name":       "my-pod",
						"k8s.namespace.name": "default",
					},
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
			Scope: &schema.Scope{
				ID:         "scope1",
				Name:       "@opentelemetry/instrumentation-system",
				Version:    "2.0.0",
				SchemaURL:  "https://opentelemetry.io/schemas/2.0.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get telemetries for a different scope that has no telemetries
	telemetries, err := repo.ListTelemetriesByScope(ctx, "@opentelemetry/instrumentation-http")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByScope_LatestSchemaVersionOnly(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	// Insert test data with multiple versions of the same telemetry
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_v1_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			Brief:         "HTTP server request duration v1",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 10,
			CreatedAt: earlier,
			UpdatedAt: earlier,
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name": "my-service",
					},
					FirstSeen: earlier,
					LastSeen:  earlier,
				},
			},
			Scope: &schema.Scope{
				ID:         "scope1",
				Name:       "@opentelemetry/instrumentation-http",
				Version:    "1.0.0",
				SchemaURL:  "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  earlier,
				LastSeen:   earlier,
			},
		},
		{
			SchemaID:      "metric1_v2_schema_id",
			SchemaKey:     "http.server.duration",
			TelemetryType: schema.TelemetryTypeMetric,
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "s", // Changed unit
			Brief:         "HTTP server request duration v2",
			// Log fields
			LogSeverityNumber:         0,
			LogSeverityText:           "",
			LogBody:                   "",
			LogFlags:                  0,
			LogTraceID:                "",
			LogSpanID:                 "",
			LogEventName:              "",
			LogDroppedAttributesCount: 0,
			// Span fields
			SpanKind:    "",
			SpanName:    "",
			SpanID:      "",
			SpanTraceID: "",
			// Profile fields
			ProfileSampleAggregationTemporality: "",
			ProfileSampleUnit:                   "",
			// Common fields
			Protocol:  schema.TelemetryProtocolOTLP,
			SeenCount: 15,
			CreatedAt: later,
			UpdatedAt: later,
			Entities: map[string]*schema.Entity{
				"entity1": {
					ID:   "entity1",
					Type: "service",
					Attributes: map[string]interface{}{
						"service.name": "my-service",
					},
					FirstSeen: later,
					LastSeen:  later,
				},
			},
			Scope: &schema.Scope{
				ID:         "scope1",
				Name:       "@opentelemetry/instrumentation-http",
				Version:    "1.0.0",
				SchemaURL:  "https://opentelemetry.io/schemas/1.0.0",
				Attributes: map[string]interface{}{},
				FirstSeen:  later,
				LastSeen:   later,
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Should get only the latest version
	telemetries, err := repo.ListTelemetriesByScope(ctx, "@opentelemetry/instrumentation-http")

	require.NoError(t, err)
	require.Len(t, telemetries, 1)

	// Should be the latest version (v2)
	telemetry := telemetries[0]
	require.Equal(t, "http.server.duration", telemetry.SchemaKey)
	require.Equal(t, "s", telemetry.MetricUnit) // Should be the updated unit
	require.Equal(t, "HTTP server request duration v2", telemetry.Brief)
	require.Equal(t, "metric1_v2_schema_id", telemetry.SchemaID)
}
