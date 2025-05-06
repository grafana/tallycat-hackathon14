package grpcserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	"github.com/tallycat/tallycat/internal/repository"
)

type LogsServiceServer struct {
	logspb.UnimplementedLogsServiceServer
	schemaRepo repository.SchemaProvider
	logger     *slog.Logger
}

func NewLogsServiceServer(schemaRepo repository.SchemaProvider, logger *slog.Logger) *LogsServiceServer {
	return &LogsServiceServer{
		schemaRepo: schemaRepo,
		logger:     logger,
	}
}

func (s *LogsServiceServer) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	for _, resourceLogs := range req.ResourceLogs {
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			for _, logRecord := range scopeLogs.LogRecords {
				if logRecord.Body != nil && logRecord.Body.GetStringValue() == "tallycat.schema.extracted" {
					schema := &repository.Schema{
						FieldTypes:   make(map[string]string),
						FieldSources: make(map[string]string),
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.id" {
							schema.SchemaID = attr.Value.GetStringValue()
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.signal_type" {
							schema.SignalType = attr.Value.GetStringValue()
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.scope_name" {
							schema.ScopeName = attr.Value.GetStringValue()
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.seen_count" {
							if count, err := strconv.Atoi(attr.Value.GetStringValue()); err == nil {
								schema.SeenCount = count
							}
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.field_names" {
							if err := json.Unmarshal([]byte(attr.Value.GetStringValue()), &schema.FieldNames); err != nil {
								s.logger.Error("failed to parse field names", "error", err)
								return nil, err
							}
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.field_types" {
							if err := json.Unmarshal([]byte(attr.Value.GetStringValue()), &schema.FieldTypes); err != nil {
								s.logger.Error("failed to parse field types", "error", err)
								return nil, err
							}
						}
					}

					for _, attr := range logRecord.Attributes {
						if attr.Key == "tallycat.schema.field_sources" {
							if err := json.Unmarshal([]byte(attr.Value.GetStringValue()), &schema.FieldSources); err != nil {
								s.logger.Error("failed to parse field sources", "error", err)
								return nil, err
							}
						}
					}

					if err := s.schemaRepo.RegisterSchema(ctx, schema); err != nil {
						s.logger.Error("failed to register schema", "error", err, "schema_id", schema.SchemaID)
						return nil, err
					}
				}
			}
		}
	}

	return &logspb.ExportLogsServiceResponse{}, nil
}
