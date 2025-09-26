package duckdb

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
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

		CREATE TABLE IF NOT EXISTS schema_producers (
			schema_id TEXT,
			producer_id TEXT,
			name TEXT,
			namespace TEXT,
			version TEXT,
			instance_id TEXT,
			first_seen TIMESTAMP,
			last_seen TIMESTAMP,
			FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id),
			PRIMARY KEY (schema_id, producer_id)
		);
	`)
	require.NoError(t, err)

	return &TelemetrySchemaRepository{
		pool: pool,
	}
}

func TestListTelemetriesByProducer_ProducerNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	telemetries, err := repo.ListTelemetriesByProducer(ctx, "non-existent-service", "1.0.0")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByProducer_ProducerWithMetrics(t *testing.T) {
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
			LogSeverityNumber: 0,
			LogSeverityText:   "",
			LogBody:           "",
			LogFlags:          0,
			LogTraceID:        "",
			LogSpanID:         "",
			LogEventName:      "",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     10,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
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
			LogSeverityNumber: 0,
			LogSeverityText:   "",
			LogBody:           "",
			LogFlags:          0,
			LogTraceID:        "",
			LogSpanID:         "",
			LogEventName:      "",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     5,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
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
			LogSeverityNumber: 0,
			LogSeverityText:   "",
			LogBody:           "",
			LogFlags:          0,
			LogTraceID:        "",
			LogSpanID:         "",
			LogEventName:      "",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     3,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer2": {
					Name:      "other-service",
					Version:   "2.0.0",
					Namespace: "default",
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get metrics for my-service v1.0.0
	telemetries, err := repo.ListTelemetriesByProducer(ctx, "my-service", "1.0.0")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right metrics
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "http.server.duration")
	require.Contains(t, schemaKeys, "http.server.requests")
	require.NotContains(t, schemaKeys, "cpu.usage")
}

func TestListTelemetriesByProducer_ProducerWithNoMetrics(t *testing.T) {
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
			LogSeverityNumber: 0,
			LogSeverityText:   "",
			LogBody:           "",
			LogFlags:          0,
			LogTraceID:        "",
			LogSpanID:         "",
			LogEventName:      "",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "other-service",
					Version:   "2.0.0",
					Namespace: "default",
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Look for a producer that has no metrics
	telemetries, err := repo.ListTelemetriesByProducer(ctx, "my-service", "1.0.0")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByProducer_ProducerWithLogRecords(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "log1_schema_id",
			SchemaKey:     "Reading CPU usage",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType:    schema.MetricTypeEmpty,
			MetricUnit:    "",
			Brief:         "",
			// Log fields
			LogSeverityNumber: 1,
			LogSeverityText:   "INFO",
			LogBody:           "Reading CPU usage",
			LogFlags:          0,
			LogTraceID:        "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:         "1234567890ABCDEF",
			LogEventName:      "Reading CPU usage",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     10,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
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
			MetricType:    schema.MetricTypeEmpty,
			MetricUnit:    "",
			Brief:         "",
			// Log fields
			LogSeverityNumber: 1,
			LogSeverityText:   "INFO",
			LogBody:           "Reading Total Memory",
			LogFlags:          0,
			LogTraceID:        "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:         "1234567890ABCDEF",
			LogEventName:      "Reading Total Memory",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     5,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
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
			MetricType:    schema.MetricTypeEmpty,
			MetricUnit:    "",
			Brief:         "",
			// Log fields
			LogSeverityNumber: 1,
			LogSeverityText:   "INFO",
			LogBody:           "HTTP server requested",
			LogFlags:          0,
			LogTraceID:        "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:         "1234567890ABCDEF",
			LogEventName:      "HTTP server requested",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     3,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer2": {
					Name:      "other-service",
					Version:   "2.0.0",
					Namespace: "default",
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get metrics for my-service v1.0.0
	telemetries, err := repo.ListTelemetriesByProducer(ctx, "my-service", "1.0.0")

	require.NoError(t, err)
	require.Len(t, telemetries, 2)

	// Verify we got the right metrics
	schemaKeys := make([]string, len(telemetries))
	for i, t := range telemetries {
		schemaKeys[i] = t.SchemaKey
	}
	require.Contains(t, schemaKeys, "Reading CPU usage")
	require.Contains(t, schemaKeys, "Reading Total Memory")
	require.NotContains(t, schemaKeys, "HTTP server requested")
}

func TestListTelemetriesByProducer_ProducerWithNoLogRecords(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Insert test data
	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "log1_schema_id",
			SchemaKey:     "Reading CPU usage",
			TelemetryType: schema.TelemetryTypeLog,
			// Metric fields
			MetricType:    schema.MetricTypeEmpty,
			MetricUnit:    "",
			Brief:         "",
			// Log fields
			LogSeverityNumber: 1,
			LogSeverityText:   "INFO",
			LogBody:           "Reading CPU usage",
			LogFlags:          0,
			LogTraceID:        "1234567890ABCDEF1234567890ABCDEF",
			LogSpanID:         "1234567890ABCDEF",
			LogEventName:      "Reading CPU usage",
			LogDroppedAttributesCount: 0,
			// Common fields
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     10,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				},
			},
		},
	}

	// Register test telemetries
	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Get metrics for my-service v1.0.0
	telemetries, err := repo.ListTelemetriesByProducer(ctx, "my-service", "2.0.0")

	require.NoError(t, err)
	require.Empty(t, telemetries)
}

func TestListTelemetriesByProducer_LatestSchemaVersionOnly(t *testing.T) {
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
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
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
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
					FirstSeen: now.Add(-2 * time.Hour),
					LastSeen:  now,
				},
			},
		},
	}

	err := repo.RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test: Should return only the latest version
	telemetries, err := repo.ListTelemetriesByProducer(ctx, "my-service", "1.0.0")

	require.NoError(t, err)
	require.Len(t, telemetries, 1)
	require.Equal(t, "metric1_v2_schema_id", telemetries[0].SchemaID)
	require.Equal(t, "HTTP server request duration v2", telemetries[0].Brief)
}
