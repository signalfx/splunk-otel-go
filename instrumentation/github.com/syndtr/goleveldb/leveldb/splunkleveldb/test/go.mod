module github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb/test

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb v0.8.0
	github.com/stretchr/testify v1.7.1
	github.com/syndtr/goleveldb v1.0.0
	go.opentelemetry.io/otel v1.6.2
	go.opentelemetry.io/otel/sdk v1.6.2
	go.opentelemetry.io/otel/trace v1.6.2
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../../internal/
)
