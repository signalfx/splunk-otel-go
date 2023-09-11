module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/test

go 1.19

require (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.7.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.17.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.opentelemetry.io/otel/sdk/metric v0.39.0
	go.opentelemetry.io/otel/trace v1.17.0
	go.uber.org/goleak v1.2.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.7.0 // indirect
	go.opentelemetry.io/otel/metric v1.17.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
