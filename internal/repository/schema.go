package repository

import (
	"context"
	"time"
)

// Schema represents a telemetry schema in the repository
type Schema struct {
	SchemaID         string
	SignalType       string
	SignalKey        string // Generic key: metric_name for metrics, operation for spans, etc.
	ScopeName        string
	ScopeVersion     string
	SchemaURL        string
	MetricType       *string
	Unit             *string
	FieldNames       []string
	FieldTypes       map[string]string
	FieldSources     map[string]string
	FieldCardinality map[string]bool
	SeenCount        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SchemaProvider defines the interface for schema storage
type SchemaProvider interface {
	// RegisterSchema registers a new schema or updates an existing one
	RegisterSchema(ctx context.Context, schema *Schema) error

	// GetSchema retrieves a schema by its ID
	GetSchema(ctx context.Context, schemaID string) (*Schema, error)

	// GetSchemasByScope retrieves schemas by scope name and version
	GetSchemasByScope(ctx context.Context, scopeName, scopeVersion string) ([]*Schema, error)

	// GetSchemasBySignalType retrieves schemas by signal type
	GetSchemasBySignalType(ctx context.Context, signalType string) ([]*Schema, error)

	// UpdateSchemaSeenCount updates the seen count for a schema
	UpdateSchemaSeenCount(ctx context.Context, schemaID string, count int) error

	// DeleteSchema deletes a schema by its ID
	DeleteSchema(ctx context.Context, schemaID string) error

	// UpdateFieldCardinality updates the cardinality information for a field
	UpdateFieldCardinality(ctx context.Context, fieldName string, isHighCardinality bool) error
}
