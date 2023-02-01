module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi

go 1.18

require (
	github.com/go-chi/chi v1.5.4
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.3.0
	go.opentelemetry.io/otel v1.12.0
	go.opentelemetry.io/otel/trace v1.12.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/signalfx/splunk-otel-go v1.3.0 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
