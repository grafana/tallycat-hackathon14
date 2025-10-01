package grpcserver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

type LogsServiceServer struct {
	logspb.UnimplementedLogsServiceServer
	schemaRepo repository.TelemetrySchemaRepository
	logger     *slog.Logger
}

func NewLogsServiceServer(schemaRepo repository.TelemetrySchemaRepository) *LogsServiceServer {
	return &LogsServiceServer{
		schemaRepo: schemaRepo,
	}
}

func (s *LogsServiceServer) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs()
	rls.EnsureCapacity(len(req.ResourceLogs))

	for _, rl := range req.ResourceLogs {
		resourceLog := rls.AppendEmpty()
		resourceLog.SetSchemaUrl(rl.SchemaUrl)

		// Convert resource attributes
		if rl.Resource != nil {
			for _, attr := range rl.Resource.Attributes {
				resourceLog.Resource().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
			}
		}

		// Convert scope logs
		sls := resourceLog.ScopeLogs()
		sls.EnsureCapacity(len(rl.ScopeLogs))

		for _, sl := range rl.ScopeLogs {
			scopeLog := sls.AppendEmpty()
			scopeLog.SetSchemaUrl(sl.SchemaUrl)

			// Convert scope
			if sl.Scope != nil {
				scopeLog.Scope().SetName(sl.Scope.Name)
				scopeLog.Scope().SetVersion(sl.Scope.Version)
				scopeLog.SetSchemaUrl(sl.SchemaUrl)
				for _, attr := range sl.Scope.Attributes {
					scopeLog.Scope().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
				}
			}

			// Convert logs
			ls := scopeLog.LogRecords()
			ls.EnsureCapacity(len(sl.LogRecords))

			for _, l := range sl.LogRecords {
				logRecord := ls.AppendEmpty()
				logRecord.SetTimestamp(pcommon.Timestamp(l.TimeUnixNano))
				logRecord.SetObservedTimestamp(pcommon.Timestamp(l.ObservedTimeUnixNano))
				logRecord.SetSeverityNumber(plog.SeverityNumber(l.SeverityNumber))
				logRecord.SetSeverityText(l.SeverityText)
				logRecord.Body().SetStr(l.Body.GetStringValue())
				logRecord.SetFlags(plog.LogRecordFlags(l.Flags))
				if len(l.TraceId) == 16 {
					logRecord.SetTraceID(pcommon.TraceID(l.TraceId))
				}
				if len(l.SpanId) == 8 {
					logRecord.SetSpanID(pcommon.SpanID(l.SpanId))
				}
				logRecord.SetEventName(l.EventName)

				for _, attr := range l.Attributes {
					logRecord.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
				}
				logRecord.SetDroppedAttributesCount(l.DroppedAttributesCount)
			}
		}
	}

	// Extract schemas from the converted logs
	schemas := schema.ExtractFromLogs(logs)

	if err := s.schemaRepo.RegisterTelemetrySchemas(ctx, schemas); err != nil {
		slog.Error("failed to register schemas", "error", err, "signal", "logs")
		return nil, err
	}

	return &logspb.ExportLogsServiceResponse{}, nil
}
