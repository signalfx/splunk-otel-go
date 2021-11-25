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
	splunkotel "github.com/signalfx/splunk-otel-go"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// Option applies options to a configuration.
type Option interface {
	apply(*Config)
}

type OptionFunc func(*Config)

func (o OptionFunc) apply(c *Config) {
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
