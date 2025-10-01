package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/ptrace"
	profilepb "go.opentelemetry.io/proto/otlp/profiles/v1development"
)

// generateMetricSchemaID creates a deterministic schema ID based on telemetry characteristics.
// The schema ID is a unique identifier that represents the structure of a telemetry object,
// including its metric name, unit, type, temporality, and all its attributes.
//
// This is useful for:
// - Identifying identical telemetry schemas across different services
// - Tracking schema evolution over time
// - Enabling schema-based routing and processing
// - Supporting schema versioning and compatibility checks
//
// The ID is generated using xxhash for performance, and is deterministic
// meaning the same telemetry structure will always produce the same ID.
func generateMetricSchemaID(telemetry Telemetry) string {
	attributeNames := make([]string, 0, len(telemetry.Attributes))
	for _, attr := range telemetry.Attributes {
		// Accorting to the spec, only data point attributes are part of the schema
		// which means we will only detect drift if the telemetry attributes change
		if attr.Source == AttributeSourceDataPoint {
			attributeNames = append(attributeNames, attr.Name)
		}
	}
	sort.Strings(attributeNames)

	parts := []string{
		telemetry.SchemaKey,
		telemetry.MetricUnit,
		string(telemetry.MetricType),
		string(telemetry.MetricTemporality),
		strings.Join(attributeNames, ","),
	}

	h := xxhash.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}

func ExtractFromMetrics(metrics pmetric.Metrics) []Telemetry {
	telemetries := map[string]Telemetry{}

	for i := range metrics.ResourceMetrics().Len() {
		resourceMetric := metrics.ResourceMetrics().At(i)
		resourceAttributes := resourceMetric.Resource().Attributes()

		for k := range resourceMetric.ScopeMetrics().Len() {
			scopeMetric := resourceMetric.ScopeMetrics().At(k)
			scopeAttributes := scopeMetric.Scope().Attributes()

			for l := range scopeMetric.Metrics().Len() {
				metric := scopeMetric.Metrics().At(l)
				metricAttributes := pcommon.Map{}
				metricTemporality := MetricTemporalityUnspecified
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					metricAttributes = metric.Gauge().DataPoints().At(0).Attributes()
				case pmetric.MetricTypeSum:
					metricAttributes = metric.Sum().DataPoints().At(0).Attributes()
					metricTemporality = MetricTemporality(metric.Sum().AggregationTemporality().String())
				case pmetric.MetricTypeHistogram:
					metricAttributes = metric.Histogram().DataPoints().At(0).Attributes()
					metricTemporality = MetricTemporality(metric.Histogram().AggregationTemporality().String())
				case pmetric.MetricTypeExponentialHistogram:
					metricAttributes = metric.ExponentialHistogram().DataPoints().At(0).Attributes()
					metricTemporality = MetricTemporality(metric.ExponentialHistogram().AggregationTemporality().String())
				case pmetric.MetricTypeSummary:
					metricAttributes = metric.Summary().DataPoints().At(0).Attributes()
				}

				telemetry := Telemetry{
					SchemaURL:         scopeMetric.SchemaUrl(),
					TelemetryType:     TelemetryTypeMetric,
					SchemaKey:         metric.Name(),
					MetricUnit:        metric.Unit(),
					MetricType:        MetricType(metric.Type().String()),
					MetricTemporality: metricTemporality,
					Brief:             metric.Description(),
					Note:              metric.Description(),
					Attributes:        make([]Attribute, 0, resourceAttributes.Len()+scopeAttributes.Len()+metricAttributes.Len()),
					Protocol:          TelemetryProtocolOTLP,
					SeenCount:         1,
					CreatedAt:         time.Now(),
					UpdatedAt:         time.Now(),
					Entities:          make(map[string]*Entity),
				}

				// Extract entities from resource attributes
				entities := DetectEntities(resourceAttributes)
				for _, entity := range entities {
					telemetry.Entities[entity.ID] = &entity
				}

				scope := DetectScopes(scopeMetric.Scope(), scopeMetric.SchemaUrl())
				telemetry.Scope = &scope

				resourceAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceResource,
					})
					return true
				})

				scopeAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceScope,
					})
					return true
				})

				metricAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceDataPoint,
					})
					return true
				})

				telemetry.SchemaID = generateMetricSchemaID(telemetry)
				if existing, ok := telemetries[telemetry.SchemaID]; ok {
					existing.SeenCount++
				} else {
					telemetries[telemetry.SchemaID] = telemetry
				}
			}
		}
	}

	result := make([]Telemetry, 0, len(telemetries))
	for _, telemetry := range telemetries {
		result = append(result, telemetry)
	}

	return result
}

