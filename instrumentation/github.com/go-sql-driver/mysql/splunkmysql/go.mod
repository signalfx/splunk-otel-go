module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql

go 1.15

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/signalfx/splunk-otel-go v0.6.0
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go/ => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
)
