module github.com/signalfx/splunk-otel-go

go 1.16

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/propagators/b3 v1.2.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/exporters/jaeger v1.2.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)
