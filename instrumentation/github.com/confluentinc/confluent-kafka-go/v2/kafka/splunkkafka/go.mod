module github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka

go 1.25.0

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.15.0
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.34.0
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../../internal
