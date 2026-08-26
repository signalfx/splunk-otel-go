// Deprecated: this module is no longer supported.
// See https://github.com/signalfx/splunk-otel-go/issues/4397 for more details.
module github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo

go 1.25.0

require (
	github.com/gomodule/redigo v1.9.3
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.35.0
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/otel v1.46.0
	go.opentelemetry.io/otel/trace v1.46.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
