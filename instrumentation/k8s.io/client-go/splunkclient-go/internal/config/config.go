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

// Package config provides configuration options.
package config

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

// InstrumentationName is the instrumentation library identifier for a Tracer.
const InstrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go"

// Config contains tracing configuration options.
type Config struct {
	Tracer           trace.Tracer
	Propagators      propagation.TextMapPropagator
	DefaultStartOpts []trace.SpanStartOption
}

// NewConfig returns a Config with all options applied and defaults set.
func NewConfig(options ...Option) *Config {
	c := Config{}

	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}

	if c.Tracer == nil {
		c.Tracer = otel.Tracer(
			InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}

	return &c
}

// ResolveTracer returns an OpenTelemetry Tracer from the appropriate
// TracerProvider.
//
// If the passed context contains a Span, the TracerProvider that created the
// Tracer that created that Span will be returned. Otherwise, c.Tracer is
// returned.
func (c *Config) ResolveTracer(ctx context.Context) trace.Tracer {
	// There is a possibility that the config was not created with newConfig,
	// try to handle this situation gracefully.
	if c == nil {
		return otel.Tracer(
			InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}

	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}

	if c.Tracer == nil {
		return otel.Tracer(
			InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	return c.Tracer
}

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

	var merged []trace.SpanStartOption
	if c == nil || len(c.DefaultStartOpts) == 0 {
		merged = make([]trace.SpanStartOption, len(opts))
		copy(merged, opts)
	} else {
		merged = make([]trace.SpanStartOption, len(c.DefaultStartOpts)+len(opts))
		copy(merged, c.DefaultStartOpts)
		copy(merged[len(c.DefaultStartOpts):], opts)
	}
	return merged
}

// Option applies options to a configuration.
type Option interface {
	apply(*Config)
}

// OptionFunc applies a functional setting to a configuration.
type OptionFunc func(*Config)

func (o OptionFunc) apply(c *Config) {
	o(c)
}
