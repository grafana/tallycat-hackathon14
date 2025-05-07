package repository

import (
	"context"
	"database/sql"
	"time"
)

type ConnectionProvider interface {
	GetConnection() *sql.DB
	Close() error
	HealthCheck() error
}

type SchemaProvider interface {
	CreateSchemaTable(ctx context.Context) error
	RegisterSchema(ctx context.Context, schema *Schema) error
	GetSchema(ctx context.Context, schemaID string) (*Schema, error)
}

type Schema struct {
	SchemaID         string            `json:"schema_id"`
	SignalType       string            `json:"signal_type"`
	MetricType       *string           `json:"metric_type,omitempty"`
	Unit             *string           `json:"unit,omitempty"`
	ScopeName        string            `json:"scope_name"`
	ScopeVersion     string            `json:"scope_version"`
	FieldNames       []string          `json:"field_names"`
	FieldTypes       map[string]string `json:"field_types"`
	FieldSources     map[string]string `json:"field_sources"`
	FieldCardinality map[string]bool   `json:"field_cardinality"`
	SeenCount        int               `json:"seen_count"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
