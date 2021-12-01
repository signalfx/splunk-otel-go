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

// Package option provides configuration options for the splunkclient-go
// package and its subpackages.
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
		c.Propagators = p
	})
}
