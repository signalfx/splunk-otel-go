module github.com/signalfx/splunk-otel-go/instrumentation/internal

go 1.16

replace github.com/signalfx/splunk-otel-go => ../../

require (
	github.com/signalfx/splunk-otel-go v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)
