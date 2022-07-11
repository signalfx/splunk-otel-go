module github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka/test

go 1.16

require (
	github.com/confluentinc/confluent-kafka-go v1.9.0
	github.com/ory/dockertest/v3 v3.9.1
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka v1.0.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.8.0
	go.uber.org/goleak v1.1.12
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../../internal/
)
