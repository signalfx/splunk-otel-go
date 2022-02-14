module github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb

go 1.16

require (
	github.com/signalfx/splunk-otel-go v0.7.0
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0
	go.opentelemetry.io/otel v1.4.0
	go.opentelemetry.io/otel/trace v1.4.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../../../
