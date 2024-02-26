module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi

go 1.20

require (
	github.com/go-chi/chi v1.5.5
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.14.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
