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

//go:build cgo && linux
// +build cgo,linux

package splunkkafka

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"
	"go.uber.org/goleak"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
)

var (
	key, val = []byte("test key"), []byte("test value")

	testGroupID = "gotest"
	testTopic   = "gotest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("Skipping running heavy integration test in short mode.")
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	confNet, err := pool.CreateNetwork("confluent")
	if err != nil {
		log.Fatalf("Could not create docker network: %v", err)
	}

	zkRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-zookeeper",
		Tag:        "6.2.0",
		NetworkID:  confNet.Network.ID,
		Hostname:   "zookeeper",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"2181/tcp": {{HostIP: "zookeeper", HostPort: "2181/tcp"}},
		},
		Env: []string{
			"ZOOKEEPER_CLIENT_PORT=2181",
			"ZOOKEEPER_TICK_TIME=2000",
		},
	})
	if err != nil {
		log.Fatalf("Could not create zookeeper: %v", err)
	}

	// Wait for the Kafka to come up using an exponential-backoff retry.
	if err = pool.Retry(func() error {
		_, dialErr := net.Dial("tcp", "localhost:2181")
		return dialErr
	}); err != nil {
		log.Fatalf("Could not connect to Kafka broker: %v", err)
	}

	kafkaRes, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "confluentinc/cp-kafka",
		Tag:        "6.2.0",
		NetworkID:  confNet.Network.ID,
		Hostname:   "broker",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"29092/tcp": {{HostIP: "broker", HostPort: "29092/tcp"}},
			"9092/tcp":  {{HostIP: "localhost", HostPort: "9092/tcp"}},
		},
		Env: []string{
			"KAFKA_BROKER_ID=1",
			"KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://broker:29092,PLAINTEXT_HOST://localhost:9092",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
			fmt.Sprintf("KAFKA_CREATE_TOPICS=%s:1:1", testTopic),
		},
	})
	if err != nil {
		log.Fatalf("Could not create kakfa container: %v", err)
	}

	// Wait for the Kafka to come up using an exponential-backoff retry.
	if err = pool.Retry(verifyCanProduceToKafka); err != nil {
		log.Fatalf("Could not connect to Kafka broker: %v", err)
	}

	code := m.Run()

	// Run sequentially becauase os.Exit will skip a defer.
	if err := pool.Purge(kafkaRes); err != nil {
		log.Fatalf("Could not purge kafka: %v", err)
	}
	if err := pool.Purge(zkRes); err != nil {
		log.Fatalf("Could not purge zookeeper: %v", err)
	}
	if err := confNet.Close(); err != nil {
		log.Fatalf("Could not remove network: %v", err)
	}

	os.Exit(code)
}

func verifyCanProduceToKafka() error {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
	})
	if err != nil {
		return err
	}
	defer p.Close()

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &testTopic, Partition: 0},
		Key:            key,
		Value:          val,
	}, nil)
	if err != nil {
		return err
	}

	e := <-p.Events()
	if m, ok := e.(*kafka.Message); !ok {
		return fmt.Errorf("event: %s", e.String())
	} else if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	return nil
}

func TestChannelBasedProducer(t *testing.T) {
	defer goleak.VerifyNone(t)

	partition := int32(0)
	sr, opts := newFixtures()
	p := newProducer(t, opts...)

	go func() {
		p.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: partition,
			},
			Key:   key,
			Value: val,
		}
	}()

	// Wait for the delivery report goroutine to finish.
	sent := requireEventIsMessage(t, <-p.Events())
	require.NoError(t, sent.TopicPartition.Error)

	// Ensure all Producer operations complete and all spans are done.
	for {
		if remaining := p.Flush(100); remaining == 0 {
			break
		}
	}
	// Decommission the producer, ensure it is not used anymore.
	p.Close()

	recv := consumeMessage(t, kafka.TopicPartition{
		Topic:     &testTopic,
		Partition: partition,
		Offset:    sent.TopicPartition.Offset,
	}, opts...)

	assert.Equal(t, sent.String(), recv.String())

	spans := sr.Ended()
	require.Len(t, spans, 2)
	pSpan, cSpan := spans[0], spans[1]
	// The should be linked via propagated headers.
	assert.Equal(t, pSpan.SpanContext().TraceID(), cSpan.SpanContext().TraceID())
	assertProducerSpan(t, pSpan)
	assertConsumerSpan(t, cSpan)
}

