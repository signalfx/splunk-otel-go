module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/test

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-20210921155913-50ba051311e1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)

replace github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../
