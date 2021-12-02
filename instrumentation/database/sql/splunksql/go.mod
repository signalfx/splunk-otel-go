module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql

go 1.16

require (
	github.com/signalfx/splunk-otel-go v0.6.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/trace v1.2.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../
