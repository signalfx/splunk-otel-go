module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql

go 1.15

require (
	github.com/signalfx/splunk-otel-go v0.5.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/trace v1.0.0-RC2
)

replace github.com/signalfx/splunk-otel-go => ../../../../
