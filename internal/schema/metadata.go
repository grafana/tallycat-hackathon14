package schema

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// SchemaVersion represents a specific version of a schema
type SchemaVersion struct {
	VersionID       string    `json:"versionId"`
	SchemaID        string    `json:"schemaId"`
	PreviousVersion string    `json:"previousVersion,omitempty"`
	ChangeType      string    `json:"changeType"`
	ChangedAt       time.Time `json:"changedAt"`
	Changes         []Change  `json:"changes"`
}

// Change represents a single change in a schema version
type Change struct {
	FieldName   string `json:"fieldName"`
	ChangeType  string `json:"changeType"` // ADD, REMOVE, MODIFY
	OldValue    string `json:"oldValue,omitempty"`
	NewValue    string `json:"newValue,omitempty"`
	Description string `json:"description"`
}

// SchemaField represents a field in a schema with cardinality tracking
type SchemaField struct {
	Name              string    `json:"name"`
	Type              ValueType `json:"type"`
	Source            string    `json:"source"`
	IsHighCardinality bool      `json:"isHighCardinality"`
	Example           string    `json:"example,omitempty"`
	LastUpdated       time.Time `json:"lasUpdated"`
}

// BaseSchema represents the common fields for all schema types
type BaseSchema struct {
	SchemaID     string        `json:"schemaId"`
	SignalType   string        `json:"signalType"`
	ScopeName    string        `json:"scopeName"`
	ScopeVersion string        `json:"scopeVersion"`
	SchemaURL    string        `json:"schemaUrl"`
	Fields       []SchemaField `json:"fields"`
	SeenCount    int           `json:"seenCount"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
	Version      SchemaVersion `json:"version"`
	Producers    []Producer    `json:"producers"`
	Consumers    []Consumer    `json:"consumers"`
}

// Producer represents a service that produces this schema
type Producer struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
}

// Consumer represents a service that consumes this schema
type Consumer struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	FirstUsed time.Time `json:"firstUsed"`
	LastUsed  time.Time `json:"lastUsed"`
}

// MetricSchema represents a metric schema
type MetricSchema struct {
	BaseSchema
	SignalKey   string `json:"signalKey"` // metric name
	MetricType  string `json:"metricType"`
	Unit        string `json:"unit"`
	IsMonotonic bool   `json:"isMonotonic"`
	Temporality string `json:"temporality"`
}

// LogSchema represents a log schema
type LogSchema struct {
	BaseSchema
	SignalKey   string `json:"signalKey"` // log identifier/name
	BodyType    string `json:"bodyType"`
	HasSeverity bool   `json:"hasSeverity"`
}

// TraceSchema represents a trace schema
type TraceSchema struct {
	BaseSchema
	SignalKey string `json:"signalKey"` // span operation
	SpanKind  string `json:"spanKind"`
	HasStatus bool   `json:"hasStatus"`
	HasEvents bool   `json:"hasEvents"`
}

// ValueType represents the type of a field value
type ValueType string

const (
	ValueTypeString ValueType = "String"
	ValueTypeInt    ValueType = "Int"
	ValueTypeDouble ValueType = "Double"
	ValueTypeBool   ValueType = "Bool"
	ValueTypeMap    ValueType = "Map"
	ValueTypeArray  ValueType = "Array"
)

// processAttributes processes a single set of attributes in one pass, extracting types and tracking cardinality.
// It is used for processing attributes from a single source (e.g., resource attributes, scope attributes).
// The function:
//   - Creates a new map to store attribute types
//   - Processes all attributes in the source
//   - Tracks cardinality for every attribute
//   - Returns a new map containing all processed attributes
func processAttributes(attrs pcommon.Map, tracker *CardinalityTracker) map[string]ValueType {
	result := make(map[string]ValueType, attrs.Len())
	attrs.Range(func(k string, v pcommon.Value) bool {
		// Track cardinality
		tracker.TrackValue(k, v.AsString())
		// Store type
		result[k] = convertValueType(v.Type())
		return true
	})
	return result
}

// mergeProcessedAttributes merges attributes from multiple sources into a target map while avoiding duplicates.
// It is used when combining attributes from multiple sources (e.g., multiple metric data points).
// The function:
//   - Takes an existing map as the target
//   - Only processes attributes that don't exist in the target
//   - Only tracks cardinality for new attributes
//   - Modifies the target map in-place
//
// This is more efficient than processAttributes when dealing with multiple sources that may have overlapping attributes.
func mergeProcessedAttributes(target map[string]ValueType, source pcommon.Map, tracker *CardinalityTracker) {
	source.Range(func(k string, v pcommon.Value) bool {
		if _, exists := target[k]; !exists {
			// Track cardinality for new attributes
			tracker.TrackValue(k, v.AsString())
			target[k] = convertValueType(v.Type())
		}
		return true
	})
}

func ExtractLogSchemas(logs plog.Logs) []LogSchema {
	schemaMap := make(map[string]*LogSchema)
	tracker := NewCardinalityTracker(1000, 24*time.Hour)

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		rl := logs.ResourceLogs().At(i)
		resourceAttrs := processAttributes(rl.Resource().Attributes(), tracker)

		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			scopeName := sl.Scope().Name()
			scopeVersion := sl.Scope().Version()
			scopeAttrs := processAttributes(sl.Scope().Attributes(), tracker)

			for k := 0; k < sl.LogRecords().Len(); k++ {
				log := sl.LogRecords().At(k)
				fields := make([]SchemaField, 0)

				// Add resource attributes
				for name, attrType := range resourceAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "resource",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Add scope attributes
				for name, attrType := range scopeAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "scope",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Add log attributes
				logAttrs := processAttributes(log.Attributes(), tracker)
				for name, attrType := range logAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "log",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Get schema URL from scope attributes
				schemaURL := ""
				sl.Scope().Attributes().Range(func(k string, v pcommon.Value) bool {
					if k == "schema_url" {
						schemaURL = v.Str()
						return false
					}
					return true
				})

				schema := &LogSchema{
					BaseSchema: BaseSchema{
						SignalType:   "log",
						ScopeName:    scopeName,
						ScopeVersion: scopeVersion,
						SchemaURL:    schemaURL,
						Fields:       fields,
						SeenCount:    1,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					},
					SignalKey:   "TBD", // TODO: add signal key
					BodyType:    string(convertValueType(log.Body().Type())),
					HasSeverity: log.SeverityNumber() != 0,
				}

				schema.SchemaID = generateSchemaID(schema)

				if existing, ok := schemaMap[schema.SchemaID]; ok {
					existing.SeenCount++
					existing.UpdatedAt = time.Now()
				} else {
					schemaMap[schema.SchemaID] = schema
				}
			}
		}
	}

	result := make([]LogSchema, 0, len(schemaMap))
	for _, schema := range schemaMap {
		result = append(result, *schema)
	}
	return result
}

func ExtractTraceSchemas(traces ptrace.Traces) []TraceSchema {
	schemaMap := make(map[string]*TraceSchema)
	tracker := NewCardinalityTracker(1000, 24*time.Hour)

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		resourceAttrs := processAttributes(rs.Resource().Attributes(), tracker)

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			scopeName := ss.Scope().Name()
			scopeVersion := ss.Scope().Version()
			scopeAttrs := processAttributes(ss.Scope().Attributes(), tracker)

			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)
				fields := make([]SchemaField, 0)

				// Add resource attributes
				for name, attrType := range resourceAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "resource",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Add scope attributes
				for name, attrType := range scopeAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "scope",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Add span attributes
				spanAttrs := processAttributes(span.Attributes(), tracker)
				for name, attrType := range spanAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "span",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Get schema URL from scope attributes
				schemaURL := ""
				ss.Scope().Attributes().Range(func(k string, v pcommon.Value) bool {
					if k == "schema_url" {
						schemaURL = v.Str()
						return false
					}
					return true
				})

				schema := &TraceSchema{
					BaseSchema: BaseSchema{
						SignalType:   "trace",
						ScopeName:    scopeName,
						ScopeVersion: scopeVersion,
						SchemaURL:    schemaURL,
						Fields:       fields,
						SeenCount:    1,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					},
					SignalKey: span.Name(),
					SpanKind:  span.Kind().String(),
					HasStatus: span.Status().Code() != 0,
					HasEvents: span.Events().Len() > 0,
				}

				schema.SchemaID = generateSchemaID(schema)

				if existing, ok := schemaMap[schema.SchemaID]; ok {
					existing.SeenCount++
					existing.UpdatedAt = time.Now()
				} else {
					schemaMap[schema.SchemaID] = schema
				}
			}
		}
	}

	result := make([]TraceSchema, 0, len(schemaMap))
	for _, schema := range schemaMap {
		result = append(result, *schema)
	}
	return result
}

func ExtractMetricSchema(metrics pmetric.Metrics) []MetricSchema {
	schemaMap := make(map[string]*MetricSchema)
	tracker := NewCardinalityTracker(1000, 24*time.Hour)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		resourceAttrs := processAttributes(rm.Resource().Attributes(), tracker)

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scopeName := sm.Scope().Name()
			scopeVersion := sm.Scope().Version()
			scopeAttrs := processAttributes(sm.Scope().Attributes(), tracker)

			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				fields := make([]SchemaField, 0)

				// Add resource attributes
				for name, attrType := range resourceAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "resource",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Add scope attributes
				for name, attrType := range scopeAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "scope",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				// Process data point attributes
				dpAttrs := make(map[string]ValueType)
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					for i := 0; i < m.Gauge().DataPoints().Len(); i++ {
						mergeProcessedAttributes(dpAttrs, m.Gauge().DataPoints().At(i).Attributes(), tracker)
					}
				case pmetric.MetricTypeSum:
					for i := 0; i < m.Sum().DataPoints().Len(); i++ {
						mergeProcessedAttributes(dpAttrs, m.Sum().DataPoints().At(i).Attributes(), tracker)
					}
				case pmetric.MetricTypeHistogram:
					for i := 0; i < m.Histogram().DataPoints().Len(); i++ {
						mergeProcessedAttributes(dpAttrs, m.Histogram().DataPoints().At(i).Attributes(), tracker)
					}
				case pmetric.MetricTypeExponentialHistogram:
					for i := 0; i < m.ExponentialHistogram().DataPoints().Len(); i++ {
						mergeProcessedAttributes(dpAttrs, m.ExponentialHistogram().DataPoints().At(i).Attributes(), tracker)
					}
				case pmetric.MetricTypeSummary:
					for i := 0; i < m.Summary().DataPoints().Len(); i++ {
						mergeProcessedAttributes(dpAttrs, m.Summary().DataPoints().At(i).Attributes(), tracker)
					}
				}

				// Add data point attributes
				for name, attrType := range dpAttrs {
					fields = append(fields, SchemaField{
						Name:              InternField(name),
						Type:              attrType,
						Source:            "datapoint",
						IsHighCardinality: tracker.IsHighCardinality(name),
						LastUpdated:       time.Now(),
					})
				}

				schema := &MetricSchema{
					BaseSchema: BaseSchema{
						SignalType:   "metric",
						ScopeName:    scopeName,
						ScopeVersion: scopeVersion,
						SchemaURL:    sm.SchemaUrl(),
						Fields:       fields,
						SeenCount:    1,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					},
					SignalKey:   m.Name(),
					MetricType:  m.Type().String(),
					Unit:        m.Unit(),
					IsMonotonic: isMonotonic(m),
					Temporality: convertTemporality(m),
				}

				schema.SchemaID = generateSchemaID(schema)

				slog.Debug("extracted schema", "schema_id", schema.SchemaID, "metric", m.Name())

				if existing, ok := schemaMap[schema.SchemaID]; ok {
					existing.SeenCount++
					existing.UpdatedAt = time.Now()
				} else {
					schemaMap[schema.SchemaID] = schema
				}
			}
		}
	}

	result := make([]MetricSchema, 0, len(schemaMap))
	for _, schema := range schemaMap {
		result = append(result, *schema)
	}
	return result
}

func convertValueType(t pcommon.ValueType) ValueType {
	switch t {
	case pcommon.ValueTypeStr:
		return ValueTypeString
	case pcommon.ValueTypeInt:
		return ValueTypeInt
	case pcommon.ValueTypeDouble:
		return ValueTypeDouble
	case pcommon.ValueTypeBool:
		return ValueTypeBool
	case pcommon.ValueTypeMap:
		return ValueTypeMap
	case pcommon.ValueTypeSlice:
		return ValueTypeArray
	default:
		return ValueTypeString
	}
}

func isMonotonic(m pmetric.Metric) bool {
	return m.Type() == pmetric.MetricTypeSum && m.Sum().IsMonotonic()
}

func convertTemporality(m pmetric.Metric) string {
	switch m.Type() {
	case pmetric.MetricTypeSum:
		return m.Sum().AggregationTemporality().String()
	case pmetric.MetricTypeHistogram:
		return m.Histogram().AggregationTemporality().String()
	case pmetric.MetricTypeExponentialHistogram:
		return m.ExponentialHistogram().AggregationTemporality().String()
	default:
		return ""
	}
}

// SchemaIDGenerator is an interface for types that can generate schema IDs
type SchemaIDGenerator interface {
	GetSchemaID() string
	SetSchemaID(id string)
}

// Ensure our schema types implement SchemaIDGenerator
func (s *MetricSchema) GetSchemaID() string   { return s.SchemaID }
func (s *MetricSchema) SetSchemaID(id string) { s.SchemaID = id }
func (s *LogSchema) GetSchemaID() string      { return s.SchemaID }
func (s *LogSchema) SetSchemaID(id string)    { s.SchemaID = id }
func (s *TraceSchema) GetSchemaID() string    { return s.SchemaID }
func (s *TraceSchema) SetSchemaID(id string)  { s.SchemaID = id }

// generateSchemaID generates a unique ID for any schema type
func generateSchemaID(schema SchemaIDGenerator) string {
	return generateSchemaIDWithXXHash(schema)
}

// generateSchemaIDWithXXHash generates a schema ID using xxHash
func generateSchemaIDWithXXHash(schema SchemaIDGenerator) string {
	// Create a map of fields to ensure consistent ordering
	fieldMap := make(map[string]interface{})

	// Add common fields
	fieldMap["signal_type"] = getSignalType(schema)
	fieldMap["signal_key"] = getSignalKey(schema)
	fieldMap["scope_name"] = getScopeName(schema)
	fieldMap["scope_version"] = getScopeVersion(schema)

	// Add type-specific fields
	switch s := schema.(type) {
	case *MetricSchema:
		fieldMap["metric_type"] = s.MetricType
		fieldMap["unit"] = s.Unit
		fieldMap["is_monotonic"] = s.IsMonotonic
		fieldMap["temporality"] = s.Temporality
	case *LogSchema:
		fieldMap["body_type"] = s.BodyType
		fieldMap["has_severity"] = s.HasSeverity
	case *TraceSchema:
		fieldMap["span_kind"] = s.SpanKind
		fieldMap["has_status"] = s.HasStatus
		fieldMap["has_events"] = s.HasEvents
	}

	// Add fields, but only include structural information
	fields := getFields(schema)
	structuralFields := make([]map[string]interface{}, len(fields))
	for i, field := range fields {
		structuralFields[i] = map[string]interface{}{
			"name":                field.Name,
			"type":                field.Type,
			"source":              field.Source,
			"is_high_cardinality": field.IsHighCardinality,
		}
	}
	sort.Slice(structuralFields, func(i, j int) bool {
		return structuralFields[i]["name"].(string) < structuralFields[j]["name"].(string)
	})
	fieldMap["fields"] = structuralFields

	// Convert to JSON for hashing
	data, err := json.Marshal(fieldMap)
	if err != nil {
		// Fallback to timestamp-based ID if marshaling fails
		return fmt.Sprintf("schema_%d", time.Now().UnixNano())
	}

	// Generate hash
	h := xxhash.New()
	h.Write(data)
	return fmt.Sprintf("schema_%x", h.Sum64())
}

// Helper functions to get common fields
func getSignalType(s SchemaIDGenerator) string {
	switch s.(type) {
	case *MetricSchema:
		return "metric"
	case *LogSchema:
		return "log"
	case *TraceSchema:
		return "trace"
	default:
		return "unknown"
	}
}

func getScopeName(s SchemaIDGenerator) string {
	switch s := s.(type) {
	case *MetricSchema:
		return s.ScopeName
	case *LogSchema:
		return s.ScopeName
	case *TraceSchema:
		return s.ScopeName
	default:
		return ""
	}
}

func getScopeVersion(s SchemaIDGenerator) string {
	switch s := s.(type) {
	case *MetricSchema:
		return s.ScopeVersion
	case *LogSchema:
		return s.ScopeVersion
	case *TraceSchema:
		return s.ScopeVersion
	default:
		return ""
	}
}

func getFields(s SchemaIDGenerator) []SchemaField {
	switch s := s.(type) {
	case *MetricSchema:
		return s.Fields
	case *LogSchema:
		return s.Fields
	case *TraceSchema:
		return s.Fields
	default:
		return nil
	}
}

// Helper function to get signal key
func getSignalKey(s SchemaIDGenerator) string {
	switch s := s.(type) {
	case *MetricSchema:
		return s.SignalKey
	case *LogSchema:
		return s.SignalKey
	case *TraceSchema:
		return s.SignalKey
	default:
		return ""
	}
}
