package schema

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/stretchr/testify/require"
)

func TestExtractMetricSchema(t *testing.T) {
	testCases := []struct {
		name     string
		expected int
	}{
		{
			name:     "single_metric_single_schema",
			expected: 1,
		},
		{
			name:     "single_metric_multiple_schemas",
			expected: 2,
		},
		{
			name:     "two_metrics_two_schemas",
			expected: 2,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			md, err := golden.ReadMetrics(filepath.Join("testdata", fmt.Sprintf("%s.yaml", tc.name)))
			require.NoError(t, err)

			schemas := ExtractMetricSchema(md)
			require.Equal(t, tc.expected, len(schemas))
		})
	}

}
