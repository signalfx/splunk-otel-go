module github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka

go 1.23.0

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.10.0
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.26.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../../internal
