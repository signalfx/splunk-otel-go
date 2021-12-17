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

//go:build !(go1.1 || go1.2 || go1.3 || go1.4 || go1.5 || go1.6 || go1.7 || go1.8 || go1.9 || go1.10 || go1.11 || go1.12 || go1.13 || go1.14 || go1.15 || go1.16)
// +build !go1.1,!go1.2,!go1.3,!go1.4,!go1.5,!go1.6,!go1.7,!go1.8,!go1.9,!go1.10,!go1.11,!go1.12,!go1.13,!go1.14,!go1.15,!go1.16

package option

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/internal/config"
)

// Option applies options to a configuration.
type Option interface {
	config.Option
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return config.OptionFunc(func(c *config.Config) {
		c.Tracer = tp.Tracer(
			config.InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	})
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return config.OptionFunc(func(c *config.Config) {
		c.DefaultStartOpts = append(
			c.DefaultStartOpts,
			trace.WithAttributes(attr...),
		)
	})
}

// WithPropagator returns an Option that sets p as the TextMapPropagator used
// when propagating a span context.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return config.OptionFunc(func(c *config.Config) {
		c.Propagator = p
	})
}
