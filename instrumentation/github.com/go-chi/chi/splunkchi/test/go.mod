module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi/test

go 1.20

require (
	github.com/go-chi/chi v1.5.5
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi v1.10.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.20.0
	go.opentelemetry.io/otel/sdk v1.20.0
	go.opentelemetry.io/otel/trace v1.20.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.10.0 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
)