func generateLogSchemaID(telemetry Telemetry) string {
	attributeNames := make([]string, 0, len(telemetry.Attributes))
	for _, attr := range telemetry.Attributes {
		// Accorting to the spec, only data point attributes are part of the schema
		// which means we will only detect drift if the telemetry attributes change
		if attr.Source == AttributeSourceDataPoint {
			attributeNames = append(attributeNames, attr.Name)
		}
	}
	sort.Strings(attributeNames)

	parts := []string{
		telemetry.SchemaKey,
		strconv.Itoa(telemetry.LogSeverityNumber),
		telemetry.LogSeverityText,
		telemetry.LogBody,
		strconv.Itoa(telemetry.LogFlags),
		telemetry.LogTraceID,
		telemetry.LogSpanID,
		telemetry.LogEventName,
		strconv.Itoa(telemetry.LogDroppedAttributesCount),
		strings.Join(attributeNames, ","),
	}

	h := xxhash.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}

func generateTraceSchemaID(telemetry Telemetry) string {
	attributeNames := make([]string, 0, len(telemetry.Attributes))
	for _, attr := range telemetry.Attributes {
		if attr.Source == AttributeSourceSpan {
			attributeNames = append(attributeNames, attr.Name)
		}
	}
	sort.Strings(attributeNames)

	parts := []string{
		telemetry.SchemaKey,
		string(telemetry.SpanKind),
		telemetry.SpanName,
		telemetry.SpanTraceID,
		strings.Join(attributeNames, ","),
	}

	h := xxhash.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}

func ExtractFromLogs(logs plog.Logs) []Telemetry {
	telemetries := map[string]Telemetry{}

	for i := range logs.ResourceLogs().Len() {
		resourceLog := logs.ResourceLogs().At(i)
		resourceAttributes := resourceLog.Resource().Attributes()

		for k := range resourceLog.ScopeLogs().Len() {
			scopeLog := resourceLog.ScopeLogs().At(k)
			scopeAttributes := scopeLog.Scope().Attributes()

			for l := range scopeLog.LogRecords().Len() {
				logRecord := scopeLog.LogRecords().At(l)
				logAttributes := logRecord.Attributes()

				// Determine schema key: use event name if available, otherwise use service name from attributes
				schemaKey := logRecord.EventName()
				if schemaKey == "" {
					// Try to get message from log attributes
					if serviceAttr, exists := logRecord.Attributes().Get("message"); exists {
						schemaKey = serviceAttr.Str()
					} else if serviceAttr, exists := logRecord.Attributes().Get("msg"); exists {
						schemaKey = serviceAttr.Str()
					} else if logRecord.Body().AsString() != "" {
						schemaKey = logRecord.Body().AsString()
					} else {
						// Fallback to a generic log schema key
						schemaKey = "application_log"
					}
				}

				telemetry := Telemetry{
					SchemaURL:                 scopeLog.SchemaUrl(),
					TelemetryType:             TelemetryTypeLog,
					SchemaKey:                 schemaKey,
					LogSeverityNumber:         int(logRecord.SeverityNumber()),
					LogSeverityText:           logRecord.SeverityText(),
					LogBody:                   logRecord.Body().AsString(),
					LogFlags:                  int(logRecord.Flags()),
					LogTraceID:                logRecord.TraceID().String(),
					LogSpanID:                 logRecord.SpanID().String(),
					LogEventName:              logRecord.EventName(),
					LogDroppedAttributesCount: int(logRecord.DroppedAttributesCount()),
					Attributes:                make([]Attribute, 0, resourceAttributes.Len()+scopeAttributes.Len()+logAttributes.Len()),
					Protocol:                  TelemetryProtocolOTLP,
					SeenCount:                 1,
					CreatedAt:                 time.Now(),
					UpdatedAt:                 time.Now(),
					Entities:                  make(map[string]*Entity),
				}

				// Extract entities from resource attributes
				entities := DetectEntities(resourceAttributes)
				for _, entity := range entities {
					telemetry.Entities[entity.ID] = &entity
				}

				scope := DetectScopes(scopeLog.Scope(), scopeLog.SchemaUrl())
				telemetry.Scope = &scope

				resourceAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceResource,
					})
					return true
				})

				scopeAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceScope,
					})
					return true
				})

				logAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceLogRecord,
					})
					return true
				})

				telemetry.SchemaID = generateLogSchemaID(telemetry)
				if existing, ok := telemetries[telemetry.SchemaID]; ok {
					existing.SeenCount++
				} else {
					telemetries[telemetry.SchemaID] = telemetry
				}
			}
		}
	}

	result := make([]Telemetry, 0, len(telemetries))
	for _, telemetry := range telemetries {
		result = append(result, telemetry)
	}

	return result
}

