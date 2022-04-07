module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.2
	go.opentelemetry.io/otel/trace v1.6.2
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../internal/
)
