package repository

import (
	"context"

	"github.com/tallycat/tallycat/internal/repository/query"
	"github.com/tallycat/tallycat/internal/schema"
)

type TelemetrySchemaRepository interface {
	RegisterTelemetrySchemas(ctx context.Context, schemas []schema.Telemetry) error
	ListSchemas(ctx context.Context, params query.ListQueryParams) ([]schema.Telemetry, int, error)
	GetSchemaByKey(ctx context.Context, schemaKey string) (*schema.Telemetry, error)
}