func ExtractFromTraces(traces ptrace.Traces) []Telemetry {
	telemetries := map[string]Telemetry{}

	for i := range traces.ResourceSpans().Len() {
		resourceSpan := traces.ResourceSpans().At(i)
		resourceAttributes := resourceSpan.Resource().Attributes()

		for k := range resourceSpan.ScopeSpans().Len() {
			scopeSpan := resourceSpan.ScopeSpans().At(k)
			scopeAttributes := scopeSpan.Scope().Attributes()

			for l := range scopeSpan.Spans().Len() {
				span := scopeSpan.Spans().At(l)
				spanAttributes := span.Attributes()

				telemetry := Telemetry{
					SchemaURL:     scopeSpan.SchemaUrl(),
					TelemetryType: TelemetryTypeSpan,
					SchemaKey:     span.Name(),
					SpanKind:      SpanKind(span.Kind().String()),
					SpanName:      span.Name(),
					SpanTraceID:   span.TraceID().String(),
					Attributes:    make([]Attribute, 0, resourceAttributes.Len()+scopeAttributes.Len()+spanAttributes.Len()),
					Protocol:      TelemetryProtocolOTLP,
					SeenCount:     1,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
					Entities:      make(map[string]*Entity),
				}

				// Extract entities from resource attributes
				entities := DetectEntities(resourceAttributes)
				for _, entity := range entities {
					telemetry.Entities[entity.ID] = &entity
				}

				scope := DetectScopes(scopeSpan.Scope(), scopeSpan.SchemaUrl())
				telemetry.Scope = &scope

				resourceAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceResource,
					})
					return true
				})

				scopeAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceScope,
					})
					return true
				})

				spanAttributes.Range(func(key string, value pcommon.Value) bool {
					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceSpan,
					})
					return true
				})

				telemetry.SchemaID = generateTraceSchemaID(telemetry)
				if existing, ok := telemetries[telemetry.SchemaID]; ok {
					existing.SeenCount++
				} else {
					telemetries[telemetry.SchemaID] = telemetry
				}
			}
		}
	}

	result := make([]Telemetry, 0, len(telemetries))
	for _, telemetry := range telemetries {
		result = append(result, telemetry)
	}

	return result
}

