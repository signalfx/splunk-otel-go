module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi/test

go 1.24.0

require (
	github.com/go-chi/chi v1.5.5
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi v1.28.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.28.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
)
