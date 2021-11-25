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

package internal

import (
	"context"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// Config contains configuration options.
type Config struct {
	// instName is the name of the instrumentation this Config is used for.
	instName string

	Tracer           trace.Tracer
	DefaultStartOpts []trace.SpanStartOption
}

func NewConfig(instrumentationName string, options ...Option) *Config {
	c := Config{instName: instrumentationName}

	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}

	if c.Tracer == nil {
		c.Tracer = otel.Tracer(
			c.instName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	}

	return &c
}

// Copy returns a deep copy of c.
func (c *Config) Copy() *Config {
	newC := Config{
		instName: c.instName,
		Tracer:   c.Tracer,
		// FIXME: is this right?
		DefaultStartOpts: make([]trace.SpanStartOption, 0, len(c.DefaultStartOpts)),
	}

	copy(newC.DefaultStartOpts, c.DefaultStartOpts)

	return &newC
}

// ResolveTracer returns an OpenTelemetry tracer from the appropriate
// TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c *Config) ResolveTracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			c.instName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	}
	return c.Tracer
}

// mergedSpanStartOptions returns a copy of opts with any DefaultStartOpts
// that c is configured with prepended.
func (c *Config) mergedSpanStartOptions(opts ...trace.SpanStartOption) []trace.SpanStartOption {
	if c == nil {
		if len(opts) == 0 {
			return nil
		}
	} else {
		if len(opts)+len(c.DefaultStartOpts) == 0 {
			return nil
		}
	}

	// FIXME: make sure to test capacity is exact.
	merged := make([]trace.SpanStartOption, len(c.DefaultStartOpts)+len(opts))
	if c == nil || len(c.DefaultStartOpts) == 0 {
		copy(merged, opts)
	} else {
		copy(merged, c.DefaultStartOpts)
		copy(merged[len(c.DefaultStartOpts):], opts)
	}
	return merged
}

// WithSpan wraps the function f with a span named name.
func (c *Config) WithSpan(ctx context.Context, name string, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	sso := c.mergedSpanStartOptions(opts...)
	ctx, span := c.ResolveTracer(ctx).Start(ctx, name, sso...)
	err := f(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()

	return err
}
