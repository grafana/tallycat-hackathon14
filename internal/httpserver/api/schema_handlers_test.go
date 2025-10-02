package api

import (
	"context"
	"encoding/json"
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

func (m *MockTelemetrySchemaRepository) ListTelemetriesByEntity(ctx context.Context, entityType string) ([]schema.Telemetry, error) {
	args := m.Called(ctx, entityType)
	return args.Get(0).([]schema.Telemetry), args.Error(1)
}

func (m *MockTelemetrySchemaRepository) ListTelemetriesByScope(ctx context.Context, scopeName string) ([]schema.Telemetry, error) {
	args := m.Called(ctx, scopeName)
	return args.Get(0).([]schema.Telemetry), args.Error(1)
}

func (m *MockTelemetrySchemaRepository) ListScopes(ctx context.Context, params query.ListQueryParams) ([]schema.Scope, int, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]schema.Scope), args.Int(1), args.Error(2)
}

func (m *MockTelemetrySchemaRepository) ListEntities(ctx context.Context, params query.ListQueryParams) ([]schema.Entity, int, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]schema.Entity), args.Int(1), args.Error(2)
}

func (m *MockTelemetrySchemaRepository) ListEntitiesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Entity, error) {
	args := m.Called(ctx, telemetryKey)
	return args.Get(0).([]schema.Entity), args.Error(1)
}

func (m *MockTelemetrySchemaRepository) ListScopesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Scope, error) {
	args := m.Called(ctx, telemetryKey)
	return args.Get(0).([]schema.Scope), args.Error(1)
}

func TestHandleEntityWeaverSchemaExport_EntityNotFound(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByEntity", mock.Anything, "service").
		Return([]schema.Telemetry{}, nil)

	handler := HandleEntityWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/entities/service/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("entityType", "service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	// Since we can't distinguish between "not found" and "no metrics", we return 204
	require.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandleEntityWeaverSchemaExport_EntityWithNoMetrics(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByEntity", mock.Anything, "service").
		Return([]schema.Telemetry{}, nil)

	handler := HandleEntityWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/entities/service/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("entityType", "service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandleEntityWeaverSchemaExport_EntityWithMetrics(t *testing.T) {
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
	mockRepo.On("ListTelemetriesByEntity", mock.Anything, "service").
		Return(mockTelemetries, nil)

	handler := HandleEntityWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/entities/service/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("entityType", "service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/zip", w.Header().Get("Content-Type"))
	require.Equal(t, "attachment; filename=service.zip", w.Header().Get("Content-Disposition"))
	require.NotEmpty(t, w.Body.Bytes())

	mockRepo.AssertExpectations(t)
}

func TestHandleEntityWeaverSchemaExport_InvalidEntityFormat(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByEntity", mock.Anything, "service").
		Return([]schema.Telemetry{}, nil)

	handler := HandleEntityWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/entities/service/weaver-schema.zip", nil)

	// Set up chi context with URL parameters - using a format without hyphen
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("entityType", "service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)

	mockRepo.AssertExpectations(t)
}

func TestHandleEntityWeaverSchemaExport_RepositoryError(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListTelemetriesByEntity", mock.Anything, "service").
		Return([]schema.Telemetry{}, fmt.Errorf("database error"))

	handler := HandleEntityWeaverSchemaExport(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/entities/service/weaver-schema.zip", nil)

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("entityType", "service")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "failed to get telemetries for entity")

	mockRepo.AssertExpectations(t)
}

func TestHandleScopeList_EmptyResult(t *testing.T) {
	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListScopes", mock.Anything, mock.AnythingOfType("query.ListQueryParams")).
		Return([]schema.Scope{}, 0, nil)

	handler := HandleScopeList(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/scopes", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ListResponse[schema.Scope]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Empty(t, response.Items)
	require.Equal(t, 0, response.Total)

	mockRepo.AssertExpectations(t)
}

func TestHandleScopeList_WithScopes(t *testing.T) {
	mockScopes := []schema.Scope{
		{
			ID:        "scope1",
			Name:      "@opentelemetry/instrumentation-http",
			Version:   "0.45.0",
			SchemaURL: "https://opentelemetry.io/schemas/1.21.0",
		},
		{
			ID:        "scope2",
			Name:      "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			Version:   "0.46.1",
			SchemaURL: "https://opentelemetry.io/schemas/1.22.0",
		},
	}

	mockRepo := new(MockTelemetrySchemaRepository)
	mockRepo.On("ListScopes", mock.Anything, mock.AnythingOfType("query.ListQueryParams")).
		Return(mockScopes, 2, nil)

	handler := HandleScopeList(mockRepo)

	req := httptest.NewRequest("GET", "/api/v1/scopes", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ListResponse[schema.Scope]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response.Items, 2)
	require.Equal(t, 2, response.Total)
	require.Equal(t, "@opentelemetry/instrumentation-http", response.Items[0].Name)
	require.Equal(t, "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp", response.Items[1].Name)

	mockRepo.AssertExpectations(t)
}
