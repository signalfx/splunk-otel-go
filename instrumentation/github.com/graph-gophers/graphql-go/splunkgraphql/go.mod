module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql

go 1.16

require (
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.0.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/trace v1.8.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
