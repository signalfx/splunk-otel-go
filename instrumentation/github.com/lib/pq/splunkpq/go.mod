module github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq

go 1.16

require (
	github.com/lib/pq v1.10.6
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.9.0
	github.com/stretchr/testify v1.7.2
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
