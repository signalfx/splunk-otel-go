module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.0.0
	github.com/stretchr/testify v1.7.5
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../internal/
)
