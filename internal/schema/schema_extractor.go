package schema

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// generateTelemetrySchemaID creates a deterministic schema ID based on telemetry characteristics.
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
func generateTelemetrySchemaID(telemetry Telemetry) string {
	attributeNames := make([]string, 0, len(telemetry.Attributes))
	for _, attr := range telemetry.Attributes {
		attributeNames = append(attributeNames, attr.Name)
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

				if metric.Name() == "otelcol_exporter_queue_size" || metric.Name() == "otelcol.exporter.queue.capacity" {
					slog.Info("metric", "metric", metric)
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
					Producers:         make(map[string]*Producer),
				}

				producer := &Producer{
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				}

				resourceAttributes.Range(func(key string, value pcommon.Value) bool {
					switch key {
					case "service.name":
						producer.Name = value.Str()
					case "service.namespace":
						producer.Namespace = value.Str()
					case "service.version":
						producer.Version = value.Str()
					case "service.instance.id":
						producer.InstanceID = value.Str()
					}

					telemetry.Attributes = append(telemetry.Attributes, Attribute{
						Name:   key,
						Type:   AttributeType(value.Type().String()),
						Source: AttributeSourceResource,
					})
					return true
				})

				// Add producer if it has a name
				if producer.Name != "" {
					telemetry.Producers[producer.ProducerID()] = producer
				}

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

				telemetry.SchemaID = generateTelemetrySchemaID(telemetry)
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
