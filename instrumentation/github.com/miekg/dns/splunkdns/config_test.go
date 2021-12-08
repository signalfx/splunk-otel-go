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

package splunkdns

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
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

func TestConfigDefaultTracer(t *testing.T) {
	c := newConfig()
	expect := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expect, c.tracer)
}

func TestWithTracer(t *testing.T) {
	tracer := &fnTracer{}
	// Default is to use the global TracerProvider. This will override that.
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return tracer
		},
	}
	c := newConfig(WithTracerProvider(tp))
	assert.Same(t, tracer, c.tracer)
}

func TestEmptyConfigTracer(t *testing.T) {
	// If a config is directly created, fallback to the OTel global.
	c := config{}
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	got := c.resolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestConfigTracerFromGlobal(t *testing.T) {
	c := newConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	got := c.resolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestConfigTracerFromConfig(t *testing.T) {
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{}
		},
	}
	c := newConfig(WithTracerProvider(tp))
	expected := tp.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	got := c.resolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestConfigTracerFromContext(t *testing.T) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	// This context will contain a non-recording span.
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	// Use the global TracerProvider in the config and override with the
	// passed context to the tracer method.
	c := newConfig()
	got := c.resolveTracer(ctx)
	expected := trace.NewNoopTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, got)
}

func TestWithAttributes(t *testing.T) {
	attr := []attribute.KeyValue{
		attribute.String("key", "value"),
	}
	c := newConfig(WithAttributes(attr))
	assert.Len(t, c.defaultStartOpts, 1)
	sc := trace.NewSpanStartConfig(c.defaultStartOpts...)
	assert.Equal(t, attr, sc.Attributes())
}
