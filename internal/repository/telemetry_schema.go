package repository

import (
	"context"

	"github.com/tallycat/tallycat/internal/repository/query"
	"github.com/tallycat/tallycat/internal/schema"
)

type TelemetrySchemaRepository interface {
	RegisterTelemetrySchemas(ctx context.Context, schemas []schema.Telemetry) error
	ListTelemetries(ctx context.Context, params query.ListQueryParams) ([]schema.Telemetry, int, error)
	GetTelemetry(ctx context.Context, schemaKey string) (*schema.Telemetry, error)
	ListTelemetrySchemas(ctx context.Context, schemaKey string, params query.ListQueryParams) ([]schema.TelemetrySchema, int, error)
	AssignTelemetrySchemaVersion(ctx context.Context, assignment schema.SchemaAssignment) error
	GetTelemetrySchema(ctx context.Context, schemaId string) (*schema.TelemetrySchema, error)
	ListEntities(ctx context.Context, params query.ListQueryParams) ([]schema.Entity, int, error)
	ListEntitiesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Entity, error)
	ListTelemetriesByEntity(ctx context.Context, entityType string) ([]schema.Telemetry, error)
	ListScopes(ctx context.Context, params query.ListQueryParams) ([]schema.Scope, int, error)
	ListScopesByTelemetry(ctx context.Context, telemetryKey string) ([]schema.Scope, error)
	ListTelemetriesByScope(ctx context.Context, scopeName string) ([]schema.Telemetry, error)
}

type TelemetryHistoryRepository interface {
	InsertTelemetryHistory(ctx context.Context, h *schema.TelemetryHistory) error
	ListTelemetryHistory(ctx context.Context, telemetryID string, page, pageSize int) ([]schema.TelemetryHistory, int, error)
}
