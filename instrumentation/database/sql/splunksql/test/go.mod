module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/test

go 1.15

require (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.0.0-RC2.0.20210812161231-a8bb0bf89f3b // indirect
	go.opentelemetry.io/otel/sdk v1.0.0-RC2.0.20210812161231-a8bb0bf89f3b
	go.opentelemetry.io/otel/trace v1.0.0-RC2
)

replace github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../
