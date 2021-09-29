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
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// A Producer wraps a kafka.Producer and traces its operations.
type Producer struct {
	*kafka.Producer
	cfg            config
	produceChannel chan *kafka.Message
}

// NewProducer calls kafka.NewProducer and wraps the resulting Producer with
// tracing instrumentation.
func NewProducer(conf *kafka.ConfigMap, opts ...Option) (*Producer, error) {
	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, err
	}
	return WrapProducer(p, opts...), nil
}

// WrapProducer wraps a kafka.Producer so that any produced events are traced.
func WrapProducer(p *kafka.Producer, opts ...Option) *Producer {
	wrapped := &Producer{
		Producer: p,
		cfg:      newConfig(opts...),
	}
	wrapped.produceChannel = wrapped.traceProduceChannel(p.ProduceChannel())
	return wrapped
}

func (p *Producer) traceProduceChannel(out chan *kafka.Message) chan *kafka.Message {
	if out == nil {
		return nil
	}

	in := make(chan *kafka.Message, cap(out))
	go func() {
		for msg := range in {
			span := p.startSpan(msg)
			out <- msg
			span.End()
		}
	}()

	return in
}

func (p *Producer) startSpan(msg *kafka.Message) trace.Span {
	carrier := NewMessageCarrier(msg)
	psc := p.cfg.Propagator.Extract(context.Background(), carrier)

	offset := strconv.FormatInt(int64(msg.TopicPartition.Offset), 10)
	attrs := append(p.cfg.Attributes,
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationKey.String(*msg.TopicPartition.Topic),
		semconv.MessagingMessageIDKey.String(offset),
		semconv.MessagingKafkaMessageKeyKey.String(string(msg.Key)),
		semconv.MessagingKafkaPartitionKey.Int64(int64(msg.TopicPartition.Partition)),
	)
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}

	name := fmt.Sprintf("%s send", *msg.TopicPartition.Topic)
	ctx, span := p.cfg.Tracer.Start(psc, name, opts...)
	// Inject the current span into the original message so it can be used to
	// propagate the span.
	p.cfg.Propagator.Inject(ctx, carrier)
	return span
}

// Close calls the wrapped Producer.Close and closes the producer channel.
func (p *Producer) Close() {
	close(p.produceChannel)
	p.Producer.Close()
}

// Produce calls the wrapped Producer.Produce and traces the request.
func (p *Producer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	span := p.startSpan(msg)

	// if the user has selected a delivery channel, we will wrap it and
	// wait for the delivery event to finish the span
	if deliveryChan != nil {
		oldDeliveryChan := deliveryChan
		deliveryChan = make(chan kafka.Event)
		go func() {
			var err error
			evt := <-deliveryChan
			if respMsg, ok := evt.(*kafka.Message); ok {
				// delivery errors are returned via TopicPartition.Error
				err = respMsg.TopicPartition.Error
			}
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			span.End()
			oldDeliveryChan <- evt
		}()
	}

	err := p.Producer.Produce(msg, deliveryChan)
	// with no delivery channel, finish immediately
	if deliveryChan == nil {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}

	return err
}

// ProduceChannel returns the traced producer channel.
func (p *Producer) ProduceChannel() chan *kafka.Message {
	return p.produceChannel
}
