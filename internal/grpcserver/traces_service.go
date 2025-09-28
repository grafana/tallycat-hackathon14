package grpcserver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"

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
				resourceSpan.Resource().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
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
				for _, attr := range st.Scope.Attributes {
					scopeSpan.Scope().Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
				}
			}

			// Convert logs
			ls := scopeSpan.Spans()
			ls.EnsureCapacity(len(st.Spans))

			for _, s := range st.Spans {
				span := ls.AppendEmpty()
				span.SetKind(ptrace.SpanKind(s.Kind))
				span.SetName(s.Name)
				span.SetParentSpanID(pcommon.SpanID(s.ParentSpanId))
				span.SetSpanID(pcommon.SpanID(s.SpanId))
				span.SetTraceID(pcommon.TraceID(s.TraceId))
				span.SetStartTimestamp(pcommon.Timestamp(s.StartTimeUnixNano))
				span.SetEndTimestamp(pcommon.Timestamp(s.EndTimeUnixNano))
				span.SetFlags(uint32(s.Flags))
				span.Status().SetCode(ptrace.StatusCode(s.Status.Code))
				span.Status().SetMessage(s.Status.Message)
				
				for _, attr := range s.Attributes {
					span.Attributes().PutStr(attr.Key, attr.Value.GetStringValue())
				}
				span.SetDroppedAttributesCount(s.DroppedAttributesCount)
				span.SetDroppedLinksCount(s.DroppedLinksCount)
				
				for _, event := range s.Events {
					spanEvent := span.Events().AppendEmpty()
					spanEvent.SetName(event.Name)
					spanEvent.SetTimestamp(pcommon.Timestamp(event.TimeUnixNano))
					spanEvent.SetDroppedAttributesCount(event.DroppedAttributesCount)
				}
				span.SetDroppedEventsCount(s.DroppedEventsCount)

				for _, link := range s.Links {
					spanLink := span.Links().AppendEmpty()
					spanLink.SetTraceID(pcommon.TraceID(link.TraceId))
					spanLink.SetSpanID(pcommon.SpanID(link.SpanId))
					spanLink.SetDroppedAttributesCount(link.DroppedAttributesCount)
					spanLink.SetFlags(uint32(link.Flags))
				}
				span.SetDroppedLinksCount(s.DroppedLinksCount)
			}
		}
	}

	// Extract schemas from the converted traces
	schemas := schema.ExtractFromTraces(traces)

	if err := s.schemaRepo.RegisterTelemetrySchemas(ctx, schemas); err != nil {
		slog.Error("failed to register schemas", "error", err)
		return nil, err
	}

	return &tracespb.ExportTraceServiceResponse{}, nil
}
