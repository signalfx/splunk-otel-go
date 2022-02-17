module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql

go 1.16

require (
	github.com/signalfx/splunk-otel-go v0.7.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.4.1
	go.opentelemetry.io/otel/trace v1.4.1
)

replace github.com/signalfx/splunk-otel-go => ../../../../
