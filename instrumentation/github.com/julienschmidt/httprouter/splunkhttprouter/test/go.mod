module github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter/test

go 1.20

require (
	github.com/julienschmidt/httprouter v1.3.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter v1.13.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.48.0
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/sdk v1.23.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	go.opentelemetry.io/otel/trace v1.23.1 // indirect
	golang.org/x/sys v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter => ../
