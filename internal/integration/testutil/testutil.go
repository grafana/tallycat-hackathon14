package testutil

import (
	"database/sql"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tallycat/tallycat/internal/grpcserver"
	"github.com/tallycat/tallycat/internal/repository/duckdb"
	"github.com/tallycat/tallycat/internal/repository/duckdb/migrator"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	profilespb "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	tracespb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestDB represents a test database instance
type TestDB struct {
	conn *sql.DB
	pool *duckdb.ConnectionPool
	repo *duckdb.TelemetrySchemaRepository
}

// NewTestDB creates a new test database instance
func NewTestDB(t *testing.T) *TestDB {
	pool, err := duckdb.NewConnectionPool(&duckdb.Config{
		DatabasePath:    ":memory:", // Use in-memory database for tests
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 5,
	}, slog.Default())
	require.NoError(t, err)

	conn := pool.GetConnection()
	repo := duckdb.NewTelemetrySchemaRepository(pool.(*duckdb.ConnectionPool))

	return &TestDB{
		conn: conn,
		pool: pool.(*duckdb.ConnectionPool),
		repo: repo,
	}
}

// Close closes the test database connection
func (db *TestDB) Close() error {
	return db.pool.Close()
}

// Repo returns the telemetry schema repository
func (db *TestDB) Repo() *duckdb.TelemetrySchemaRepository {
	return db.repo
}

// SetupTestDB sets up the test database with the required schema
func (db *TestDB) SetupTestDB(t *testing.T) {
	// Apply migrations instead of direct schema creation
	err := migrator.ApplyMigrations(db.conn)
	require.NoError(t, err)
}

// CleanupTestDB cleans up the test database
func (db *TestDB) CleanupTestDB(t *testing.T) {
	// Since we're using in-memory database, we don't need to explicitly clean up
	// The database will be destroyed when the connection is closed
}

// TestServer represents a test gRPC server
type TestServer struct {
	server         *grpc.Server
	LogsClient     collectorlogspb.LogsServiceClient
	MetricsClient  metricspb.MetricsServiceClient
	ProfilesClient profilespb.ProfilesServiceClient
	TracesClient   tracespb.TraceServiceClient
	conn           *grpc.ClientConn
}

// NewTestServer creates a new test gRPC server
func NewTestServer(t *testing.T, db *TestDB) *TestServer {
	server := grpc.NewServer()
	logsServer := grpcserver.NewLogsServiceServer(db.repo)
	metricsServer := grpcserver.NewMetricsServiceServer(db.repo)
	profilesServer := grpcserver.NewProfilesServiceServer(db.repo)
	tracesServer := grpcserver.NewTracesServiceServer(db.repo)
	collectorlogspb.RegisterLogsServiceServer(server, logsServer)
	metricspb.RegisterMetricsServiceServer(server, metricsServer)
	profilespb.RegisterProfilesServiceServer(server, profilesServer)
	tracespb.RegisterTraceServiceServer(server, tracesServer)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() {
		err := server.Serve(lis)
		require.NoError(t, err)
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	logsClient := collectorlogspb.NewLogsServiceClient(conn)
	metricsClient := metricspb.NewMetricsServiceClient(conn)
	profilesClient := profilespb.NewProfilesServiceClient(conn)
	tracesClient := tracespb.NewTraceServiceClient(conn)

	return &TestServer{
		server:         server,
		LogsClient:     logsClient,
		MetricsClient:  metricsClient,
		ProfilesClient: profilesClient,
		TracesClient:   tracesClient,
		conn:           conn,
	}
}

// Close closes the test server
func (s *TestServer) Close() {
	s.server.Stop()
	s.conn.Close()
}

// convertAttributes converts pcommon.Map to []*commonpb.KeyValue
func convertAttributes(attrs pcommon.Map) []*commonpb.KeyValue {
	result := make([]*commonpb.KeyValue, 0, attrs.Len())
	attrs.Range(func(k string, v pcommon.Value) bool {
		kv := &commonpb.KeyValue{
			Key:   k,
			Value: convertValue(v),
		}
		result = append(result, kv)
		return true
	})
	return result
}

// convertValue converts pcommon.Value to commonpb.AnyValue
func convertValue(v pcommon.Value) *commonpb.AnyValue {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: v.Str(),
			},
		}
	case pcommon.ValueTypeInt:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: v.Int(),
			},
		}
	case pcommon.ValueTypeDouble:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_DoubleValue{
				DoubleValue: v.Double(),
			},
		}
	case pcommon.ValueTypeBool:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BoolValue{
				BoolValue: v.Bool(),
			},
		}
	default:
		return &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: v.AsString(),
			},
		}
	}
}

