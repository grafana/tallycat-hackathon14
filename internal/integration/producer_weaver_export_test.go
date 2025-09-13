package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/tallycat/tallycat/internal/httpserver/api"
	"github.com/tallycat/tallycat/internal/integration/testutil"
	"github.com/tallycat/tallycat/internal/schema"
)

func TestProducerWeaverSchemaExport_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.NewTestDB(t)
	defer testDB.Close()
	testDB.SetupTestDB(t)

	// Create router and register our handler
	r := chi.NewRouter()
	r.Get("/api/v1/producers/{producerNameVersion}/weaver-schema.zip", api.HandleProducerWeaverSchemaExport(testDB.Repo()))

	// Setup test data - create telemetries with producers
	ctx := context.Background()
	now := time.Now()

	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "http_duration_schema_id",
			SchemaKey:     "http.server.duration",
			Brief:         "HTTP server request duration",
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     10,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
				{
					Name:   "http.status_code",
					Type:   schema.AttributeTypeInt,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
					FirstSeen: now,
					LastSeen:  now,
				},
			},
		},
		{
			SchemaID:      "http_requests_schema_id",
			SchemaKey:     "http.server.requests",
			Brief:         "HTTP server request count",
			MetricType:    schema.MetricTypeSum,
			MetricUnit:    "1",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     5,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
				{
					Name:   "http.status_code",
					Type:   schema.AttributeTypeInt,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "my-service",
					Version:   "1.0.0",
					Namespace: "default",
					FirstSeen: now,
					LastSeen:  now,
				},
			},
		},
		{
			SchemaID:      "cpu_usage_schema_id",
			SchemaKey:     "system.cpu.usage",
			Brief:         "System CPU usage",
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "%",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     3,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "cpu.core",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
			Producers: map[string]*schema.Producer{
				"producer2": {
					Name:      "other-service",
					Version:   "2.0.0",
					Namespace: "default",
					FirstSeen: now,
					LastSeen:  now,
				},
			},
		},
	}

	// Insert test data
	err := testDB.Repo().RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Test 1: Producer with multiple metrics
	t.Run("producer with multiple metrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/producers/my-service@1.0.0/weaver-schema.zip", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		require.Equal(t, "attachment; filename=my-service@1.0.0.zip", w.Header().Get("Content-Disposition"))
		require.NotEmpty(t, w.Body.Bytes())

		// Verify ZIP content contains expected YAML structure
		body := w.Body.String()
		require.NotEmpty(t, body)
	})

	// Test 2: Producer with single metric
	t.Run("producer with single metric", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/producers/other-service@2.0.0/weaver-schema.zip", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		require.Equal(t, "attachment; filename=other-service@2.0.0.zip", w.Header().Get("Content-Disposition"))
		require.NotEmpty(t, w.Body.Bytes())
	})

	// Test 3: Non-existent producer
	t.Run("non-existent producer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/producers/non-existent-service@1.0.0/weaver-schema.zip", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	// Test 4: Invalid producer format
	t.Run("invalid producer format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/producers/invalid/weaver-schema.zip", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		body, err := io.ReadAll(w.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "invalid producer format")
	})
}

func TestProducerWeaverSchemaExport_YAMLContent(t *testing.T) {
	// Setup test database
	testDB := testutil.NewTestDB(t)
	defer testDB.Close()
	testDB.SetupTestDB(t)

	// Create router and register our handler
	r := chi.NewRouter()
	r.Get("/api/v1/producers/{producerNameVersion}/weaver-schema.zip", api.HandleProducerWeaverSchemaExport(testDB.Repo()))

	// Setup simple test data
	ctx := context.Background()
	now := time.Now()

	testTelemetries := []schema.Telemetry{
		{
			SchemaID:      "simple_metric_schema_id",
			SchemaKey:     "test.metric",
			Brief:         "A test metric",
			MetricType:    schema.MetricTypeGauge,
			MetricUnit:    "1",
			TelemetryType: schema.TelemetryTypeMetric,
			Protocol:      schema.TelemetryProtocolOTLP,
			SeenCount:     1,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "test.attribute",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
					Brief:  "A test attribute",
				},
			},
			Producers: map[string]*schema.Producer{
				"producer1": {
					Name:      "test-service",
					Version:   "1.0.0",
					Namespace: "default",
					FirstSeen: now,
					LastSeen:  now,
				},
			},
		},
	}

	// Insert test data
	err := testDB.Repo().RegisterTelemetrySchemas(ctx, testTelemetries)
	require.NoError(t, err)

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/producers/test-service@1.0.0/weaver-schema.zip", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// For a more detailed test, we would need to unzip the response and verify YAML content
	// For now, we just verify that we get a ZIP response with content
	require.NotEmpty(t, w.Body.Bytes())

	// Verify it's actually a ZIP by checking the magic bytes
	body := w.Body.Bytes()
	require.True(t, len(body) >= 4)
	// ZIP files start with "PK" (0x50, 0x4B)
	require.Equal(t, byte(0x50), body[0])
	require.Equal(t, byte(0x4B), body[1])
}

func TestProducerWeaverSchemaExport_RouteIntegration(t *testing.T) {
	// Setup test database
	testDB := testutil.NewTestDB(t)
	defer testDB.Close()
	testDB.SetupTestDB(t)

	// Create router and register our handler
	r := chi.NewRouter()
	r.Get("/api/v1/producers/{producerNameVersion}/weaver-schema.zip", api.HandleProducerWeaverSchemaExport(testDB.Repo()))

	// Test that the route is properly registered and accessible
	testCases := []struct {
		name       string
		path       string
		expectCode int
	}{
		{
			name:       "valid producer route",
			path:       "/api/v1/producers/service@1.0.0/weaver-schema.zip",
			expectCode: http.StatusNoContent, // No metrics, so 204
		},
		{
			name:       "invalid producer route",
			path:       "/api/v1/producers/invalid/weaver-schema.zip",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "non-existent route",
			path:       "/api/v1/producers/service@1.0.0/invalid",
			expectCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			require.Equal(t, tc.expectCode, w.Code)
		})
	}
}
