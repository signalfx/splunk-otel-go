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
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// Consumer wraps a kafka.Consumer and traces its operations.
type Consumer struct {
	*kafka.Consumer
	cfg       *internal.Config
	waitGroup sync.WaitGroup
	events    chan kafka.Event
	// Use an unsafe.Pointer instead of an atomic.Value to support Go versions
	// prior to 1.17 which do not include a Value.Swap method.
	activeSpan unsafe.Pointer // *consumerSpan
}

// NewConsumer calls kafka.NewConsumer and wraps the resulting Consumer with
// tracing instrumentation.
func NewConsumer(conf *kafka.ConfigMap, opts ...Option) (*Consumer, error) {
	c, err := kafka.NewConsumer(conf)
	if err != nil {
		return nil, err
	}
	cfg := newConfig(opts...)
	// The kafka Consumer does not expose consumer group, but we want to
	// include this in the span attributes.
	cGrpVal, err := conf.Get("group.id", "")
	if err == nil {
		if groupID, ok := cGrpVal.(string); ok {
			cfg.DefaultStartOpts = append(
				cfg.DefaultStartOpts,
				trace.WithAttributes(
					semconv.MessagingKafkaConsumerGroupKey.String(groupID),
				),
			)
		}
	}
	return wrapConsumer(c, cfg), nil
}

// WrapConsumer wraps a kafka.Consumer so that any consumed events are traced.
func WrapConsumer(c *kafka.Consumer, opts ...Option) *Consumer {
	return wrapConsumer(c, newConfig(opts...))
}

// consumerSpan is a wrapper around an OpenTelemetry Span that can be used as
// an atomic value.
type consumerSpan struct {
	otelSpan trace.Span
}

// End completes the wrapped OpenTelemetry span if one exists.
func (s consumerSpan) End(options ...trace.SpanEndOption) {
	if s.otelSpan != nil {
		s.otelSpan.End(options...)
	}
}

func wrapConsumer(c *kafka.Consumer, cfg *internal.Config) *Consumer {
	// Common attributes for all spans this consumer will produce.
	cfg.DefaultStartOpts = append(
		cfg.DefaultStartOpts,
		trace.WithAttributes(
			semconv.MessagingDestinationKindTopic,
			semconv.MessagingOperationReceive,
			semconv.MessagingKafkaClientIDKey.String(c.String()),
		),
	)
	wrapped := &Consumer{
		Consumer: c,
		cfg:      cfg,
		// Set an empty spanHolder to set the activeSpan to empty and ensure that
		// the unsafe.Pointer is set to the correct type.
		activeSpan: unsafe.Pointer(&consumerSpan{}),
	}
	wrapped.events = wrapped.traceEventsChannel(c.Events())
	return wrapped
}

func (c *Consumer) traceEventsChannel(in chan kafka.Event) chan kafka.Event {
	// If the events channel is disabled, in will be nil.
	if in == nil {
		return nil
	}

	out := make(chan kafka.Event, cap(in))
	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()
		defer close(out)
		for evt := range in {
			endTime := time.Now()

			// Only trace messages.
			var s consumerSpan
			if msg, ok := evt.(*kafka.Message); ok {
				s = c.startSpan(msg)
			}

			out <- evt

			active := atomic.SwapPointer(&c.activeSpan, unsafe.Pointer(&s))
			(*consumerSpan)(active).End(trace.WithTimestamp(endTime))
		}
		// finish any remaining span
		(*consumerSpan)(atomic.LoadPointer(&c.activeSpan)).End()
	}()

	return out
}

func (c *Consumer) startSpan(msg *kafka.Message) consumerSpan {
	carrier := NewMessageCarrier(msg)
	psc := c.cfg.Propagator.Extract(context.Background(), carrier)

	const base10 = 10
	offset := strconv.FormatInt(int64(msg.TopicPartition.Offset), base10)
	opts := c.cfg.MergedSpanStartOptions(
		trace.WithAttributes(
			semconv.MessagingDestinationKey.String(*msg.TopicPartition.Topic),
			semconv.MessagingMessageIDKey.String(offset),
			semconv.MessagingKafkaMessageKeyKey.String(string(msg.Key)),
			semconv.MessagingKafkaPartitionKey.Int64(int64(msg.TopicPartition.Partition)),
		),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)

	name := fmt.Sprintf("%s receive", *msg.TopicPartition.Topic)
	ctx, otelSpan := c.cfg.Tracer.Start(psc, name, opts...)
	if err := msg.TopicPartition.Error; err != nil {
		otelSpan.RecordError(err)
		otelSpan.SetStatus(codes.Error, err.Error())
	}

	// Inject the current span into the original message so it can be used to
	// propagate the span.
	c.cfg.Propagator.Inject(ctx, carrier)

	return consumerSpan{otelSpan: otelSpan}
}

// Close calls the underlying Consumer.Close and if polling is enabled, ends
// any remaining span.
func (c *Consumer) Close() error {
	err := c.Consumer.Close()

	// Only close the previous span if consuming via the events channel is not
	// enabled. Otherwise, let that goroutine end the span.
	if c.events == nil {
		// finish any remaining span
		(*consumerSpan)(atomic.LoadPointer(&c.activeSpan)).End()
	}

	// Wait for all spawed goroutines to finish.
	c.waitGroup.Wait()
	return err
}

// Events returns the kafka Events channel. Message events are traced.
func (c *Consumer) Events() chan kafka.Event {
	return c.events
}

// Poll polls the consumer for events. Message events are traced.
//
// Will block for at most timeoutMs milliseconds.
//
// The following callbacks may be triggered:
//   Subscribe()'s rebalanceCb
//
// Returns nil on timeout, else an Event
func (c *Consumer) Poll(timeoutMS int) (event kafka.Event) {
	endTime := time.Now()
	evt := c.Consumer.Poll(timeoutMS)
	if msg, ok := evt.(*kafka.Message); ok {
		s := c.startSpan(msg)
		active := atomic.SwapPointer(&c.activeSpan, unsafe.Pointer(&s))
		(*consumerSpan)(active).End(trace.WithTimestamp(endTime))
	}
	return evt
}

// ReadMessage polls the consumer for a message and traces the read.
//
// This is a convenience API that wraps Poll() and only returns messages or
// errors. All other event types are discarded.
//
// The call will block for at most `timeout` waiting for a new message or
// error. `timeout` may be set to -1 for indefinite wait.
//
// Timeout is returned as (nil, err) where err is `err.(kafka.Error).Code() ==
// kafka.ErrTimedOut`.
//
// Messages are returned as (msg, nil), while general errors are returned as
// (nil, err), and partition-specific errors are returned as (msg, err) where
// msg.TopicPartition provides partition-specific information (such as topic,
// partition and offset).
//
// All other event types, such as PartitionEOF, AssignedPartitions, etc, are
// silently discarded.
func (c *Consumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	endTime := time.Now()
	msg, err := c.Consumer.ReadMessage(timeout)
	if msg != nil {
		s := c.startSpan(msg)
		active := atomic.SwapPointer(&c.activeSpan, unsafe.Pointer(&s))
		(*consumerSpan)(active).End(trace.WithTimestamp(endTime))
	}
	return msg, err
}
