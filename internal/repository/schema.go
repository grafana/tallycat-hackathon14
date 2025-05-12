package repository

import (
	"context"
	"time"

	"github.com/tallycat/tallycat/internal/repository/query"
)

// Schema represents a telemetry schema in the repository
type Schema struct {
	SchemaID           string
	SignalType         string
	SignalKey          string // Generic key: metric_name for metrics, operation for spans, etc.
	ScopeName          string
	ScopeVersion       string
	SchemaURL          string
	MetricType         *string
	Unit               *string
	FieldNames         []string
	FieldTypes         map[string]string
	FieldSources       map[string]string
	FieldCardinality   map[string]bool
	SeenCount          int
	SchemaVersionCount int // Number of versions for this schema
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SchemaProvider defines the interface for schema storage
type SchemaProvider interface {
	// RegisterSchema registers a new schema or updates an existing one
	RegisterSchema(ctx context.Context, schema *Schema) error

	// ListSchemas returns a paginated, filtered, and searched list of schemas
	ListSchemas(ctx context.Context, params query.ListQueryParams) ([]*Schema, int, error)
}
