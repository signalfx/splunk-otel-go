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
	"testing"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const iName = "github.com/signalfx/splunk-otel-go/instrumentation/internal"

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
	c := NewConfig(iName)
	expected := otel.Tracer(
		iName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, c.Tracer)
}

func TestWithTracer(t *testing.T) {
	// Default is to use the global TracerProvider. This will override that.
	c := NewConfig(iName, WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(iName)
	assert.Same(t, expected, c.Tracer)
}

func TestResolveTracerFromGlobal(t *testing.T) {
	c := NewConfig(iName)
	expected := otel.Tracer(
		iName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	got := c.ResolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestConfigTracerFromConfig(t *testing.T) {
	c := NewConfig(iName, WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(
		iName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	got := c.ResolveTracer(context.Background())
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
	c := NewConfig(iName)
	got := c.ResolveTracer(ctx)
	expected := trace.NewNoopTracerProvider().Tracer(
		iName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, got)
}

func TestMergedSpanStartOptionsNilConifg(t *testing.T) {
	c := (*Config)(nil)
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsEmptyConfigNoPassedOpts(t *testing.T) {
	c := NewConfig(iName)
	c.DefaultStartOpts = nil
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsPassedNoOptsWithDefaults(t *testing.T) {
	c := Config{
		DefaultStartOpts: []trace.SpanStartOption{trace.WithAttributes()},
	}
	sso := c.mergedSpanStartOptions()
	assert.Len(t, sso, 1)
	assert.Equal(t, 1, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedNoOptsNoDefaults(t *testing.T) {
	c := Config{DefaultStartOpts: nil}
	sso := c.mergedSpanStartOptions()
	assert.Len(t, sso, 0)
	assert.Equal(t, 0, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedOptsWithDefaults(t *testing.T) {
	c := Config{
		DefaultStartOpts: []trace.SpanStartOption{trace.WithAttributes()},
	}
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 2)
	assert.Equal(t, 2, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedOptsNoDefaults(t *testing.T) {
	c := Config{DefaultStartOpts: nil}
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
	assert.Equal(t, 1, cap(sso), "incorrectly sized slice")
}

func TestConfigDefaultPropagator(t *testing.T) {
	c := NewConfig(iName)
	expected := otel.GetTextMapPropagator()
	assert.Equal(t, expected, c.Propagator)
}
