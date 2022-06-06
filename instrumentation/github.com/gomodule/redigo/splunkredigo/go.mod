module github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo

go 1.16

require (
	github.com/gomodule/redigo v1.8.8
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.9.0
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
