module github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter/test

go 1.16

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter v1.0.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/sdk v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter => ../
)
