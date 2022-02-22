module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx

go 1.16

require (
	github.com/jackc/pgx/v4 v4.15.0
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.7.0
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
)
