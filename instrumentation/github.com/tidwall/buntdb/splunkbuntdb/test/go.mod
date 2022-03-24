module github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb/test

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb v0.7.0
	github.com/stretchr/testify v1.7.1
	github.com/tidwall/buntdb v1.2.9
	go.opentelemetry.io/otel v1.6.0
	go.opentelemetry.io/otel/sdk v1.5.0
	go.opentelemetry.io/otel/trace v1.6.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
