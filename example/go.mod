module example

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/signalfx/splunk-otel-go/distro v0.8.0
	github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp v0.8.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.32.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
)

replace github.com/signalfx/splunk-otel-go/distro => ../distro

replace github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp => ../instrumentation/net/http/splunkhttp
