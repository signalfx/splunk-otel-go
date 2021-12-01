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

package option

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/internal/config"
)

type fnTracerProvider struct {
	tracer func(string, ...trace.TracerOption) trace.Tracer
}

func (fn *fnTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return fn.tracer(name, opts...)
}

type fnTracer struct {
	start func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

func (fn *fnTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return fn.start(ctx, name, opts...)
}

var mockTracerProvider = &fnTracerProvider{
	tracer: func() func(string, ...trace.TracerOption) trace.Tracer {
		registry := make(map[string]trace.Tracer)
		return func(name string, opts ...trace.TracerOption) trace.Tracer {
			t, ok := registry[name]
			if !ok {
				t = &fnTracer{}
				registry[name] = t
			}
			return t
		}
	}(),
}

func TestWithTracerProvider(t *testing.T) {
	// Default is to use the global TracerProvider. This will override that.
	c := config.NewConfig(WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(config.InstrumentationName)
	assert.Same(t, expected, c.Tracer)
}

func TestConfigTracerFromConfig(t *testing.T) {
	c := config.NewConfig(WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(
		config.InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	got := c.ResolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestWithAttributes(t *testing.T) {
	attr := attribute.String("key", "value")
	c := config.NewConfig(WithAttributes([]attribute.KeyValue{attr}))
	ssc := trace.NewSpanStartConfig(c.DefaultStartOpts...)
	assert.Contains(t, ssc.Attributes(), attr)
}

func TestWithPropagator(t *testing.T) {
	p := propagation.NewCompositeTextMapPropagator()
	// Use a non-nil value.
	p = propagation.NewCompositeTextMapPropagator(p)
	assert.Equal(t, p, config.NewConfig(WithPropagator(p)).Propagator)
}
