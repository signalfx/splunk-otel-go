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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

// Option applies options to a configuration.
type Option interface {
	Apply(*Config)
}

// OptionFunc is a generic way to set an option using a func.
type OptionFunc func(*Config)

// Apply applies the configuration option.
func (o OptionFunc) Apply(c *Config) {
	o(c)
}

// WithTracerProvider returns an Option that sets the TracerProvider used for
// a configuration.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return OptionFunc(func(c *Config) {
		c.Tracer = tp.Tracer(
			c.instName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
	})
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created.
func WithAttributes(attr []attribute.KeyValue) Option {
	return OptionFunc(func(c *Config) {
		c.DefaultStartOpts = append(
			c.DefaultStartOpts,
			trace.WithAttributes(attr...),
		)
	})
}

// WithPropagator returns an Option that sets p as the TextMapPropagator used
// when propagating a span context.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return OptionFunc(func(c *Config) {
		c.Propagator = p
	})
}
