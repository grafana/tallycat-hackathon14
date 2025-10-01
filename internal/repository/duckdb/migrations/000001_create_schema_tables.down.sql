-- Drop tables in correct order (child table first due to foreign key)
DROP TABLE IF EXISTS schema_scopes;
DROP TABLE IF EXISTS scope_attributes;
DROP TABLE IF EXISTS telemetry_scopes;
DROP TABLE IF EXISTS schema_attributes;
DROP TABLE IF EXISTS schema_entities;
DROP TABLE IF EXISTS entity_attributes;
DROP TABLE IF EXISTS telemetry_entities;
DROP TABLE IF EXISTS telemetry_schemas;