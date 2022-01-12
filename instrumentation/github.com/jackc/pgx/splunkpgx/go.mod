module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx

go 1.16

require (
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.14.1
	github.com/signalfx/splunk-otel-go v0.7.0
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.7.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
)
