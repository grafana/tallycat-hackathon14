package grpcserver

import (
	"context"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

// LogsServiceServer implements the OpenTelemetry LogsService interface
type LogsServiceServer struct {
	logspb.UnimplementedLogsServiceServer
}

// NewLogsServiceServer creates a new LogsServiceServer
func NewLogsServiceServer() *LogsServiceServer {
	return &LogsServiceServer{}
}

// Export implements the Export method of the LogsService interface
func (s *LogsServiceServer) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	// TODO: Implement log processing logic here
	return &logspb.ExportLogsServiceResponse{}, nil
}
