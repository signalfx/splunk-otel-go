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
// +build cgo
// +build linux darwin

package splunkkafka

import (
	"context"
	"fmt"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Consumer wraps a kafka.Consumer and traces its operations.
type Consumer struct {
	*kafka.Consumer
	cfg    config
	group  string
	events chan kafka.Event
	prev   trace.Span
}

// NewConsumer calls kafka.NewConsumer and wraps the resulting Consumer with
// tracing instrumentation.
func NewConsumer(conf *kafka.ConfigMap, opts ...Option) (*Consumer, error) {
	c, err := kafka.NewConsumer(conf)
	if err != nil {
		return nil, err
	}
	// The kafka Consumer does not expose this. Give a best effort to add it.
	var consumerGroup string
	cGrpVal, err := conf.Get("group.id", "")
	if err == nil {
		consumerGroup, _ = cGrpVal.(string)
	}
	wrapped := &Consumer{
		Consumer: c,
		group:    consumerGroup,
		cfg:      newConfig(opts...),
	}
	wrapped.events = wrapped.traceEventsChannel(c.Events())
	return wrapped, nil
}

// WrapConsumer wraps a kafka.Consumer so that any consumed events are traced.
func WrapConsumer(c *kafka.Consumer, opts ...Option) *Consumer {
	wrapped := &Consumer{
		Consumer: c,
		cfg:      newConfig(opts...),
	}
	wrapped.events = wrapped.traceEventsChannel(c.Events())
	return wrapped
}

func (c *Consumer) traceEventsChannel(in chan kafka.Event) chan kafka.Event {
	// in will be nil when consuming via the events channel is not enabled.
	if in == nil {
		return nil
	}

	out := make(chan kafka.Event, cap(in))
	go func() {
		defer close(out)
		for evt := range in {
			var next trace.Span

			// only trace messages
			if msg, ok := evt.(*kafka.Message); ok {
				next = c.startSpan(msg)
			}

			out <- evt

			if c.prev != nil {
				c.prev.End()
			}
			c.prev = next
		}
		// finish any remaining span
		if c.prev != nil {
			c.prev.End()
			c.prev = nil
		}
	}()

	return out
}

func (c *Consumer) startSpan(msg *kafka.Message) trace.Span {
	carrier := NewMessageCarrier(msg)
	psc := c.cfg.Propagator.Extract(context.Background(), carrier)

	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("kafka"),
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationKey.String(*msg.TopicPartition.Topic),
		semconv.MessagingOperationReceive,
		semconv.MessagingMessageIDKey.String(strconv.FormatInt(int64(msg.TopicPartition.Offset), 10)), // nolint: gomnd
		semconv.MessagingKafkaMessageKeyKey.String(string(msg.Key)),
		semconv.MessagingKafkaClientIDKey.String(c.Consumer.String()),
		semconv.MessagingKafkaPartitionKey.Int64(int64(msg.TopicPartition.Partition)),
	}
	if c.group != "" {
		attrs = append(attrs, semconv.MessagingKafkaConsumerGroupKey.String(c.group))
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	name := fmt.Sprintf("%s receive", *msg.TopicPartition.Topic)
	ctx, span := c.cfg.Tracer.Start(psc, name, opts...)
	// Inject the current span into the original message so it can be used to
	// propagate the span.
	c.cfg.Propagator.Inject(ctx, carrier)
	return span
}

// Close calls the underlying Consumer.Close and if polling is enabled, ends
// any remaining span.
func (c *Consumer) Close() error {
	err := c.Consumer.Close()
	// Only close the previous span if consuming via the events channel is not
	// enabled. Otherwise, there would be a data race from the consuming
	// goroutine.
	if c.events == nil && c.prev != nil {
		c.prev.End()
		c.prev = nil
	}
	return err
}

// Events returns the kafka Events channel. Message events are traced.
func (c *Consumer) Events() chan kafka.Event {
	return c.events
}

// Poll polls the consumer for events. Message events are traced.
func (c *Consumer) Poll(timeoutMS int) (event kafka.Event) {
	if c.prev != nil {
		c.prev.End()
		c.prev = nil
	}
	evt := c.Consumer.Poll(timeoutMS)
	if msg, ok := evt.(*kafka.Message); ok {
		c.prev = c.startSpan(msg)
	}
	return evt
}
