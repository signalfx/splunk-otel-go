module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql

go 1.16

require (
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/signalfx/splunk-otel-go v0.7.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../..
