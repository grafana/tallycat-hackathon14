-- Create telemetry_schemas table
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

-- Create schema_attributes table
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
