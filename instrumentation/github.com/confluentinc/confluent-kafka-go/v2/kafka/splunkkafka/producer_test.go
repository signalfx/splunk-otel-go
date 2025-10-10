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

//go:build cgo && (linux || darwin)

package splunkkafka

import (
	"context"
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func TestNewProducerType(t *testing.T) {
	p, err := NewProducer(&kafka.ConfigMap{})
	require.NoError(t, err)
	assert.IsType(t, &Producer{}, p)
}

func TestNewProducerReturnsError(t *testing.T) {
	// go.batch.producer not being a bool type will cause an error.
	_, err := NewProducer(&kafka.ConfigMap{"go.batch.producer": 1})
	require.Error(t, err)
}

func TestWrapProducerType(t *testing.T) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{})
	require.NoError(t, err)
	assert.IsType(t, Producer{}, *WrapProducer(p))
}

func TestProducerEventsChanCreated(t *testing.T) {
	chSize := 500
	p, err := NewProducer(&kafka.ConfigMap{
		"go.produce.channel.size": chSize,
	})
	require.NoError(t, err)
	assert.NotNil(t, p.ProduceChannel())
	assert.Equal(t, chSize, cap(p.ProduceChannel()))
}

func TestProducerChannelSpan(t *testing.T) {
	sr := make(spanRecorder)
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{start: sr.start}
		},
	}
	prop := propagation.TraceContext{}
	p, err := NewProducer(&kafka.ConfigMap{}, WithTracerProvider(tp), WithPropagator(prop))
	require.NoError(t, err)

	keys := []string{"key1", "key2"}
	produceChannel := make(chan *kafka.Message, len(keys))
	p.produceChannel = p.traceProduceChannel(produceChannel)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
		Remote:  true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	for _, k := range keys {
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: 1,
				Offset:    1,
			},
			Key:   []byte(k),
			Value: []byte("value"),
		}
		// Test that context is propagated.
		prop.Inject(ctx, NewMessageCarrier(msg))
		p.ProduceChannel() <- msg
		got := <-produceChannel
		assert.Equal(t, []byte(k), got.Key)
	}

	require.Len(t, sr, 1)
	expectedName := fmt.Sprintf("%s send", testTopic)
	require.Contains(t, sr, expectedName)
	spans := sr.get(expectedName)
	assert.Len(t, spans, len(keys))
	for i, record := range spans {
		assert.Equal(t, sc, record.SpanContext)
		assert.Equal(t, record.SpanConfig.SpanKind(), trace.SpanKindProducer)
		attrs := record.SpanConfig.Attributes()
		assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
		assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
		assert.Contains(t, attrs, semconv.MessagingDestinationNameKey.String(testTopic))
		assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String("1"))
		assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(keys[i]))
		assert.Contains(t, attrs, semconv.MessagingKafkaDestinationPartitionKey.Int64(1))
	}
}

func TestProduceSpan(t *testing.T) {
	sr := make(spanRecorder)
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{start: sr.start}
		},
	}
	prop := propagation.TraceContext{}
	commonAttr := attribute.String("key", "value")
	p, err := NewProducer(
		&kafka.ConfigMap{},
		WithTracerProvider(tp),
		WithPropagator(prop),
		WithAttributes([]attribute.KeyValue{commonAttr}),
	)
	require.NoError(t, err)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
		Remote:  true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	keys := []string{"key1", "key2"}
	for _, k := range keys {
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: 1,
				Offset:    1,
			},
			Key:   []byte(k),
			Value: []byte("value"),
		}
		// Test that context is propagated.
		prop.Inject(ctx, NewMessageCarrier(msg))
		assert.NoError(t, p.Produce(msg, nil))
	}

	require.Len(t, sr, 1)
	expectedName := fmt.Sprintf("%s send", testTopic)
	require.Contains(t, sr, expectedName)
	spans := sr.get(expectedName)
	assert.Len(t, spans, len(keys))
	for i, record := range spans {
		assert.Equal(t, sc, record.SpanContext)
		assert.Equal(t, record.SpanConfig.SpanKind(), trace.SpanKindProducer)
		attrs := record.SpanConfig.Attributes()
		assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
		assert.Contains(t, attrs, commonAttr)
		assert.Contains(t, attrs, semconv.MessagingDestinationKindTopic)
		assert.Contains(t, attrs, semconv.MessagingDestinationNameKey.String(testTopic))
		assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String("1"))
		assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(keys[i]))
		assert.Contains(t, attrs, semconv.MessagingKafkaDestinationPartitionKey.Int64(1))
	}
}
