package schema

import (
	"testing"
)

func TestScopeTransformer_Transform(t *testing.T) {
	transformer := NewScopeTransformer()

	testCases := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		// 1. OpenTelemetry Collector Components -> "opentelemetry-collector"
		{
			name:           "Collector exporter helper",
			input:          "go.opentelemetry.io/collector/exporter/exporterhelper",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector batch processor",
			input:          "go.opentelemetry.io/collector/processor/batchprocessor",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector memory limiter processor",
			input:          "go.opentelemetry.io/collector/processor/memorylimiterprocessor",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector processor helper",
			input:          "go.opentelemetry.io/collector/processor/processorhelper",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector receiver helper",
			input:          "go.opentelemetry.io/collector/receiver/receiverhelper",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector scraper helper",
			input:          "go.opentelemetry.io/collector/scraper/scraperhelper",
			expectedOutput: "opentelemetry-collector",
		},
		{
			name:           "Collector service",
			input:          "go.opentelemetry.io/collector/service",
			expectedOutput: "opentelemetry-collector",
		},

		// 2. Docker -> "docker"
		{
			name:           "Docker stats receiver",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver",
			expectedOutput: "docker",
		},

		// 3. HTTP -> "http"
		{
			name:           "HTTP check receiver",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver",
			expectedOutput: "http",
		},
		{
			name:           "Go HTTP instrumentation",
			input:          "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp",
			expectedOutput: "http",
		},
		{
			name:           "Node.js HTTP instrumentation",
			input:          "@opentelemetry/instrumentation-http",
			expectedOutput: "http",
		},
		{
			name:           "Java OTLP HTTP exporter",
			input:          "io.opentelemetry.exporters.otlp-http",
			expectedOutput: "http",
		},
		{
			name:           "Nginx receiver",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver",
			expectedOutput: "nginx",
		},

		// 4. gRPC -> "grpc"
		{
			name:           "Go gRPC instrumentation",
			input:          "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc",
			expectedOutput: "grpc",
		},
		{
			name:           "Java gRPC instrumentation",
			input:          "io.opentelemetry.grpc-1.6",
			expectedOutput: "grpc",
		},
		{
			name:           "Python gRPC instrumentation",
			input:          "opentelemetry.instrumentation.grpc",
			expectedOutput: "grpc",
		},

		// 5. Host Metrics -> "host"
		{
			name:           "CPU scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper",
			expectedOutput: "host",
		},
		{
			name:           "Disk scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper",
			expectedOutput: "host",
		},
		{
			name:           "Filesystem scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper",
			expectedOutput: "host",
		},
		{
			name:           "Load scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/loadscraper",
			expectedOutput: "host",
		},
		{
			name:           "Memory scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/memoryscraper",
			expectedOutput: "host",
		},
		{
			name:           "Network scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper",
			expectedOutput: "host",
		},
		{
			name:           "Paging scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/pagingscraper",
			expectedOutput: "host",
		},
		{
			name:           "Processes scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper",
			expectedOutput: "host",
		},
		{
			name:           "Process scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper",
			expectedOutput: "host",
		},
		{
			name:           "System scraper",
			input:          "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/systemscraper",
			expectedOutput: "host",
		},

		// 6. Go -> "go"
		{
			name:           "Go runtime instrumentation",
			input:          "go.opentelemetry.io/contrib/instrumentation/runtime",
			expectedOutput: "go",
		},

		// 7. Node.js -> "nodejs"
		{
			name:           "Node.js runtime instrumentation",
			input:          "@opentelemetry/instrumentation-runtime-node",
			expectedOutput: "nodejs",
		},

		// 8. Java -> "java"
		{
			name:           "Java runtime telemetry",
			input:          "io.opentelemetry.runtime-telemetry-java8",
			expectedOutput: "java",
		},
		{
			name:           "Java SDK",
			input:          "io.opentelemetry.sdk",
			expectedOutput: "java",
		},
		{
			name:           "Java SDK logs",
			input:          "io.opentelemetry.sdk.logs",
			expectedOutput: "java",
		},
		{
			name:           "Java SDK trace",
			input:          "io.opentelemetry.sdk.trace",
			expectedOutput: "java",
		},

		// 9. Python -> "python"
		{
			name:           "Python requests instrumentation",
			input:          "opentelemetry.instrumentation.requests",
			expectedOutput: "http",
		},
		{
			name:           "Python system metrics instrumentation",
			input:          "opentelemetry.instrumentation.system_metrics",
			expectedOutput: "host",
		},

		// 10. .NET -> "dotnet"
		{
			name:           ".NET process instrumentation",
			input:          "OpenTelemetry.Instrumentation.Process",
			expectedOutput: "dotnet",
		},
		{
			name:           ".NET runtime instrumentation",
			input:          "OpenTelemetry.Instrumentation.Runtime",
			expectedOutput: "dotnet",
		},

		// 11. Keep original names (no transformation)
		{
			name:           "Connect RPC",
			input:          "connectrpc.com/otelconnect",
			expectedOutput: "connectrpc.com/otelconnect",
		},
		{
			name:           "Flagd service",
			input:          "flagd",
			expectedOutput: "flagd",
		},
		{
			name:           "JSON evaluator",
			input:          "jsonEvaluator",
			expectedOutput: "jsonEvaluator",
		},
		{
			name:           "Locust file",
			input:          "locustfile",
			expectedOutput: "locustfile",
		},
		{
			name:           "Nginx service",
			input:          "nginx",
			expectedOutput: "nginx",
		},
		{
			name:           "Product catalog service",
			input:          "product-catalog",
			expectedOutput: "product-catalog",
		},
		{
			name:           "Root scope",
			input:          "root",
			expectedOutput: "root",
		},
		{
			name:           "Span metrics connector",
			input:          "spanmetricsconnector",
			expectedOutput: "spanmetricsconnector",
		},

		// Edge cases
		{
			name:           "Empty scope name",
			input:          "",
			expectedOutput: "",
		},
		{
			name:           "Unknown scope pattern",
			input:          "some.unknown.scope.pattern",
			expectedOutput: "some.unknown.scope.pattern",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := transformer.Transform(tc.input)
			if result != tc.expectedOutput {
				t.Errorf("Transform(%q) = %q, want %q", tc.input, result, tc.expectedOutput)
			}
		})
	}
}
