-- Create schema_core table
CREATE TABLE IF NOT EXISTS schema_core (
    schema_id TEXT PRIMARY KEY,
    schema_url TEXT,
    signal_type TEXT,
    signal_key TEXT,  -- Generic key: metric_name for metrics, operation for spans, etc.
    scope_name TEXT,
    scope_version TEXT,
    seen_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Create schema_details table
CREATE TABLE IF NOT EXISTS schema_details (
    schema_id TEXT PRIMARY KEY,
    metric_type TEXT,
    unit TEXT,
    field_names TEXT[],
    field_types JSON,
    field_sources JSON,
    field_cardinality JSON,
    FOREIGN KEY (schema_id) REFERENCES schema_core(schema_id)
); 