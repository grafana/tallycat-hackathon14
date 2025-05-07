package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type SchemaField struct {
	Name              string
	Type              string
	Source            string
	IsHighCardinality bool
	Example           string
}

type BaseSchema struct {
	SchemaID     string
	SignalType   string
	ScopeName    string
	ScopeVersion string
	Fields       []SchemaField
	SeenCount    int
}

type MetricSchema struct {
	BaseSchema
	MetricType  string
	Unit        string
	IsMonotonic bool
	Temporality string
}

type LogSchema struct {
	BaseSchema
	BodyType    string
	HasSeverity bool
}

type TraceSchema struct {
	BaseSchema
	SpanKind  string
	HasStatus bool
	HasEvents bool
}

func ExtractLogSchemas(logs plog.Logs) []LogSchema {
	// TODO: Implement log schema extraction
	return nil
}

func ExtractTraceSchemas(traces ptrace.Traces) []TraceSchema {
	// TODO: Implement trace schema extraction
	return nil
}

func ExtractMetricSchema(metrics pmetric.Metrics) []MetricSchema {
	schemaMap := make(map[string]*MetricSchema)

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		resourceAttrs := flattenAttributeTypes(rm.Resource().Attributes())

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scopeName := sm.Scope().Name()
			scopeVersion := sm.Scope().Version()
			scopeAttrs := flattenAttributeTypes(sm.Scope().Attributes())

			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)

				fields := make([]SchemaField, 0)

				for name, attrType := range resourceAttrs {
					fields = append(fields, SchemaField{
						Name:              name,
						Type:              string(attrType),
						Source:            "resource",
						IsHighCardinality: isHighCardinality(name, attrType, "resource"),
					})
				}

				for name, attrType := range scopeAttrs {
					fields = append(fields, SchemaField{
						Name:              name,
						Type:              string(attrType),
						Source:            "scope",
						IsHighCardinality: isHighCardinality(name, attrType, "scope"),
					})
				}

				dpAttrs := extractDataPointAttributeTypes(m)
				for name, attrType := range dpAttrs {
					fields = append(fields, SchemaField{
						Name:              name,
						Type:              string(attrType),
						Source:            "datapoint",
						IsHighCardinality: isHighCardinality(name, attrType, "datapoint"),
					})
				}

				schema := &MetricSchema{
					BaseSchema: BaseSchema{
						SignalType:   "metric",
						ScopeName:    scopeName,
						ScopeVersion: scopeVersion,
						Fields:       fields,
						SeenCount:    1,
					},
					MetricType:  string(convertMetricType(m.Type())),
					Unit:        m.Unit(),
					IsMonotonic: isMonotonic(m),
					Temporality: string(convertTemporality(m)),
				}

				schema.SchemaID = generateSchemaID(schema)

				slog.Debug("extracted schema", "schema_id", schema.SchemaID, "metric", m.Name())

				if existing, ok := schemaMap[schema.SchemaID]; ok {
					existing.SeenCount++
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

func flattenAttributeTypes(attrs pcommon.Map) map[string]ValueType {
	result := make(map[string]ValueType, attrs.Len())
	attrs.Range(func(k string, v pcommon.Value) bool {
		result[k] = convertValueType(v.Type())
		return true
	})
	return result
}

func extractDataPointAttributeTypes(m pmetric.Metric) map[string]ValueType {
	attrs := make(map[string]ValueType)
	switch m.Type() {
	case pmetric.MetricTypeGauge:
		for i := 0; i < m.Gauge().DataPoints().Len(); i++ {
			mergeAttributeTypes(attrs, m.Gauge().DataPoints().At(i).Attributes())
		}
	case pmetric.MetricTypeSum:
		for i := 0; i < m.Sum().DataPoints().Len(); i++ {
			mergeAttributeTypes(attrs, m.Sum().DataPoints().At(i).Attributes())
		}
	case pmetric.MetricTypeHistogram:
		for i := 0; i < m.Histogram().DataPoints().Len(); i++ {
			mergeAttributeTypes(attrs, m.Histogram().DataPoints().At(i).Attributes())
		}
	case pmetric.MetricTypeExponentialHistogram:
		for i := 0; i < m.ExponentialHistogram().DataPoints().Len(); i++ {
			mergeAttributeTypes(attrs, m.ExponentialHistogram().DataPoints().At(i).Attributes())
		}
	case pmetric.MetricTypeSummary:
		for i := 0; i < m.Summary().DataPoints().Len(); i++ {
			mergeAttributeTypes(attrs, m.Summary().DataPoints().At(i).Attributes())
		}
	}
	return attrs
}

func mergeAttributeTypes(target map[string]ValueType, source pcommon.Map) {
	source.Range(func(k string, v pcommon.Value) bool {
		if _, exists := target[k]; !exists {
			target[k] = convertValueType(v.Type())
		}
		return true
	})
}

type ValueType string

const (
	ValueTypeString ValueType = "String"
	ValueTypeInt    ValueType = "Int"
	ValueTypeDouble ValueType = "Double"
	ValueTypeBool   ValueType = "Bool"
	ValueTypeMap    ValueType = "Map"
	ValueTypeArray  ValueType = "Array"
)

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

func convertMetricType(t pmetric.MetricType) string {
	switch t {
	case pmetric.MetricTypeGauge:
		return "Gauge"
	case pmetric.MetricTypeSum:
		return "Sum"
	case pmetric.MetricTypeHistogram:
		return "Histogram"
	case pmetric.MetricTypeExponentialHistogram:
		return "ExponentialHistogram"
	case pmetric.MetricTypeSummary:
		return "Summary"
	default:
		return "Gauge"
	}
}

func isMonotonic(m pmetric.Metric) bool {
	return m.Type() == pmetric.MetricTypeSum && m.Sum().IsMonotonic()
}

func convertTemporality(m pmetric.Metric) string {
	switch m.Type() {
	case pmetric.MetricTypeSum:
		return fromTemporality(m.Sum().AggregationTemporality())
	case pmetric.MetricTypeHistogram:
		return fromTemporality(m.Histogram().AggregationTemporality())
	case pmetric.MetricTypeExponentialHistogram:
		return fromTemporality(m.ExponentialHistogram().AggregationTemporality())
	default:
		return "unspecified"
	}
}

func fromTemporality(t pmetric.AggregationTemporality) string {
	switch t {
	case pmetric.AggregationTemporalityDelta:
		return "delta"
	case pmetric.AggregationTemporalityCumulative:
		return "cumulative"
	default:
		return "unspecified"
	}
}

func generateSchemaID(schema *MetricSchema) string {
	var sb strings.Builder
	sb.WriteString(schema.SignalType)
	sb.WriteString("|")
	sb.WriteString(schema.ScopeName)
	sb.WriteString("|")
	sb.WriteString(schema.MetricType)
	sb.WriteString("|")
	sb.WriteString(schema.Unit)
	sb.WriteString("|")
	sb.WriteString(schema.Temporality)
	sb.WriteString("|")

	// Sort by field name
	sort.Slice(schema.Fields, func(i, j int) bool {
		return schema.Fields[i].Name < schema.Fields[j].Name
	})

	for _, f := range schema.Fields {
		sb.WriteString(f.Name)
		sb.WriteString(":")
		sb.WriteString(f.Type)
		sb.WriteString(":")
		sb.WriteString(f.Source)
		sb.WriteString("|")
	}

	h := sha256.New()
	h.Write([]byte(sb.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// isHighCardinality uses heuristics to determine if a field is high cardinality.
func isHighCardinality(name string, t ValueType, source string) bool {
	lc := strings.ToLower(name)
	if strings.Contains(lc, "id") || strings.Contains(lc, "uuid") {
		return true
	}
	if lc == "trace_id" || lc == "span_id" {
		return true
	}
	if source == "datapoint" && t == ValueTypeString {
		return true
	}
	return false
}
