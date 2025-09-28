package grpcserver_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/stretchr/testify/require"

	"github.com/tallycat/tallycat/internal/integration/testutil"
)

func TestExportTraces(t *testing.T) {
	tests := []struct {
		name     string
		dataFile string
		wantErr  bool
	}{
		{
			name:     "single trace single schema",
			dataFile: "single_trace_single_schema.yaml",
			wantErr:  false,
		},
		{
			name:     "single trace multiple schemas",
			dataFile: "single_trace_multiple_schemas.yaml",
			wantErr:  false,
		},
		{
			name:     "two traces two schemas",
			dataFile: "two_traces_two_schemas.yaml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		// Setup test database
		db := testutil.NewTestDB(t)
		defer db.Close()

		ctx := context.Background()
		db.SetupTestDB(t)
		// defer db.CleanupTestDB(t)

		// Setup test server
		server := testutil.NewTestServer(t, db)
		defer server.Close()

		// Load test data
		traces, err := golden.ReadTraces(filepath.Join("testdata", tt.dataFile))
		require.NoError(t, err)

		// Convert traces to request
		req := testutil.ConvertPtraceToRequest(traces)

		// Send request to server
		resp, err := server.TracesClient.Export(ctx, req)
		if tt.wantErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, resp)
	}
}