func generateProfileSchemaID(telemetry Telemetry) string {
	attributeNames := make([]string, 0, len(telemetry.Attributes))
	for _, attr := range telemetry.Attributes {
		if attr.Source == AttributeSourceProfile {
			attributeNames = append(attributeNames, attr.Name)
		}
	}
	sort.Strings(attributeNames)

	parts := []string{
		telemetry.SchemaKey,
		telemetry.ProfileSampleAggregationTemporality,
		telemetry.ProfileSampleUnit,
		strings.Join(attributeNames, ","),
	}

	h := xxhash.New()
	h.Write([]byte(strings.Join(parts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}

func ExtractFromProfiles(profiles pprofile.Profiles, dictionary *profilepb.ProfilesDictionary) []Telemetry {
	telemetries := map[string]Telemetry{}

	for i := range profiles.ResourceProfiles().Len() {
		resourceProfile := profiles.ResourceProfiles().At(i)
		resourceAttributes := resourceProfile.Resource().Attributes()

		for k := range resourceProfile.ScopeProfiles().Len() {
			scopeProfile := resourceProfile.ScopeProfiles().At(k)
			scopeAttributes := scopeProfile.Scope().Attributes()

			for l := range scopeProfile.Profiles().Len() {
				profile := scopeProfile.Profiles().At(l)

				profileAttributes := pcommon.NewMap()
				if dictionary != nil && dictionary.AttributeTable != nil && dictionary.StringTable != nil {
					for _, attrIndex := range profile.AttributeIndices().All() {
						if int(attrIndex) < len(dictionary.AttributeTable) {
							attr := dictionary.AttributeTable[attrIndex]
							if int(attr.KeyStrindex) < len(dictionary.StringTable) {
								key := dictionary.StringTable[attr.KeyStrindex]
								profileAttributes.PutStr(key, attr.Value.GetStringValue())
							}
						}
					}
				}

				for _, s := range profile.SampleType().All() {
					// One OTLP Profile message can have multiple profile types. We'll store one telemetry for each profile type.
					var profileType, profileAggregationTemporality, profileUnit string
					if dictionary != nil && dictionary.StringTable != nil {
						if int(s.TypeStrindex()) < len(dictionary.StringTable) {
							profileType = dictionary.StringTable[s.TypeStrindex()]
						}
						if int(s.AggregationTemporality()) < len(dictionary.StringTable) {
							profileAggregationTemporality = dictionary.StringTable[s.AggregationTemporality()]
						}
						if int(s.UnitStrindex()) < len(dictionary.StringTable) {
							profileUnit = dictionary.StringTable[s.UnitStrindex()]
						}
					}
					telemetry := Telemetry{
						SchemaURL:                           scopeProfile.SchemaUrl(),
						TelemetryType:                       TelemetryTypeProfile,
						SchemaKey:                           profileType,
						ProfileSampleAggregationTemporality: profileAggregationTemporality,
						ProfileSampleUnit:                   profileUnit,
						Attributes:                          make([]Attribute, 0, resourceAttributes.Len()+scopeAttributes.Len()+profileAttributes.Len()),
						Protocol:                            TelemetryProtocolOTLP,
						SeenCount:                           1,
						CreatedAt:                           time.Now(),
						UpdatedAt:                           time.Now(),
						Entities:                            make(map[string]*Entity),
					}

					// Extract entities from resource attributes
					entities := DetectEntities(resourceAttributes)
					for _, entity := range entities {
						telemetry.Entities[entity.ID] = &entity
					}

					scope := DetectScopes(scopeProfile.Scope(), scopeProfile.SchemaUrl())
					telemetry.Scope = &scope

					resourceAttributes.Range(func(key string, value pcommon.Value) bool {
						telemetry.Attributes = append(telemetry.Attributes, Attribute{
							Name:   key,
							Type:   AttributeType(value.Type().String()),
							Source: AttributeSourceResource,
						})
						return true
					})

					scopeAttributes.Range(func(key string, value pcommon.Value) bool {
						telemetry.Attributes = append(telemetry.Attributes, Attribute{
							Name:   key,
							Type:   AttributeType(value.Type().String()),
							Source: AttributeSourceScope,
						})
						return true
					})

					profileAttributes.Range(func(key string, value pcommon.Value) bool {
						telemetry.Attributes = append(telemetry.Attributes, Attribute{
							Name:   key,
							Type:   AttributeType(value.Type().String()),
							Source: AttributeSourceSpan,
						})
						return true
					})

					telemetry.SchemaID = generateProfileSchemaID(telemetry)
					if existing, ok := telemetries[telemetry.SchemaID]; ok {
						existing.SeenCount++
					} else {
						telemetries[telemetry.SchemaID] = telemetry
					}
				}
			}
		}
	}
	result := make([]Telemetry, 0, len(telemetries))
	for _, telemetry := range telemetries {
		result = append(result, telemetry)
	}

	return result
}
