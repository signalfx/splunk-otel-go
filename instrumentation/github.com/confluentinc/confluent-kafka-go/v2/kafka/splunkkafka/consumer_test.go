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
	"sync"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const grpID = "test group ID"

var testTopic = "gotest"

type fnTracerProvider struct {
	noop.TracerProvider
	tracer func(string, ...trace.TracerOption) trace.Tracer
}

func (fn *fnTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return fn.tracer(name, opts...)
}

type fnTracer struct {
	noop.Tracer
	start func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

func (fn *fnTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return fn.start(ctx, name, opts...)
}

func TestNewConsumerCapturesGroupID(t *testing.T) {
	c, err := NewConsumer(&kafka.ConfigMap{"group.id": grpID})
	require.NoError(t, err)
	sConf := trace.NewSpanStartConfig(c.cfg.DefaultStartOpts...)
	assert.Contains(t, sConf.Attributes(), semconv.MessagingKafkaConsumerGroupKey.String(grpID))
}

func TestNewConsumerType(t *testing.T) {
	c, err := NewConsumer(&kafka.ConfigMap{"group.id": grpID})
	require.NoError(t, err)
	assert.IsType(t, &Consumer{}, c)
}

func TestNewConsumerReturnsError(t *testing.T) {
	// It is an error to not specify the group.id.
	_, err := NewConsumer(&kafka.ConfigMap{})
	require.Error(t, err)
}

func TestWrapConsumerType(t *testing.T) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{"group.id": grpID})
	require.NoError(t, err)
	assert.IsType(t, Consumer{}, *WrapConsumer(c))
}

func TestConsumerEventsChanCreated(t *testing.T) {
	chSize := 500
	c, err := NewConsumer(&kafka.ConfigMap{
		// required for the events channel to be turned on
		"go.events.channel.enable": true,
		"go.events.channel.size":   chSize,
		"group.id":                 grpID,
	})
	require.NoError(t, err)
	assert.NotNil(t, c.Events())
	assert.Equal(t, chSize, cap(c.Events()))
}

type spanRecord struct {
	SpanContext trace.SpanContext
	SpanConfig  trace.SpanConfig
}

type spanRecorder map[string][]spanRecord

func (s spanRecorder) start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	existing := s[name]
	existing = append(existing, spanRecord{
		SpanContext: trace.SpanContextFromContext(ctx),
		SpanConfig:  trace.NewSpanStartConfig(opts...),
	})
	s[name] = existing
	return noop.NewTracerProvider().Tracer("").Start(ctx, name, opts...)
}

func (s spanRecorder) get(name string) []spanRecord {
	return s[name]
}

func TestConsumerSpan(t *testing.T) {
	sr := make(spanRecorder)
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{start: sr.start}
		},
	}
	prop := propagation.TraceContext{}
	commonAttr := attribute.String("key", "value")
	c, err := NewConsumer(&kafka.ConfigMap{
		// required for the events channel to be turned on
		"go.events.channel.enable": true,
		"group.id":                 grpID,
	}, WithTracerProvider(tp), WithPropagator(prop), WithAttributes([]attribute.KeyValue{commonAttr}))
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
		//nolint: staticcheck // Ensure backwards support of deprecated API.
		c.Consumer.Events() <- msg
		got := (<-c.Events()).(*kafka.Message)
		assert.Equal(t, []byte(k), got.Key)
	}

	require.Len(t, sr, 1)
	expectedName := fmt.Sprintf("%s receive", testTopic)
	require.Contains(t, sr, expectedName)
	spans := sr.get(expectedName)
	assert.Len(t, spans, len(keys))
	for i, record := range spans {
		assert.Equal(t, sc, record.SpanContext)
		assert.Equal(t, record.SpanConfig.SpanKind(), trace.SpanKindConsumer)
		attrs := record.SpanConfig.Attributes()
		assert.Contains(t, attrs, semconv.MessagingSystemKey.String("kafka"))
		assert.Contains(t, attrs, commonAttr)
		assert.Contains(t, attrs, semconv.MessagingSourceKindTopic)
		assert.Contains(t, attrs, semconv.MessagingSourceNameKey.String(testTopic))
		assert.Contains(t, attrs, semconv.MessagingOperationReceive)
		assert.Contains(t, attrs, semconv.MessagingMessageIDKey.String("1"))
		assert.Contains(t, attrs, semconv.MessagingKafkaMessageKeyKey.String(keys[i]))
		assert.Contains(t, attrs, semconv.MessagingKafkaConsumerGroupKey.String(grpID))
		assert.Contains(t, attrs, semconv.MessagingKafkaSourcePartitionKey.Int64(1))
	}
}

func TestConsumerConcurrentConsuming(t *testing.T) {
	sr := make(spanRecorder)
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{start: sr.start}
		},
	}
	c, err := NewConsumer(&kafka.ConfigMap{
		// Required for the events channel to be turned on.
		"go.events.channel.enable": true,
		"group.id":                 grpID,
	}, WithTracerProvider(tp))
	require.NoError(t, err)

	key := "test key"
	var wg sync.WaitGroup

	// Seed message to the events channel.
	wg.Add(1)
	//nolint: staticcheck // Ensure backwards support of deprecated API.
	go func() {
		defer wg.Done()
		c.Consumer.Events() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &testTopic,
				Partition: 1,
				Offset:    1,
			},
			Key:   []byte(key),
			Value: []byte("value"),
		}
	}()

	// Consume from the events channel.
	wg.Add(1)
	go func() {
		defer wg.Done()
		got := (<-c.Events()).(*kafka.Message)
		assert.Equal(t, []byte(key), got.Key)
	}()

	// Poll concurrently to the events channel consuming.
	wg.Add(1)
	go func() {
		defer wg.Done()
		// We are synthetically generating messages on the events channel so
		// we do not actually expect any events to be returned here.
		_ = c.Poll(100)
	}()

	wg.Wait()
	assert.Len(t, sr, 1)
}
