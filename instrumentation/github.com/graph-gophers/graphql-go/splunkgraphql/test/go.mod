module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql/test

go 1.21

require (
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql v1.14.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.14.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
)
