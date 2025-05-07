# Schema Implementation Plan

## Overview
This document outlines the implementation plan for the schema tracking system in TallyCat. The plan is divided into phases, each focusing on specific features and optimizations.

## Phase 1: Core Schema Management (MVP)

### 1. Schema ID Generation Optimization
- Replace SHA256 with xxHash for better performance
- Implement efficient field sorting and hashing
- Add version tracking to schema IDs

### 2. Schema Storage Optimization
```sql
-- Core schema table (frequently accessed)
CREATE TABLE schema_core (
    schema_id TEXT PRIMARY KEY,
    signal_type TEXT,
    scope_name TEXT,
    scope_version TEXT,
    seen_count INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_scope (scope_name, scope_version)
);

-- Schema details (less frequently accessed)
CREATE TABLE schema_details (
    schema_id TEXT PRIMARY KEY,
    metric_type TEXT,
    unit TEXT,
    field_names TEXT[],
    field_types JSON,
    field_sources JSON,
    field_cardinality JSON,
    FOREIGN KEY (schema_id) REFERENCES schema_core(schema_id)
);
```

### 3. Memory Efficient Field Storage
- Implement field name interning
- Use efficient JSON serialization
- Add field type normalization

## Phase 2: Performance & Scalability

### 1. Batch Processing
- Implement efficient batch operations
- Add background processing
- Implement retry mechanisms

### 2. Caching Layer
- Add in-memory caching for frequent schemas
- Implement cache invalidation
- Add cache metrics

### 3. Query Optimization
- Add materialized views for common queries
- Implement efficient indexing
- Add query performance monitoring

## Phase 3: Schema Versioning & Diff

### 1. Schema Versioning
```sql
CREATE TABLE schema_versions (
    version_id TEXT PRIMARY KEY,
    schema_id TEXT,
    previous_version_id TEXT,
    change_type TEXT,
    changed_at TIMESTAMP,
    changes JSON,
    FOREIGN KEY (schema_id) REFERENCES schema_core(schema_id)
);
```

### 2. Diff View Support
- Implement efficient diff generation
- Add diff visualization
- Implement change tracking

## Phase 4: Cardinality Tracking

### 1. HyperLogLog Implementation
```sql
CREATE TABLE schema_cardinality (
    schema_id TEXT,
    field_name TEXT,
    window_start TIMESTAMP,
    window_end TIMESTAMP,
    cardinality_estimate HLL,
    PRIMARY KEY (schema_id, field_name, window_start)
) PARTITION BY RANGE (window_start);
```

### 2. Cardinality Tracking Logic
- Implement window-based cardinality tracking
- Add background processing for cardinality updates
- Implement efficient HLL merging

## Phase 5: Schema Lineage

### 1. Lineage Tracking
```sql
CREATE TABLE schema_lineage (
    schema_id TEXT,
    producer_id TEXT,
    producer_type TEXT,
    first_seen TIMESTAMP,
    last_seen TIMESTAMP,
    PRIMARY KEY (schema_id, producer_id)
);

CREATE TABLE schema_consumers (
    schema_id TEXT,
    consumer_id TEXT,
    consumer_type TEXT,
    first_used TIMESTAMP,
    last_used TIMESTAMP,
    PRIMARY KEY (schema_id, consumer_id)
);
```

### 2. Service Integration
- Add service discovery integration
- Implement producer/consumer tracking
- Add usage analytics

## Key Considerations

### Performance
- Use batch processing
- Implement efficient indexing
- Use connection pooling
- Add caching where appropriate

### Memory Efficiency
- Use string interning
- Implement field compression
- Use efficient data structures

### Scalability
- Design for horizontal scaling
- Implement efficient partitioning
- Use background processing

### Monitoring
- Add performance metrics
- Implement health checks
- Add usage analytics

## Implementation Notes
- Each phase should be implemented and tested independently
- Performance metrics should be collected before and after each phase
- Documentation should be updated as features are implemented
- Backward compatibility should be maintained throughout the implementation 