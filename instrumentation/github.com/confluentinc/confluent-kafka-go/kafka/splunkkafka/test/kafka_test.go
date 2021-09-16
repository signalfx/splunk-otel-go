// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows

package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	traceapi "go.opentelemetry.io/otel/trace"
)

var (
	testGroupID = "gotest"
	testTopic   = "gotest"
)

func TestConsumerChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Test consuming via the Events channel by artificially sending messages.

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	c, err := splunkkafka.NewConsumer(&kafka.ConfigMap{
		"go.events.channel.enable": true, // required for the events channel to be turned on
		"group.id":                 testGroupID,
		"session.timeout.ms":       10,
		"enable.auto.offset.store": false,
		// "debug":                    "all",
	}, splunkkafka.WithTracerProvider(tp))
	require.NoError(t, err)

	err = c.Subscribe(testTopic, nil)
	require.NoError(t, err)

	go func() {
		c.Consumer.Events() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: 1,
				Offset:    1,
			},
			Key:   []byte("key1"),
			Value: []byte("value1"),
		}
		c.Consumer.Events() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: 1,
				Offset:    2,
			},
			Key:   []byte("key2"),
			Value: []byte("value2"),
		}
	}()

	msg1 := (<-c.Events()).(*kafka.Message)
	assert.Equal(t, []byte("key1"), msg1.Key)
	msg2 := (<-c.Events()).(*kafka.Message)
	assert.Equal(t, []byte("key2"), msg2.Key)

	c.Close()
	// wait for the events channel to be closed
	<-c.Events()

	spans := sr.Ended()
	assert.Len(t, spans, 2)
	expectedName := fmt.Sprintf("%s receive", testTopic)
	for i, s := range spans {
		assert.Equal(t, expectedName, s.Name())
		assert.Equal(t, traceapi.SpanKindConsumer, s.SpanKind())
		attrs := s.Attributes()
		assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
		assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
		assert.Contains(t, attrs, semconv.MessagingDestinationKey.String(testTopic))
		assert.Contains(t, attrs, semconv.MessagingOperationReceive)
		offset := kafka.Offset(i + 1)
		assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String(offset.String()))
		key := fmt.Sprintf("key%d", i+1)
		assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(key))
		assert.Contains(t, attrs, semconv.MessagingKafkaConsumerGroupKey.String(testGroupID))
		assert.Contains(t, attrs, semconv.MessagingKafkaPartitionKey.Int64(1))
	}
}

/*
to run the integration test locally:
    docker network create confluent
    docker run --rm \
        --name zookeeper \
        --network confluent \
        -p 2181:2181 \
        -e ZOOKEEPER_CLIENT_PORT=2181 \
        confluentinc/cp-zookeeper:5.0.0
    docker run --rm \
        --name kafka \
        --network confluent \
        -p 9092:9092 \
        -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
        -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
        -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
        -e KAFKA_CREATE_TOPICS=gotest:1:1 \
        -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
        confluentinc/cp-kafka:5.0.0
*/

func TestConsumerPoll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	if _, ok := os.LookupEnv("INTEGRATION"); !ok {
		t.Skip("to enable integration test, set the INTEGRATION environment variable")
	}

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	prop := propagation.TraceContext{}

	// first write a message to the topic

	p, err := splunkkafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   "127.0.0.1:9092",
		"go.delivery.reports": true,
	}, splunkkafka.WithTracerProvider(tp), splunkkafka.WithPropagator(prop))
	require.NoError(t, err)
	delivery := make(chan kafka.Event, 1)
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &testTopic,
			Partition: 0,
		},
		Key:   []byte("key2"),
		Value: []byte("value2"),
	}, delivery)
	require.NoError(t, err)
	msg1, _ := (<-delivery).(*kafka.Message)
	p.Close()

	// next attempt to consume the message

	c, err := splunkkafka.NewConsumer(&kafka.ConfigMap{
		"group.id":                 testGroupID,
		"bootstrap.servers":        "127.0.0.1:9092",
		"socket.timeout.ms":        6000,
		"session.timeout.ms":       1000,
		"enable.auto.offset.store": false,
		// "debug":                    "all",
	}, splunkkafka.WithTracerProvider(tp), splunkkafka.WithPropagator(prop))
	require.NoError(t, err)

	require.NoError(t, c.Assign([]kafka.TopicPartition{
		{Topic: &testTopic, Partition: 0, Offset: msg1.TopicPartition.Offset},
	}))

	msg2, _ := c.Poll(3000).(*kafka.Message)
	assert.Equal(t, msg1.String(), msg2.String())

	c.Close()

	spans := sr.Ended()
	require.Len(t, spans, 2)
	producerSpan, consumerSpan := spans[0], spans[1]
	// they should be linked via headers
	assert.Equal(t, producerSpan.SpanContext().TraceID(), consumerSpan.SpanContext().TraceID())

	assert.Equal(t, fmt.Sprintf("%s send", testTopic), producerSpan.Name())
	assert.Equal(t, traceapi.SpanKindProducer, producerSpan.SpanKind())
	attrs := producerSpan.Attributes()
	assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
	assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
	assert.Contains(t, attrs, semconv.MessagingDestinationKey.String(testTopic))
	assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String("0"))
	assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String("key2"))
	assert.Contains(t, attrs, semconv.MessagingKafkaPartitionKey.Int64(0))

	assert.Equal(t, fmt.Sprintf("%s receive", testTopic), consumerSpan.Name())
	assert.Equal(t, traceapi.SpanKindConsumer, consumerSpan.SpanKind())
	attrs = consumerSpan.Attributes()
	assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
	assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
	assert.Contains(t, attrs, semconv.MessagingDestinationKey.String(testTopic))
	assert.Contains(t, attrs, semconv.MessagingOperationReceive)
	assert.Contains(t, attrs, semconv.MessagingKafkaConsumerGroupKey.String(testGroupID))
	assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String("key2"))
	assert.Contains(t, attrs, semconv.MessagingKafkaPartitionKey.Int64(0))
}
