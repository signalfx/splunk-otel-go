module github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter

go 1.16

require (
	github.com/julienschmidt/httprouter v1.3.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.30.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../..