func TestFunctionBasedProducer(t *testing.T) {
	defer goleak.VerifyNone(t)

	partition := int32(0)
	sr, opts := newFixtures()
	p := newProducer(t, opts...)

	deliveryCh := make(chan kafka.Event, 1)
	err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &testTopic,
			Partition: partition,
		},
		Key:   key,
		Value: val,
	}, deliveryCh)
	require.NoError(t, err)
	sent := requireEventIsMessage(t, <-deliveryCh)
	require.NoError(t, sent.TopicPartition.Error)

	// Ensure all Producer operations complete and all spans are done.
	for {
		if remaining := p.Flush(100); remaining == 0 {
			break
		}
	}
	// Decommission the producer, ensure it is not used anymore.
	p.Close()

	recv := consumeMessage(t, kafka.TopicPartition{
		Topic:     &testTopic,
		Partition: partition,
		Offset:    sent.TopicPartition.Offset,
	}, opts...)

	assert.Equal(t, sent.String(), recv.String())

	spans := sr.Ended()
	require.Len(t, spans, 2)
	pSpan, cSpan := spans[0], spans[1]
	// The should be linked via propagated headers.
	assert.Equal(t, pSpan.SpanContext().TraceID(), cSpan.SpanContext().TraceID())
	assertProducerSpan(t, pSpan)
	assertConsumerSpan(t, cSpan)
}

func newFixtures() (*tracetest.SpanRecorder, []splunkkafka.Option) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	prop := propagation.TraceContext{}
	return sr, []splunkkafka.Option{splunkkafka.WithTracerProvider(tp), splunkkafka.WithPropagator(prop)}
}

func newProducer(t *testing.T, opts ...splunkkafka.Option) *splunkkafka.Producer {
	p, err := splunkkafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   "127.0.0.1:9092",
		"go.delivery.reports": true,
	}, opts...)
	require.NoError(t, err)
	return p
}

func newConsumer(t *testing.T, opts ...splunkkafka.Option) *splunkkafka.Consumer {
	c, err := splunkkafka.NewConsumer(&kafka.ConfigMap{
		"group.id":          testGroupID,
		"bootstrap.servers": "127.0.0.1:9092",
		"socket.timeout.ms": 6000,
	}, opts...)
	require.NoError(t, err)
	return c
}

func consumeMessage(t *testing.T, tp kafka.TopicPartition, opts ...splunkkafka.Option) *kafka.Message {
	c := newConsumer(t, opts...)
	require.NoError(t, c.Assign([]kafka.TopicPartition{tp}))
	recv := requireEventIsMessage(t, func() kafka.Event {
		for {
			if e := c.Poll(100); e != nil {
				return e
			}
		}
	}())
	assert.NoError(t, recv.TopicPartition.Error)
	_, err := c.CommitMessage(recv)
	assert.NoError(t, err)
	assert.NoError(t, c.Unassign())

	// Ensure all Consumer operations complete and all spans are done.
	c.Close()
	return recv
}

func assertProducerSpan(t *testing.T, span trace.ReadOnlySpan) {
	assert.Equal(t, fmt.Sprintf("%s send", testTopic), span.Name())
	assert.Equal(t, traceapi.SpanKindProducer, span.SpanKind())
	assert.Equal(t, splunkkafka.Version(), span.InstrumentationLibrary().Version)
	attrs := span.Attributes()
	assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
	assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
	assert.Contains(t, attrs, semconv.MessagingDestinationNameKey.String(testTopic))
	assert.Contains(t, attrs, semconv.MessagingOperationPublish)
	assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String("0"))
	assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(string(key)))
	assert.Contains(t, attrs, semconv.MessagingKafkaDestinationPartitionKey.Int64(0))
}

func assertConsumerSpan(t *testing.T, span trace.ReadOnlySpan) {
	assert.Equal(t, fmt.Sprintf("%s receive", testTopic), span.Name())
	assert.Equal(t, traceapi.SpanKindConsumer, span.SpanKind())
	assert.Equal(t, splunkkafka.Version(), span.InstrumentationLibrary().Version)
	attrs := span.Attributes()
	assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
	assert.Contains(t, attrs, semconv.MessagingSourceKindTopic)
	assert.Contains(t, attrs, semconv.MessagingSourceNameKey.String(testTopic))
	assert.Contains(t, attrs, semconv.MessagingOperationReceive)
	assert.Contains(t, attrs, semconv.MessagingKafkaConsumerGroupKey.String(testGroupID))
	assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(string(key)))
	assert.Contains(t, attrs, semconv.MessagingKafkaSourcePartitionKey.Int64(0))
}

func requireEventIsMessage(t *testing.T, e kafka.Event) *kafka.Message {
	m, ok := e.(*kafka.Message)
	require.Truef(t, ok, "invalid response from Kafka: %v", e)
	require.NoError(t, m.TopicPartition.Error)
	return m
}
