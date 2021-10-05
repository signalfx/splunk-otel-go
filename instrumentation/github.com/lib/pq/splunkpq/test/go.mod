module github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq/test

go 1.15

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq => ../
)

require (
	github.com/ory/dockertest/v3 v3.8.0
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-00010101000000-000000000000
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq v0.0.0-00010101000000-000000000000
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/sdk v1.0.0
)
