module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/test

go 1.15

require (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.0.0
	go.opentelemetry.io/otel/sdk v1.0.0-RC3
	go.opentelemetry.io/otel/trace v1.0.0
)

replace github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../
