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

package splunkleveldb

import (
	"context"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb"

// config contains tracing configuration options.
type config struct {
	ctx              context.Context
	tracer           trace.Tracer
	defaultStartOpts []trace.SpanStartOption
}

func newConfig(options ...Option) *config {
	var c config
	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}

	if c.tracer == nil {
		c.tracer = otel.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}

	if c.ctx == nil {
		c.ctx = context.Background()
	}

	return &c
}

// resolveTracer returns an OTel tracer from the appropriate TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c *config) resolveTracer() trace.Tracer {
	if span := trace.SpanFromContext(c.ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	// There is a possibility that the config was not created with newConfig,
	// try to handle this situation gracefully.
	if c == nil || c.tracer == nil {
		return otel.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	return c.tracer
}

// withSpan wraps the function f with a span.
func (c *config) withSpan(name string, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	var o []trace.SpanStartOption
	if c == nil || len(c.defaultStartOpts) == 0 {
		o = make([]trace.SpanStartOption, len(opts))
		copy(o, opts)
	} else {
		o = make([]trace.SpanStartOption, len(c.defaultStartOpts)+len(opts))
		copy(o, c.defaultStartOpts)
		copy(o[len(c.defaultStartOpts):], opts)
	}

	ctx, span := c.resolveTracer().Start(c.ctx, name, o...)
	err := f(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()

	return err
}

// Option applies options to a configuration.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		c.tracer = tp.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	})
}

// WithContext returns an Option that sets the Context used with this
// instrumentation library by default. This is used to pass context of any
// existing trace to the instrumentation.
func WithContext(ctx context.Context) Option {
	return optionFunc(func(c *config) {
		c.ctx = ctx
	})
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.defaultStartOpts = append(
			c.defaultStartOpts,
			trace.WithAttributes(attr...),
		)
	})
}
