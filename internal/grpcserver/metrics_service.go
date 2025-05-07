package grpcserver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

type MetricsServiceServer struct {
	metricspb.UnimplementedMetricsServiceServer
	schemaRepo repository.SchemaProvider
	logger     *slog.Logger
}

func NewMetricsServiceServer(schemaRepo repository.SchemaProvider, logger *slog.Logger) *MetricsServiceServer {
	return &MetricsServiceServer{
		schemaRepo: schemaRepo,
		logger:     logger,
	}
}

func (s *MetricsServiceServer) Export(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
	metrics := pmetric.NewMetrics()
	rms := metrics.ResourceMetrics()
	rms.EnsureCapacity(len(req.ResourceMetrics))

	for _, rm := range req.ResourceMetrics {
		resourceMetric := rms.AppendEmpty()
		resourceMetric.SetSchemaUrl(rm.SchemaUrl)

		// Convert resource attributes
		if rm.Resource != nil {
			for _, attr := range rm.Resource.Attributes {
				resourceMetric.Resource().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
			}
		}

		// Convert scope metrics
		sms := resourceMetric.ScopeMetrics()
		sms.EnsureCapacity(len(rm.ScopeMetrics))

		for _, sm := range rm.ScopeMetrics {
			scopeMetric := sms.AppendEmpty()
			scopeMetric.SetSchemaUrl(sm.SchemaUrl)

			// Convert scope
			if sm.Scope != nil {
				scopeMetric.Scope().SetName(sm.Scope.Name)
				scopeMetric.Scope().SetVersion(sm.Scope.Version)
			}

			// Convert metrics
			ms := scopeMetric.Metrics()
			ms.EnsureCapacity(len(sm.Metrics))

			for _, m := range sm.Metrics {
				metric := ms.AppendEmpty()
				metric.SetName(m.Name)
				metric.SetDescription(m.Description)
				metric.SetUnit(m.Unit)

				// Convert data points based on metric type
				switch m.Data.(type) {
				case *metricpb.Metric_Gauge:
					gauge := metric.SetEmptyGauge()
					gdps := gauge.DataPoints()
					gdps.EnsureCapacity(len(m.GetGauge().DataPoints))

					for _, dp := range m.GetGauge().DataPoints {
						dataPoint := gdps.AppendEmpty()
						dataPoint.SetTimestamp(pcommon.Timestamp(dp.TimeUnixNano))
						dataPoint.SetDoubleValue(dp.GetAsDouble())

						// Convert data point attributes
						for _, attr := range dp.Attributes {
							dataPoint.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
						}
					}

				case *metricpb.Metric_Sum:
					sum := metric.SetEmptySum()
					sum.SetIsMonotonic(m.GetSum().IsMonotonic)
					sum.SetAggregationTemporality(pmetric.AggregationTemporality(m.GetSum().AggregationTemporality))

					sdps := sum.DataPoints()
					sdps.EnsureCapacity(len(m.GetSum().DataPoints))

					for _, dp := range m.GetSum().DataPoints {
						dataPoint := sdps.AppendEmpty()
						dataPoint.SetTimestamp(pcommon.Timestamp(dp.TimeUnixNano))
						dataPoint.SetDoubleValue(dp.GetAsDouble())

						// Convert data point attributes
						for _, attr := range dp.Attributes {
							dataPoint.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
						}
					}

				case *metricpb.Metric_Histogram:
					hist := metric.SetEmptyHistogram()
					hist.SetAggregationTemporality(pmetric.AggregationTemporality(m.GetHistogram().AggregationTemporality))

					hdps := hist.DataPoints()
					hdps.EnsureCapacity(len(m.GetHistogram().DataPoints))

					for _, dp := range m.GetHistogram().DataPoints {
						dataPoint := hdps.AppendEmpty()
						dataPoint.SetTimestamp(pcommon.Timestamp(dp.TimeUnixNano))
						dataPoint.SetCount(dp.Count)
						if dp.Sum != nil {
							dataPoint.SetSum(*dp.Sum)
						}
						dataPoint.BucketCounts().FromRaw(dp.BucketCounts)
						dataPoint.ExplicitBounds().FromRaw(dp.ExplicitBounds)

						// Convert data point attributes
						for _, attr := range dp.Attributes {
							dataPoint.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
						}
					}

				case *metricpb.Metric_ExponentialHistogram:
					hist := metric.SetEmptyExponentialHistogram()
					hist.SetAggregationTemporality(pmetric.AggregationTemporality(m.GetExponentialHistogram().AggregationTemporality))

					hdps := hist.DataPoints()
					hdps.EnsureCapacity(len(m.GetExponentialHistogram().DataPoints))

					for _, dp := range m.GetExponentialHistogram().DataPoints {
						dataPoint := hdps.AppendEmpty()
						dataPoint.SetTimestamp(pcommon.Timestamp(dp.TimeUnixNano))
						dataPoint.SetCount(dp.Count)
						if dp.Sum != nil {
							dataPoint.SetSum(*dp.Sum)
						}
						dataPoint.SetScale(dp.Scale)
						dataPoint.SetZeroCount(dp.ZeroCount)

						// Convert data point attributes
						for _, attr := range dp.Attributes {
							dataPoint.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
						}
					}

				case *metricpb.Metric_Summary:
					summ := metric.SetEmptySummary()

					sdps := summ.DataPoints()
					sdps.EnsureCapacity(len(m.GetSummary().DataPoints))

					for _, dp := range m.GetSummary().DataPoints {
						dataPoint := sdps.AppendEmpty()
						dataPoint.SetTimestamp(pcommon.Timestamp(dp.TimeUnixNano))
						dataPoint.SetCount(dp.Count)
						dataPoint.SetSum(dp.Sum)

						// Convert data point attributes
						for _, attr := range dp.Attributes {
							dataPoint.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
						}
					}
				}
			}
		}
	}

	// Extract schemas from the converted metrics
	schemas := schema.ExtractMetricSchema(metrics)

	// Register each schema
	for _, metricSchema := range schemas {
		// Convert to repository schema
		repoSchema := &repository.Schema{
			SchemaID:         metricSchema.SchemaID,
			SignalType:       metricSchema.SignalType,
			SignalKey:        metricSchema.SignalKey,
			ScopeName:        metricSchema.ScopeName,
			ScopeVersion:     metricSchema.ScopeVersion,
			SchemaURL:        metricSchema.SchemaURL,
			MetricType:       &metricSchema.MetricType,
			Unit:             &metricSchema.Unit,
			FieldNames:       make([]string, len(metricSchema.Fields)),
			FieldTypes:       make(map[string]string, len(metricSchema.Fields)),
			FieldSources:     make(map[string]string, len(metricSchema.Fields)),
			FieldCardinality: make(map[string]bool, len(metricSchema.Fields)),
			SeenCount:        metricSchema.SeenCount,
			CreatedAt:        metricSchema.CreatedAt,
			UpdatedAt:        metricSchema.UpdatedAt,
		}

		// Convert fields
		for i, field := range metricSchema.Fields {
			repoSchema.FieldNames[i] = field.Name
			repoSchema.FieldTypes[field.Name] = string(field.Type)
			repoSchema.FieldSources[field.Name] = field.Source
			repoSchema.FieldCardinality[field.Name] = field.IsHighCardinality
		}

		// Register schema
		if err := s.schemaRepo.RegisterSchema(ctx, repoSchema); err != nil {
			s.logger.Error("failed to register schema",
				"error", err,
				"schema_id", metricSchema.SchemaID)
			return nil, err
		}
	}

	return &metricspb.ExportMetricsServiceResponse{}, nil
}
