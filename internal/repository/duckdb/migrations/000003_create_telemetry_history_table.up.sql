CREATE SEQUENCE IF NOT EXISTS telemetry_history_id_seq START 1;

CREATE TABLE IF NOT EXISTS telemetry_history (
    id INTEGER PRIMARY KEY DEFAULT nextval('telemetry_history_id_seq'),
    schema_key TEXT NOT NULL,
    version TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    author TEXT,
    summary TEXT,
    status TEXT,
    snapshot BLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
); 