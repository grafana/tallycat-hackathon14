package grpcserver

import (
	"context"
	"log/slog"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	"github.com/tallycat/tallycat/internal/repository"
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
	for _, resourceLogs := range req.ResourceLogs {
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			for _, logRecord := range scopeLogs.LogRecords {
				if logRecord.Body != nil && logRecord.Body.GetStringValue() == "tallycat.schema.extracted" {
				}
			}
		}
	}

	return &logspb.ExportLogsServiceResponse{}, nil
}
