module github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb/test

go 1.18

require (
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb v1.1.0
	github.com/stretchr/testify v1.8.1
	github.com/tidwall/buntdb v1.2.10
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/sdk v1.11.1
	go.opentelemetry.io/otel/trace v1.11.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v1.1.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.1.0 // indirect
	github.com/tidwall/btree v1.5.0 // indirect
	github.com/tidwall/gjson v1.14.3 // indirect
	github.com/tidwall/grect v0.1.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/rtred v0.1.2 // indirect
	github.com/tidwall/tinyqueue v0.1.1 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
