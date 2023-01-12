module github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq

go 1.18

require (
	github.com/lib/pq v1.10.7
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.2.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v1.2.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.2.0 // indirect
	go.opentelemetry.io/otel v1.11.2 // indirect
	go.opentelemetry.io/otel/trace v1.11.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../..
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../../../../database/sql/splunksql
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
