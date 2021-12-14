module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql/test

go 1.16

require (
	github.com/graph-gophers/graphql-go v1.2.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql => ../
)
