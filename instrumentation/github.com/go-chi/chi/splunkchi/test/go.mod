module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi/test

go 1.16

require (
	github.com/go-chi/chi v1.5.4
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi v1.0.0
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
