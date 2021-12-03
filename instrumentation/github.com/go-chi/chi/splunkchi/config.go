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

package splunkchi

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi"

// config contains configuration options.
type config struct {
	tracer           trace.Tracer
	propagator       propagation.TextMapPropagator
	defaultStartOpts []trace.SpanStartOption
}

// newConfig returns a Config for instrumentation with all options applied.
//
// If no TracerProvider or Propagator are specified with options, the default
// OpenTelemetry globals will be used.
func newConfig(options ...Option) *config {
	c := &config{
		defaultStartOpts: []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
		},
	}

	for _, o := range options {
		if o != nil {
			o.apply(c)
		}
	}

	if c.tracer == nil {
		c.tracer = otel.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	}

	if c.propagator == nil {
		c.propagator = otel.GetTextMapPropagator()
	}

	return c
}

// resolveTracer returns an OpenTelemetry tracer from the appropriate
// TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c *config) resolveTracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	}
	return c.tracer
}

// Option applies options to a configuration.
type Option interface {
	apply(*config)
}

// optionFunc is a generic way to set an option using a func.
type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithTracerProvider returns an Option that sets the TracerProvider used for
// a configuration.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		c.tracer = tp.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	})
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.defaultStartOpts = append(
			c.defaultStartOpts,
			trace.WithAttributes(attr...),
		)
	})
}

// WithPropagator returns an Option that sets p as the TextMapPropagator used
// when propagating a span context.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		c.propagator = p
	})
}
