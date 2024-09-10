module github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb

go 1.21

require (
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.19.0
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/buntdb v1.3.2
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/tidwall/gjson v1.17.3 // indirect
	github.com/tidwall/grect v0.1.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/rtred v0.1.2 // indirect
	github.com/tidwall/tinyqueue v0.1.1 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
