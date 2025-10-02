-- Performance indexes for telemetry catalogs

-- Telemetry schemas table indexes
-- Primary filtering and search patterns
CREATE INDEX IF NOT EXISTS idx_telemetry_schemas_signal_type ON telemetry_schemas(signal_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_schemas_schema_key ON telemetry_schemas(schema_key);

-- Search pattern indexes
CREATE INDEX IF NOT EXISTS idx_telemetry_schemas_schema_id ON telemetry_schemas(schema_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_schemas_metric_type ON telemetry_schemas(metric_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_schemas_unit ON telemetry_schemas(unit);

-- Entity table indexes
-- Primary filtering and search patterns
CREATE INDEX IF NOT EXISTS idx_telemetry_entities_entity_type ON telemetry_entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_telemetry_entities_entity_id ON telemetry_entities(entity_id);

-- Scope table indexes  
-- Primary search and ordering patterns
CREATE INDEX IF NOT EXISTS idx_telemetry_scopes_name ON telemetry_scopes(name);
CREATE INDEX IF NOT EXISTS idx_telemetry_scopes_version ON telemetry_scopes(version);
CREATE INDEX IF NOT EXISTS idx_telemetry_scopes_schema_url ON telemetry_scopes(schema_url);
CREATE INDEX IF NOT EXISTS idx_telemetry_scopes_last_seen ON telemetry_scopes(last_seen DESC);

-- Join table indexes
-- Schema-entity relationships (heavily used in telemetry queries)
CREATE INDEX IF NOT EXISTS idx_schema_entities_schema_id ON schema_entities(schema_id);
CREATE INDEX IF NOT EXISTS idx_schema_entities_entity_id ON schema_entities(entity_id);

-- Schema-scope relationships (heavily used in telemetry queries)
CREATE INDEX IF NOT EXISTS idx_schema_scopes_schema_id ON schema_scopes(schema_id);
CREATE INDEX IF NOT EXISTS idx_schema_scopes_scope_id ON schema_scopes(scope_id);

-- Attribute tables indexes
-- Schema attributes (used for attribute lookups)
CREATE INDEX IF NOT EXISTS idx_schema_attributes_schema_id ON schema_attributes(schema_id);
CREATE INDEX IF NOT EXISTS idx_schema_attributes_name ON schema_attributes(name);

-- Entity attributes (used for entity attribute lookups)
CREATE INDEX IF NOT EXISTS idx_entity_attributes_entity_id ON entity_attributes(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_attributes_name ON entity_attributes(name);

-- Scope attributes (used for scope attribute lookups)
CREATE INDEX IF NOT EXISTS idx_scope_attributes_scope_id ON scope_attributes(scope_id);
CREATE INDEX IF NOT EXISTS idx_scope_attributes_name ON scope_attributes(name);

