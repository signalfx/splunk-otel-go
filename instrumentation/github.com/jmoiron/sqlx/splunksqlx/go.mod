module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx

go 1.16

require (
	github.com/jmoiron/sqlx v1.3.5
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.0.0
	github.com/stretchr/testify v1.8.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