// ConvertPtraceToRequest converts ptrace.Traces to tracespb.ExportTraceServiceRequest
func ConvertPtraceToRequest(traces ptrace.Traces) *tracespb.ExportTraceServiceRequest {
	req := &tracespb.ExportTraceServiceRequest{
		ResourceSpans: make([]*tracepb.ResourceSpans, 0, traces.ResourceSpans().Len()),
	}

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		resourceSpan := &tracepb.ResourceSpans{
			SchemaUrl: rs.SchemaUrl(),
			Resource: &resourcepb.Resource{
				Attributes: convertAttributes(rs.Resource().Attributes()),
			},
			ScopeSpans: make([]*tracepb.ScopeSpans, 0, rs.ScopeSpans().Len()),
		}

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			scopeSpan := &tracepb.ScopeSpans{
				SchemaUrl: ss.SchemaUrl(),
				Scope: &commonpb.InstrumentationScope{
					Name:       ss.Scope().Name(),
					Version:    ss.Scope().Version(),
					Attributes: convertAttributes(ss.Scope().Attributes()),
				},
				Spans: make([]*tracepb.Span, 0, ss.Spans().Len()),
			}

			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)
				traceID := span.TraceID()
				spanID := span.SpanID()
				parentSpanID := span.ParentSpanID()
				pbSpan := &tracepb.Span{
					TraceId:                traceID[:],
					SpanId:                 spanID[:],
					ParentSpanId:           parentSpanID[:],
					Name:                   span.Name(),
					Kind:                   tracepb.Span_SpanKind(span.Kind()),
					StartTimeUnixNano:      uint64(span.StartTimestamp()),
					EndTimeUnixNano:        uint64(span.EndTimestamp()),
					Attributes:             convertAttributes(span.Attributes()),
					DroppedAttributesCount: span.DroppedAttributesCount(),
					DroppedEventsCount:     span.DroppedEventsCount(),
					DroppedLinksCount:      span.DroppedLinksCount(),
					Flags:                  uint32(span.Flags()),
				}

				// Convert events
				pbSpan.Events = make([]*tracepb.Span_Event, 0, span.Events().Len())
				for l := 0; l < span.Events().Len(); l++ {
					event := span.Events().At(l)
					pbEvent := &tracepb.Span_Event{
						TimeUnixNano:           uint64(event.Timestamp()),
						Name:                   event.Name(),
						Attributes:             convertAttributes(event.Attributes()),
						DroppedAttributesCount: event.DroppedAttributesCount(),
					}
					pbSpan.Events = append(pbSpan.Events, pbEvent)
				}

				// Convert links
				pbSpan.Links = make([]*tracepb.Span_Link, 0, span.Links().Len())
				for l := 0; l < span.Links().Len(); l++ {
					link := span.Links().At(l)
					linkTraceID := link.TraceID()
					linkSpanID := link.SpanID()
					pbLink := &tracepb.Span_Link{
						TraceId:                linkTraceID[:],
						SpanId:                 linkSpanID[:],
						Attributes:             convertAttributes(link.Attributes()),
						DroppedAttributesCount: link.DroppedAttributesCount(),
						Flags:                  uint32(link.Flags()),
					}
					pbSpan.Links = append(pbSpan.Links, pbLink)
				}

				// Convert status
				pbSpan.Status = &tracepb.Status{
					Code:    tracepb.Status_StatusCode(span.Status().Code()),
					Message: span.Status().Message(),
				}

				scopeSpan.Spans = append(scopeSpan.Spans, pbSpan)
			}

			resourceSpan.ScopeSpans = append(resourceSpan.ScopeSpans, scopeSpan)
		}

		req.ResourceSpans = append(req.ResourceSpans, resourceSpan)
	}

	return req
}
