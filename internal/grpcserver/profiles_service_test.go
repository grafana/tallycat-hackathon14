package grpcserver_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/stretchr/testify/require"

	"github.com/tallycat/tallycat/internal/integration/testutil"
)

func TestExportProfiles(t *testing.T) {
	tests := []struct {
		name     string
		dataFile string
		wantErr  bool
	}{
		{
			name:     "single profile single schema",
			dataFile: "single_profile_single_schema.yaml",
			wantErr:  false,
		},
		{
			name:     "single profile multiple schemas",
			dataFile: "single_profile_multiple_schemas.yaml",
			wantErr:  false,
		},
		{
			name:     "two profiles two schemas",
			dataFile: "two_profiles_two_schemas.yaml",
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
		md, err := golden.ReadProfiles(filepath.Join("testdata", tt.dataFile))
		require.NoError(t, err)

		// Convert metrics to request
		req := testutil.ConvertProfileToRequest(md)

		// Send request to server
		resp, err := server.ProfilesClient.Export(ctx, req)
		if tt.wantErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, resp)
	}
}
