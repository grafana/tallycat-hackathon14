-- Drop performance indexes for telemetry catalogs

-- Telemetry schemas table indexes
DROP INDEX IF EXISTS idx_telemetry_schemas_signal_type;
DROP INDEX IF EXISTS idx_telemetry_schemas_schema_key;
DROP INDEX IF EXISTS idx_telemetry_schemas_updated_at;
DROP INDEX IF EXISTS idx_telemetry_schemas_partition_order;
DROP INDEX IF EXISTS idx_telemetry_schemas_schema_id;
DROP INDEX IF EXISTS idx_telemetry_schemas_metric_type;
DROP INDEX IF EXISTS idx_telemetry_schemas_unit;

-- Entity table indexes
DROP INDEX IF EXISTS idx_telemetry_entities_entity_type;
DROP INDEX IF EXISTS idx_telemetry_entities_last_seen;
DROP INDEX IF EXISTS idx_telemetry_entities_entity_id;
DROP INDEX IF EXISTS idx_telemetry_entities_type_last_seen;

-- Scope table indexes
DROP INDEX IF EXISTS idx_telemetry_scopes_name;
DROP INDEX IF EXISTS idx_telemetry_scopes_version;
DROP INDEX IF EXISTS idx_telemetry_scopes_schema_url;
DROP INDEX IF EXISTS idx_telemetry_scopes_last_seen;

-- Join table indexes
DROP INDEX IF EXISTS idx_schema_entities_schema_id;
DROP INDEX IF EXISTS idx_schema_entities_entity_id;
DROP INDEX IF EXISTS idx_schema_scopes_schema_id;
DROP INDEX IF EXISTS idx_schema_scopes_scope_id;

-- Attribute tables indexes
DROP INDEX IF EXISTS idx_schema_attributes_schema_id;
DROP INDEX IF EXISTS idx_schema_attributes_name;
DROP INDEX IF EXISTS idx_entity_attributes_entity_id;
DROP INDEX IF EXISTS idx_entity_attributes_name;
DROP INDEX IF EXISTS idx_scope_attributes_scope_id;
DROP INDEX IF EXISTS idx_scope_attributes_name;

-- Telemetry history table indexes
DROP INDEX IF EXISTS idx_telemetry_history_schema_key;
DROP INDEX IF EXISTS idx_telemetry_history_version;
DROP INDEX IF EXISTS idx_telemetry_history_timestamp;
DROP INDEX IF EXISTS idx_telemetry_history_key_timestamp;
