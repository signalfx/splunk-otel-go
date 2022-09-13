module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi

go 1.17

require (
	github.com/go-chi/chi v1.5.4
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.1.0
	go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/trace v1.10.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/signalfx/splunk-otel-go v1.1.0 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
