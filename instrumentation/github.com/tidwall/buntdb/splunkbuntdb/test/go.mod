module github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb/test

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/signalfx/signalfx-go-tracing/contrib/tidwall/buntdb/splunkbuntdb v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/buntdb v1.2.7
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/trace v1.2.0
)

// replace github.com/signalfx/signalfx-go-tracing/contrib/tidwall/buntdb/splunkbuntdb => /Users/trojek/github/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb
replace github.com/signalfx/signalfx-go-tracing/contrib/tidwall/buntdb/splunkbuntdb => ../
