-- Drop tables in correct order (child table first due to foreign key)
DROP TABLE IF EXISTS schema_attributes;
DROP TABLE IF EXISTS telemetry_schemas;