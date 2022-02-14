module github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka/test

go 1.16

require (
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/ory/dockertest/v3 v3.8.1
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka v0.7.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.4.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.4.0
	go.uber.org/goleak v1.1.12
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka => ../
)
