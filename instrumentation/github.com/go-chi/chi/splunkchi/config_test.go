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

package splunkchi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

var mockTracerProvider = func(spanRecorder map[string]*mockSpan) trace.TracerProvider {
	recordSpan := func(s *mockSpan) {
		if spanRecorder != nil {
			spanRecorder[s.Name] = s
		}
	}

	return &fnTracerProvider{
		tracer: func() func(string, ...trace.TracerOption) trace.Tracer {
			registry := make(map[string]trace.Tracer)
			return func(name string, opts ...trace.TracerOption) trace.Tracer {
				t, ok := registry[name]
				if !ok {
					t = &fnTracer{
						start: func(ctx context.Context, n string, o ...trace.SpanStartOption) (context.Context, trace.Span) {
							span := &mockSpan{Name: n, StartOpts: o}
							recordSpan(span)
							ctx = trace.ContextWithSpan(ctx, span)
							return ctx, span
						},
					}
					registry[name] = t
				}
				return t
			}
		}(),
	}
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

type status struct {
	Code        codes.Code
	Description string
}

type mockSpan struct {
	trace.Span

	Name      string
	StartOpts []trace.SpanStartOption

	RecordedErrs []error
	Statuses     []status
	Ended        bool
}

func (s *mockSpan) RecordError(err error, _ ...trace.EventOption) {
	s.RecordedErrs = append(s.RecordedErrs, err)
}

func (s *mockSpan) SetStatus(c codes.Code, desc string) {
	s.Statuses = append(s.Statuses, status{Code: c, Description: desc})
}

func (s *mockSpan) End(...trace.SpanEndOption) {
	s.Ended = true
}

func TestConfigDefaultTracer(t *testing.T) {
	c := newConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, c.tracer)
}

func TestResolveTracerFromGlobal(t *testing.T) {
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
	mtp := mockTracerProvider(nil)
	c := newConfig(WithTracerProvider(mtp))
	expected := mtp.Tracer(
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

func TestConfigDefaultPropagator(t *testing.T) {
	c := newConfig()
	expected := otel.GetTextMapPropagator()
	assert.Equal(t, expected, c.propagator)
}

func TestWithTracerProvider(t *testing.T) {
	mtp := mockTracerProvider(nil)
	// Default is to use the global TracerProvider. This will override that.
	c := newConfig(WithTracerProvider(mtp))
	expected := mtp.Tracer(instrumentationName)
	assert.Same(t, expected, c.tracer)
}

func TestWithAttributes(t *testing.T) {
	attr := attribute.String("key", "value")
	c := newConfig(WithAttributes([]attribute.KeyValue{attr}))
	ssc := trace.NewSpanStartConfig(c.defaultStartOpts...)
	assert.Contains(t, ssc.Attributes(), attr)
}

func TestWithPropagator(t *testing.T) {
	p := propagation.NewCompositeTextMapPropagator()
	// Use a non-nil value.
	p = propagation.NewCompositeTextMapPropagator(p)
	assert.Equal(t, p, newConfig(WithPropagator(p)).propagator)
}
