module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go

go 1.15

require (
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.27.0
	k8s.io/client-go v0.22.4
)
