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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"

func newConfig(options ...Option) *internal.Config {
	o := append([]internal.Option{internal.OptionFunc(func(c *internal.Config) {
		c.DefaultStartOpts = []trace.SpanStartOption{
			trace.WithAttributes(semconv.MessagingSystemKey.String("kafka")),
		}
	})}, localToInternal(options)...)
	return internal.NewConfig(instrumentationName, o...)
}

func localToInternal(opts []Option) []internal.Option {
	out := make([]internal.Option, len(opts))
	for i, o := range opts {
		out[i] = internal.Option(o)
	}
	return out
}

// Option applies options to a configuration.
type Option interface {
	internal.Option
}

// WithTracerProvider returns an Option that sets the TracerProvider used for
// a configuration.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return Option(internal.WithTracerProvider(tp))
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created.
func WithAttributes(attr []attribute.KeyValue) Option {
	return Option(internal.WithAttributes(attr))
}

// WithPropagator returns an Option that sets p as the TextMapPropagator used
// when propagating a span context.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return Option(internal.WithPropagator(p))
}
