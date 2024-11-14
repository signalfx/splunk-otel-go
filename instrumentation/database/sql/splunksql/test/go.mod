module github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/test

go 1.22.0

require (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.22.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/sdk/metric v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	go.uber.org/goleak v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
)
