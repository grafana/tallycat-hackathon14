package grpcserver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"

	"github.com/tallycat/tallycat/internal/repository"
	"github.com/tallycat/tallycat/internal/schema"
)

type TracesServiceServer struct {
	tracespb.UnimplementedTraceServiceServer
	schemaRepo repository.TelemetrySchemaRepository
	logger     *slog.Logger
}

func NewTracesServiceServer(schemaRepo repository.TelemetrySchemaRepository) *TracesServiceServer {
	return &TracesServiceServer{
		schemaRepo: schemaRepo,
	}
}

// convertAttributeValue converts a protobuf AnyValue to the appropriate pcommon value
func convertAttributeValue(attrs pcommon.Map, key string, value *commonpb.AnyValue) {
	if value == nil {
		return
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		attrs.PutStr(key, v.StringValue)
	case *commonpb.AnyValue_BoolValue:
		attrs.PutBool(key, v.BoolValue)
	case *commonpb.AnyValue_IntValue:
		attrs.PutInt(key, v.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		attrs.PutDouble(key, v.DoubleValue)
	case *commonpb.AnyValue_ArrayValue:
		// For arrays, convert to string representation
		attrs.PutStr(key, value.String())
	case *commonpb.AnyValue_KvlistValue:
		// For key-value lists, convert to string representation
		attrs.PutStr(key, value.String())
	case *commonpb.AnyValue_BytesValue:
		// For bytes, convert to string
		attrs.PutStr(key, string(v.BytesValue))
	default:
		// Fallback to string representation
		attrs.PutStr(key, value.String())
	}
}

func (s *TracesServiceServer) Export(ctx context.Context, req *tracespb.ExportTraceServiceRequest) (*tracespb.ExportTraceServiceResponse, error) {
	traces := ptrace.NewTraces()
	rts := traces.ResourceSpans()
	rts.EnsureCapacity(len(req.ResourceSpans))

	for _, rt := range req.ResourceSpans {
		resourceSpan := rts.AppendEmpty()
		resourceSpan.SetSchemaUrl(rt.SchemaUrl)

		// Convert resource attributes
		if rt.Resource != nil {
			for _, attr := range rt.Resource.Attributes {
				convertAttributeValue(resourceSpan.Resource().Attributes(), attr.Key, attr.Value)
			}
		}

		// Convert scope spans
		sts := resourceSpan.ScopeSpans()
		sts.EnsureCapacity(len(rt.ScopeSpans))

		for _, st := range rt.ScopeSpans {
			scopeSpan := sts.AppendEmpty()
			scopeSpan.SetSchemaUrl(st.SchemaUrl)

			// Convert scope
			if st.Scope != nil {
				scopeSpan.Scope().SetName(st.Scope.Name)
				scopeSpan.Scope().SetVersion(st.Scope.Version)
				scopeSpan.SetSchemaUrl(st.SchemaUrl)
				for _, attr := range st.Scope.Attributes {
					convertAttributeValue(scopeSpan.Scope().Attributes(), attr.Key, attr.Value)
				}
			}

			// Convert logs
			ls := scopeSpan.Spans()
			ls.EnsureCapacity(len(st.Spans))

			for _, s := range st.Spans {
				span := ls.AppendEmpty()
				span.SetKind(ptrace.SpanKind(s.Kind))
				span.SetName(s.Name)

				// Safely convert ParentSpanID (8 bytes expected)
				if len(s.ParentSpanId) == 8 {
					span.SetParentSpanID(pcommon.SpanID(s.ParentSpanId))
				}

				// Safely convert SpanID (8 bytes expected)
				if len(s.SpanId) == 8 {
					span.SetSpanID(pcommon.SpanID(s.SpanId))
				}

				// Safely convert TraceID (16 bytes expected)
				if len(s.TraceId) == 16 {
					span.SetTraceID(pcommon.TraceID(s.TraceId))
				}
				span.SetStartTimestamp(pcommon.Timestamp(s.StartTimeUnixNano))
				span.SetEndTimestamp(pcommon.Timestamp(s.EndTimeUnixNano))
				span.SetFlags(uint32(s.Flags))
				span.Status().SetCode(ptrace.StatusCode(s.Status.Code))
				span.Status().SetMessage(s.Status.Message)

				for _, attr := range s.Attributes {
					convertAttributeValue(span.Attributes(), attr.Key, attr.Value)
				}
				span.SetDroppedAttributesCount(s.DroppedAttributesCount)
				span.SetDroppedLinksCount(s.DroppedLinksCount)

				for _, event := range s.Events {
					spanEvent := span.Events().AppendEmpty()
					spanEvent.SetName(event.Name)
					spanEvent.SetTimestamp(pcommon.Timestamp(event.TimeUnixNano))
					spanEvent.SetDroppedAttributesCount(event.DroppedAttributesCount)

					// Convert event attributes
					for _, attr := range event.Attributes {
						convertAttributeValue(spanEvent.Attributes(), attr.Key, attr.Value)
					}
				}
				span.SetDroppedEventsCount(s.DroppedEventsCount)

				for _, link := range s.Links {
					spanLink := span.Links().AppendEmpty()

					// Safely convert TraceID (16 bytes expected)
					if len(link.TraceId) == 16 {
						spanLink.SetTraceID(pcommon.TraceID(link.TraceId))
					}

					// Safely convert SpanID (8 bytes expected)
					if len(link.SpanId) == 8 {
						spanLink.SetSpanID(pcommon.SpanID(link.SpanId))
					}

					spanLink.SetDroppedAttributesCount(link.DroppedAttributesCount)
					spanLink.SetFlags(uint32(link.Flags))

					// Convert link attributes
					for _, attr := range link.Attributes {
						convertAttributeValue(spanLink.Attributes(), attr.Key, attr.Value)
					}
				}
				span.SetDroppedLinksCount(s.DroppedLinksCount)
			}
		}
	}

	// Extract schemas from the converted traces
	schemas := schema.ExtractFromTraces(traces)

	if err := s.schemaRepo.RegisterTelemetrySchemas(ctx, schemas); err != nil {
		slog.Error("failed to register schemas", "error", err, "signal", "traces")
		return nil, err
	}

	return &tracespb.ExportTraceServiceResponse{}, nil
}
