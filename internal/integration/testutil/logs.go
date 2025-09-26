package testutil

import (
	"go.opentelemetry.io/collector/pdata/plog"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// ConvertPmetricToRequest converts pmetric.Metrics to ExportMetricsServiceRequest
func ConvertPlogToRequest(ld plog.Logs) *logspb.ExportLogsServiceRequest {
	request := &logspb.ExportLogsServiceRequest{
		ResourceLogs: make([]*logpb.ResourceLogs, 0, ld.ResourceLogs().Len()),
	}

	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rm := ld.ResourceLogs().At(i)
		resourceLogs := &logpb.ResourceLogs{
			Resource: &resourcepb.Resource{
				Attributes: convertAttributes(rm.Resource().Attributes()),
			},
			ScopeLogs: make([]*logpb.ScopeLogs, 0, rm.ScopeLogs().Len()),
			SchemaUrl: rm.SchemaUrl(),
		}

		for j := 0; j < rm.ScopeLogs().Len(); j++ {
			sm := rm.ScopeLogs().At(j)
			scopeLogs := &logpb.ScopeLogs{
				Scope: &commonpb.InstrumentationScope{
					Name:    sm.Scope().Name(),
					Version: sm.Scope().Version(),
				},
				LogRecords: make([]*logpb.LogRecord, 0, sm.LogRecords().Len()),
				SchemaUrl:  sm.SchemaUrl(),
			}

			for k := 0; k < sm.LogRecords().Len(); k++ {
				l := sm.LogRecords().At(k)
				traceID := l.TraceID()
				spanID := l.SpanID()
				logRecord := &logpb.LogRecord{
					TimeUnixNano:         uint64(l.Timestamp()),
					ObservedTimeUnixNano: uint64(l.ObservedTimestamp()),
					SeverityNumber:       logpb.SeverityNumber(l.SeverityNumber()),
					SeverityText:         l.SeverityText(),
					Body: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: l.Body().AsString(),
						},
					},
					Flags:     uint32(l.Flags()),
					TraceId:   traceID[:],
					SpanId:    spanID[:],
					EventName: l.EventName(),
				}
				scopeLogs.LogRecords = append(scopeLogs.LogRecords, logRecord)
			}
			resourceLogs.ScopeLogs = append(resourceLogs.ScopeLogs, scopeLogs)
		}
		request.ResourceLogs = append(request.ResourceLogs, resourceLogs)
	}
	return request
}
