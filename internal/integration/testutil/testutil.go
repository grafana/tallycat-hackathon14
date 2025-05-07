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
	collectorlogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestDB represents a test database instance
type TestDB struct {
	conn *sql.DB
	pool *duckdb.ConnectionPool
	repo *duckdb.SchemaRepository
}

// NewTestDB creates a new test database instance
func NewTestDB(t *testing.T) *TestDB {
	pool, err := duckdb.NewConnectionPool(&duckdb.Config{
		DatabasePath:    "tallycat.db", // Use in-memory database for tests
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 5,
	}, slog.Default())
	require.NoError(t, err)

	conn := pool.GetConnection()
	repo := duckdb.NewSchemaRepository(pool.(*duckdb.ConnectionPool), slog.Default())

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
	server        *grpc.Server
	LogsClient    collectorlogspb.LogsServiceClient
	MetricsClient metricspb.MetricsServiceClient
	conn          *grpc.ClientConn
}

// NewTestServer creates a new test gRPC server
func NewTestServer(t *testing.T, db *TestDB) *TestServer {
	server := grpc.NewServer()
	logsServer := grpcserver.NewLogsServiceServer(db.repo, slog.Default())
	metricsServer := grpcserver.NewMetricsServiceServer(db.repo, slog.Default())
	collectorlogspb.RegisterLogsServiceServer(server, logsServer)
	metricspb.RegisterMetricsServiceServer(server, metricsServer)

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

	return &TestServer{
		server:        server,
		LogsClient:    logsClient,
		MetricsClient: metricsClient,
		conn:          conn,
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
