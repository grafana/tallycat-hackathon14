package schema

type ScopeTransformer struct {
	transformations map[string]string
}

func NewScopeTransformer() *ScopeTransformer {
	transformer := &ScopeTransformer{
		transformations: make(map[string]string),
	}
	transformer.initializeTransformations()

	return transformer
}

func (st *ScopeTransformer) Transform(scopeName string) string {
	if scopeName == "" {
		return scopeName
	}

	if transformed, exists := st.transformations[scopeName]; exists {
		return transformed
	}

	return scopeName
}

func (st *ScopeTransformer) initializeTransformations() {
	// 1. OpenTelemetry Collector Components -> "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/exporter/exporterhelper"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/processor/batchprocessor"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/processor/memorylimiterprocessor"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/processor/processorhelper"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/receiver/receiverhelper"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/scraper/scraperhelper"] = "opentelemetry-collector"
	st.transformations["go.opentelemetry.io/collector/service"] = "opentelemetry-collector"

	// 2. Docker -> "docker"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"] = "docker"

	// 3. HTTP -> "http"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver"] = "http"
	st.transformations["go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"] = "http"
	st.transformations["@opentelemetry/instrumentation-http"] = "http"
	st.transformations["io.opentelemetry.exporters.otlp-http"] = "http"
	st.transformations["opentelemetry.instrumentation.requests"] = "http"

	// 4. gRPC -> "grpc"
	st.transformations["go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"] = "grpc"
	st.transformations["io.opentelemetry.grpc-1.6"] = "grpc"
	st.transformations["opentelemetry.instrumentation.grpc"] = "grpc"

	// 5. Host Metrics -> "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/diskscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/loadscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/memoryscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/pagingscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processscraper"] = "host"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/systemscraper"] = "host"
	st.transformations["opentelemetry.instrumentation.system_metrics"] = "host"

	// 6. Go -> "go"
	st.transformations["go.opentelemetry.io/contrib/instrumentation/runtime"] = "go"

	// 7. Node.js -> "nodejs"
	st.transformations["@opentelemetry/instrumentation-runtime-node"] = "nodejs"

	// 8. Java -> "java"
	st.transformations["io.opentelemetry.runtime-telemetry-java8"] = "java"
	st.transformations["io.opentelemetry.sdk"] = "java"
	st.transformations["io.opentelemetry.sdk.logs"] = "java"
	st.transformations["io.opentelemetry.sdk.trace"] = "java"

	// 9. .NET -> "dotnet"
	st.transformations["OpenTelemetry.Instrumentation.Process"] = "dotnet"
	st.transformations["OpenTelemetry.Instrumentation.Runtime"] = "dotnet"

	// 10. Nginx -> "nginx"
	st.transformations["github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"] = "nginx"

	// All others keep their original names (no transformation needed)
}
