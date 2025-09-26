package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tallycat/tallycat/internal/repository/query"
	"github.com/tallycat/tallycat/internal/schema"
)

// MockTelemetrySchemaRepository is a mock implementation of the repository interface
type MockTelemetrySchemaRepository struct {
	mock.Mock
}

func (m *MockTelemetrySchemaRepository) RegisterTelemetrySchemas(ctx context.Context, schemas []schema.Telemetry) error {
	args := m.Called(ctx, schemas)
	return args.Error(0)
}

func (m *MockTelemetrySchemaRepository) ListTelemetries(ctx context.Context, params query.ListQueryParams) ([]schema.Telemetry, int, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]schema.Telemetry), args.Int(1), args.Error(2)
}

func (m *MockTelemetrySchemaRepository) GetTelemetry(ctx context.Context, schemaKey string) (*schema.Telemetry, error) {
	args := m.Called(ctx, schemaKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Telemetry), args.Error(1)
}

func (m *MockTelemetrySchemaRepository) ListTelemetrySchemas(ctx context.Context, schemaKey string, params query.ListQueryParams) ([]schema.TelemetrySchema, int, error) {
	args := m.Called(ctx, schemaKey, params)
	return args.Get(0).([]schema.TelemetrySchema), args.Int(1), args.Error(2)
}

func (m *MockTelemetrySchemaRepository) AssignTelemetrySchemaVersion(ctx context.Context, assignment schema.SchemaAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockTelemetrySchemaRepository) GetTelemetrySchema(ctx context.Context, schemaId string) (*schema.TelemetrySchema, error) {
	args := m.Called(ctx, schemaId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.TelemetrySchema), args.Error(1)
}

func (m *MockTelemetrySchemaRepository) ListTelemetriesByProducer(ctx context.Context, producerName, producerVersion string) ([]schema.Telemetry, error) {
	args := m.Called(ctx, producerName, producerVersion)
	return args.Get(0).([]schema.Telemetry), args.Error(1)
}

func TestHandleProducerWeaverSchemaExport_ProducerNotFound(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByProducer", mock.Anything, "non-existent-service", "1.0.0").
		Return([]schema.Telemetry{}, nil)

	handler := HandleProducerWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/producers/non-existent-service@1.0.0/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("producerNameVersion", "non-existent-service@1.0.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	// Since we can't distinguish between "not found" and "no metrics", we return 204
	require.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandleProducerWeaverSchemaExport_ProducerWithNoMetrics(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByProducer", mock.Anything, "empty-service", "1.0.0").
		Return([]schema.Telemetry{}, nil)

	handler := HandleProducerWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/producers/empty-service@1.0.0/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("producerNameVersion", "empty-service@1.0.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandleProducerWeaverSchemaExport_ProducerWithMetrics(t *testing.T) {
	now := time.Now()
	mockTelemetries := []schema.Telemetry{
		{
			SchemaID:      "metric1_schema_id",
			SchemaKey:     "http.server.duration",
			Brief:         "HTTP server request duration",
			MetricType:    schema.MetricTypeHistogram,
			MetricUnit:    "ms",
			TelemetryType: schema.TelemetryTypeMetric,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "http.method",
					Type:   schema.AttributeTypeStr,
					Source: schema.AttributeSourceDataPoint,
				},
			},
		},
		{
			SchemaID:      "metric2_schema_id",
			SchemaKey:     "http.server.requests",
			Brief:         "HTTP server request count",
			MetricType:    schema.MetricTypeSum,
			MetricUnit:    "1",
			TelemetryType: schema.TelemetryTypeMetric,
			CreatedAt:     now,
			UpdatedAt:     now,
			Attributes: []schema.Attribute{
				{
					Name:   "http.status_code",
					Type:   schema.AttributeTypeInt,
					Source: schema.AttributeSourceDataPoint,
				},
			},
		},
	}

	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByProducer", mock.Anything, "my-service", "1.0.0").
		Return(mockTelemetries, nil)

	handler := HandleProducerWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/producers/my-service@1.0.0/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("producerNameVersion", "my-service@1.0.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	require.Equal(t, "attachment; filename=my-service@1.0.0.zip", w.Header().Get("Content-Disposition"))
	require.NotEmpty(t, w.Body.Bytes())

	mockRepo.AssertExpectations(t)
}

func TestHandleProducerWeaverSchemaExport_InvalidProducerFormat(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)

	handler := HandleProducerWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/producers/invalid/weaver-schema.zip", nil)

	// Set up chi context with URL parameters - using a format without hyphen
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("producerNameVersion", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid producer format")

	// No repository calls should be made
	mockRepo.AssertExpectations(t)
}

func TestHandleProducerWeaverSchemaExport_RepositoryError(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByProducer", mock.Anything, "error-service", "1.0.0").
		Return([]schema.Telemetry{}, fmt.Errorf("database error"))

	handler := HandleProducerWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/producers/error-service@1.0.0/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("producerNameVersion", "error-service@1.0.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "failed to get telemetries for producer")

	mockRepo.AssertExpectations(t)
}

func TestParseProducerNameVersion(t *testing.T) {
	tests := []struct {
		input           string
		expectedName    string
		expectedVersion string
		expectError     bool
	}{
		{"my-service@1.0.0", "my-service", "1.0.0", false},
		{"service@2.1.0", "service", "2.1.0", false},
		{"complex-service-name@1.2.3", "complex-service-name", "1.2.3", false},
		{"invalid", "", "", true},
		{"@invalid", "", "", true},
		{"invalid@", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, version, err := parseProducerNameVersion(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedName, name)
				require.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}
