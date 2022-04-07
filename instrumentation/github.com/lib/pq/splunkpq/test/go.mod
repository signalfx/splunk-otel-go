module github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq/test

go 1.16

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)

require (
	github.com/ory/dockertest/v3 v3.8.1
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.8.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.2
	go.opentelemetry.io/otel/sdk v1.6.1
)
