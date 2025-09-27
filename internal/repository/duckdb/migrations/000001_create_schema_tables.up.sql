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

CREATE TABLE IF NOT EXISTS schema_producers (
    schema_id TEXT,
    producer_id TEXT,
    name TEXT,
    namespace TEXT,
    version TEXT,
    instance_id TEXT,
    first_seen TIMESTAMP,
    last_seen TIMESTAMP,
    FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id),
    PRIMARY KEY (schema_id, producer_id)
);
