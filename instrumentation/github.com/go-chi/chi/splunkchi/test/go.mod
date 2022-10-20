module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi/test

go 1.17

require (
	github.com/go-chi/chi v1.5.4
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi v1.1.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/sdk v1.11.0
	go.opentelemetry.io/otel/trace v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v1.1.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.1.0 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
