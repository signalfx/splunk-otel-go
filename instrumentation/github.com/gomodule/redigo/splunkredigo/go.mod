module github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo

go 1.16

require (
	github.com/gomodule/redigo v1.8.8
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.5.0
	go.opentelemetry.io/otel/trace v1.5.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
