module github.com/signalfx/splunk-otel-go

go 1.14

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/propagators v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.19.0
	go.opentelemetry.io/otel/sdk v0.19.0
)
