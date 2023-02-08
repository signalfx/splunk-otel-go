module github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb

go 1.18

require (
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.3.0
	github.com/stretchr/testify v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	go.opentelemetry.io/otel v1.13.0
	go.opentelemetry.io/otel/trace v1.13.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
