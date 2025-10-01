package schema

import (
	"testing"
)

func TestSanitizeScopeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{
			name:     "simple lowercase name",
			input:    "http",
			expected: "http",
		},
		{
			name:     "simple uppercase name",
			input:    "HTTP",
			expected: "http",
		},
		{
			name:     "mixed case name",
			input:    "HttpClient",
			expected: "httpclient",
		},

		// Special character replacement
		{
			name:     "hyphen replacement",
			input:    "http-client",
			expected: "http_client",
		},
		{
			name:     "dot replacement",
			input:    "database.connection",
			expected: "database_connection",
		},
		{
			name:     "slash replacement",
			input:    "MyApp/Service",
			expected: "myapp_service",
		},
		{
			name:     "multiple special characters",
			input:    "HTTP-Client.v2/Service",
			expected: "http_client_v2_service",
		},
		{
			name:     "space replacement",
			input:    "my service",
			expected: "my_service",
		},
		{
			name:     "colon replacement",
			input:    "service:name",
			expected: "service_name",
		},
		{
			name:     "multiple consecutive special chars",
			input:    "service--name..test",
			expected: "service__name__test",
		},

		// Alphanumeric preservation
		{
			name:     "alphanumeric with numbers",
			input:    "service1",
			expected: "service1",
		},
		{
			name:     "mixed alphanumeric",
			input:    "HTTP2Client",
			expected: "http2client",
		},
		{
			name:     "numbers and special chars",
			input:    "v1.2.3-beta",
			expected: "v1_2_3_beta",
		},

		// Edge cases for empty/unknown
		{
			name:     "empty string",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "unknown",
		},
		{
			name:     "UNKNOWN uppercase",
			input:    "UNKNOWN",
			expected: "unknown",
		},
		{
			name:     "unknown lowercase",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "Unknown mixed case",
			input:    "Unknown",
			expected: "unknown",
		},
		{
			name:     "UNKNOWN with whitespace",
			input:    "  UNKNOWN  ",
			expected: "unknown",
		},
		{
			name:     "unknown with whitespace",
			input:    "  unknown  ",
			expected: "unknown",
		},

		// Complex real-world examples
		{
			name:     "opentelemetry instrumentation",
			input:    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			expected: "go_opentelemetry_io_contrib_instrumentation_net_http_otelhttp",
		},
		{
			name:     "kubernetes service",
			input:    "k8s.io/client-go",
			expected: "k8s_io_client_go",
		},
		{
			name:     "github package",
			input:    "github.com/prometheus/client_golang",
			expected: "github_com_prometheus_client_golang",
		},
		{
			name:     "version with symbols",
			input:    "my-service@1.2.3",
			expected: "my_service_1_2_3",
		},

		// Unicode and special cases
		{
			name:     "unicode characters",
			input:    "service-ñame",
			expected: "service__ame",
		},
		{
			name:     "brackets and parentheses",
			input:    "service[test](v1)",
			expected: "service_test__v1_",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "single special character",
			input:    "-",
			expected: "_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeScopeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeScopeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
