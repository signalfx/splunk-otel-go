module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql

go 1.18

require (
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.1.0
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/trace v1.11.2
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/signalfx/splunk-otel-go v1.1.0 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
