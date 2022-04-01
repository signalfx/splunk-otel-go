module github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql/test

go 1.16

require (
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel/sdk v1.6.1
	go.opentelemetry.io/otel/trace v1.6.1
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
