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

//go:build !windows

package splunkkafka

import (
	splunkotel "github.com/signalfx/splunk-otel-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"

// config contains tracing configuration options.
type config struct {
	Tracer     trace.Tracer
	Propagator propagation.TextMapPropagator
}

func newConfig(options ...Option) config {
	var c config
	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}
	if c.Tracer == nil {
		c.Tracer = otel.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	if c.Propagator == nil {
		c.Propagator = otel.GetTextMapPropagator()
	}
	return c
}

// Option applies options to a tracing configuration.
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
		c.Tracer = tp.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	})
}

// WithPropagator specifies the TextMapPropagator to use when extracting and
// injecting cross-cutting concerns. If none is specified, the global
// TextMapPropagator will be used.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *config) {
		cfg.Propagator = propagator
	})
}
