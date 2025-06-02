CREATE TABLE IF NOT EXISTS schema_versions (
    schema_id     TEXT,
    version       TEXT,
    assigned_by   TEXT,
    reason        TEXT,
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP,
    FOREIGN KEY (schema_id) REFERENCES telemetry_schemas(schema_id),
    PRIMARY KEY (schema_id)
);
