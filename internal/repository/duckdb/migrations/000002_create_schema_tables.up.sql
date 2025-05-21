-- Create telemetry_schemas table
CREATE TABLE IF NOT EXISTS telemetry_schemas (
    schema_id TEXT PRIMARY KEY,
    schema_key TEXT,
    schema_version TEXT,
    schema_url TEXT,
    signal_type TEXT,
    metric_type TEXT,
    temporality TEXT,
    unit TEXT,
    brief TEXT,
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