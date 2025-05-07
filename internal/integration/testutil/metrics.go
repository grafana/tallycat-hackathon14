package testutil

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// ConvertPmetricToRequest converts pmetric.Metrics to ExportMetricsServiceRequest
func ConvertPmetricToRequest(md pmetric.Metrics) *metricspb.ExportMetricsServiceRequest {
	request := &metricspb.ExportMetricsServiceRequest{
		ResourceMetrics: make([]*metricpb.ResourceMetrics, 0, md.ResourceMetrics().Len()),
	}

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		resourceMetrics := &metricpb.ResourceMetrics{
			Resource: &resourcepb.Resource{
				Attributes: convertAttributes(rm.Resource().Attributes()),
			},
			ScopeMetrics: make([]*metricpb.ScopeMetrics, 0, rm.ScopeMetrics().Len()),
			SchemaUrl:    rm.SchemaUrl(),
		}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scopeMetrics := &metricpb.ScopeMetrics{
				Scope: &commonpb.InstrumentationScope{
					Name:    sm.Scope().Name(),
					Version: sm.Scope().Version(),
				},
				Metrics:   make([]*metricpb.Metric, 0, sm.Metrics().Len()),
				SchemaUrl: sm.SchemaUrl(),
			}

			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				metric := &metricpb.Metric{
					Name:        m.Name(),
					Description: m.Description(),
					Unit:        m.Unit(),
				}

				switch m.Type() {
				case pmetric.MetricTypeGauge:
					gauge := m.Gauge()
					metric.Data = &metricpb.Metric_Gauge{
						Gauge: &metricpb.Gauge{
							DataPoints: convertNumberDataPoints(gauge.DataPoints()),
						},
					}
				case pmetric.MetricTypeSum:
					sum := m.Sum()
					metric.Data = &metricpb.Metric_Sum{
						Sum: &metricpb.Sum{
							DataPoints:             convertNumberDataPoints(sum.DataPoints()),
							IsMonotonic:            sum.IsMonotonic(),
							AggregationTemporality: metricpb.AggregationTemporality(sum.AggregationTemporality()),
						},
					}
				case pmetric.MetricTypeHistogram:
					hist := m.Histogram()
					metric.Data = &metricpb.Metric_Histogram{
						Histogram: &metricpb.Histogram{
							DataPoints:             convertHistogramDataPoints(hist.DataPoints()),
							AggregationTemporality: metricpb.AggregationTemporality(hist.AggregationTemporality()),
						},
					}
				case pmetric.MetricTypeExponentialHistogram:
					hist := m.ExponentialHistogram()
					metric.Data = &metricpb.Metric_ExponentialHistogram{
						ExponentialHistogram: &metricpb.ExponentialHistogram{
							DataPoints:             convertExponentialHistogramDataPoints(hist.DataPoints()),
							AggregationTemporality: metricpb.AggregationTemporality(hist.AggregationTemporality()),
						},
					}
				case pmetric.MetricTypeSummary:
					summ := m.Summary()
					metric.Data = &metricpb.Metric_Summary{
						Summary: &metricpb.Summary{
							DataPoints: convertSummaryDataPoints(summ.DataPoints()),
						},
					}
				}

				scopeMetrics.Metrics = append(scopeMetrics.Metrics, metric)
			}

			resourceMetrics.ScopeMetrics = append(resourceMetrics.ScopeMetrics, scopeMetrics)
		}

		request.ResourceMetrics = append(request.ResourceMetrics, resourceMetrics)
	}

	return request
}

func convertNumberDataPoints(dps pmetric.NumberDataPointSlice) []*metricpb.NumberDataPoint {
	result := make([]*metricpb.NumberDataPoint, 0, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		numberDataPoint := &metricpb.NumberDataPoint{
			TimeUnixNano: uint64(dp.Timestamp()),
			Attributes:   convertAttributes(dp.Attributes()),
		}

		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeDouble:
			numberDataPoint.Value = &metricpb.NumberDataPoint_AsDouble{
				AsDouble: dp.DoubleValue(),
			}
		case pmetric.NumberDataPointValueTypeInt:
			numberDataPoint.Value = &metricpb.NumberDataPoint_AsInt{
				AsInt: dp.IntValue(),
			}
		}

		result = append(result, numberDataPoint)
	}
	return result
}

func convertHistogramDataPoints(dps pmetric.HistogramDataPointSlice) []*metricpb.HistogramDataPoint {
	result := make([]*metricpb.HistogramDataPoint, 0, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		sum := dp.Sum()
		histogramDataPoint := &metricpb.HistogramDataPoint{
			TimeUnixNano:   uint64(dp.Timestamp()),
			Attributes:     convertAttributes(dp.Attributes()),
			Count:          dp.Count(),
			Sum:            &sum,
			BucketCounts:   dp.BucketCounts().AsRaw(),
			ExplicitBounds: dp.ExplicitBounds().AsRaw(),
		}
		result = append(result, histogramDataPoint)
	}
	return result
}

func convertExponentialHistogramDataPoints(dps pmetric.ExponentialHistogramDataPointSlice) []*metricpb.ExponentialHistogramDataPoint {
	result := make([]*metricpb.ExponentialHistogramDataPoint, 0, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		sum := dp.Sum()
		histogramDataPoint := &metricpb.ExponentialHistogramDataPoint{
			TimeUnixNano: uint64(dp.Timestamp()),
			Attributes:   convertAttributes(dp.Attributes()),
			Count:        dp.Count(),
			Sum:          &sum,
			Scale:        dp.Scale(),
			ZeroCount:    dp.ZeroCount(),
		}

		pos := dp.Positive()
		if pos.BucketCounts().Len() > 0 {
			histogramDataPoint.Positive = &metricpb.ExponentialHistogramDataPoint_Buckets{
				Offset:       pos.Offset(),
				BucketCounts: pos.BucketCounts().AsRaw(),
			}
		}

		neg := dp.Negative()
		if neg.BucketCounts().Len() > 0 {
			histogramDataPoint.Negative = &metricpb.ExponentialHistogramDataPoint_Buckets{
				Offset:       neg.Offset(),
				BucketCounts: neg.BucketCounts().AsRaw(),
			}
		}

		result = append(result, histogramDataPoint)
	}
	return result
}

func convertSummaryDataPoints(dps pmetric.SummaryDataPointSlice) []*metricpb.SummaryDataPoint {
	result := make([]*metricpb.SummaryDataPoint, 0, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		summaryDataPoint := &metricpb.SummaryDataPoint{
			TimeUnixNano:   uint64(dp.Timestamp()),
			Attributes:     convertAttributes(dp.Attributes()),
			Count:          dp.Count(),
			Sum:            dp.Sum(),
			QuantileValues: make([]*metricpb.SummaryDataPoint_ValueAtQuantile, 0, dp.QuantileValues().Len()),
		}

		for j := 0; j < dp.QuantileValues().Len(); j++ {
			qv := dp.QuantileValues().At(j)
			summaryDataPoint.QuantileValues = append(summaryDataPoint.QuantileValues, &metricpb.SummaryDataPoint_ValueAtQuantile{
				Quantile: qv.Quantile(),
				Value:    qv.Value(),
			})
		}

		result = append(result, summaryDataPoint)
	}
	return result
}
