module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi

go 1.16

require (
	github.com/go-chi/chi v1.5.4
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.6.1
	go.opentelemetry.io/otel/trace v1.6.1
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
