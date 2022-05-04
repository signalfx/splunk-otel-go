module example

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/signalfx/splunk-otel-go/distro v0.8.0
	github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp v0.8.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.32.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

replace github.com/signalfx/splunk-otel-go/distro => ../distro

replace github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp => ../instrumentation/net/http/splunkhttp
