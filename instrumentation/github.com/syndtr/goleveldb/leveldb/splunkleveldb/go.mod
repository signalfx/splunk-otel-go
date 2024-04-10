module github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb

go 1.21

require (
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.15.0
	github.com/stretchr/testify v1.9.0
	github.com/syndtr/goleveldb v1.0.0
	go.opentelemetry.io/otel v1.25.0
	go.opentelemetry.io/otel/trace v1.25.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/metric v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
