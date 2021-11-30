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

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

func TestConfigDefaultTracer(t *testing.T) {
	c := newConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.tracer)
}

func TestWithTracer(t *testing.T) {
	// Default is to use the global TracerProvider. This will override that.
	c := newConfig(WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(instrumentationName)
	assert.Same(t, expected, c.tracer)
}

func TestResolveTracerFromGlobal(t *testing.T) {
	c := newConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	got := c.resolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestNilConfigResolvedTracer(t *testing.T) {
	c := (*config)(nil)
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.resolveTracer(context.Background()))
}

func TestEmptyConfigResolvedTracer(t *testing.T) {
	c := &config{}
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.resolveTracer(context.Background()))
}

func TestConfigTracerFromConfig(t *testing.T) {
	c := newConfig(WithTracerProvider(mockTracerProvider))
	expected := mockTracerProvider.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
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
	got := newConfig().resolveTracer(ctx)
	expected := trace.NewNoopTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, got)
}

func TestDefaultSpanStartOptions(t *testing.T) {
	c := newConfig()
	assert.Len(t, c.defaultStartOpts, 0)
}

func TestWithAttributes(t *testing.T) {
	attr := attribute.String("key", "value")
	c := newConfig(WithAttributes([]attribute.KeyValue{attr}))
	ssc := trace.NewSpanStartConfig(c.defaultStartOpts...)
	assert.Contains(t, ssc.Attributes(), attr)
}

func TestMergedSpanStartOptionsNilConifg(t *testing.T) {
	c := (*config)(nil)
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsNilConifgPassedOpts(t *testing.T) {
	c := (*config)(nil)
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
}

func TestMergedSpanStartOptionsEmptyConfigNoPassedOpts(t *testing.T) {
	c := newConfig()
	c.defaultStartOpts = nil
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsPassedNoOpts(t *testing.T) {
	c := newConfig()
	sso := c.mergedSpanStartOptions()
	assert.Len(t, sso, 0)
}

func TestMergedSpanStartOptionsPassedOpts(t *testing.T) {
	c := newConfig()
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
}
