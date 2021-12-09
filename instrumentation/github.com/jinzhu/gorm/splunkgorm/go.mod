module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm

go 1.16

require (
	github.com/jinzhu/gorm v1.9.16
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)

replace github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
