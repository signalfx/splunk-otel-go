module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx/test

go 1.16

require (
	github.com/ory/dockertest/v3 v3.8.1
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.7.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx v0.7.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.0
	go.opentelemetry.io/otel/sdk v1.6.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
