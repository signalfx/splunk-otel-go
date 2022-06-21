module github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm

go 1.16

require (
	github.com/jinzhu/gorm v1.9.16
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.0.0
	github.com/stretchr/testify v1.7.4
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
