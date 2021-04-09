module github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp

go 1.14

replace go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => ../otelhttp

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.19.0
	go.opentelemetry.io/otel/oteltest v0.19.0
	go.opentelemetry.io/otel/trace v0.19.0
)
