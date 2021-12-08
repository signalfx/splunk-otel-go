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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

const iName = "github.com/signalfx/splunk-otel-go/instrumentation/internal"

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
	c := NewConfig(iName)
	expected := otel.Tracer(
		iName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, c.Tracer)
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
	mtp := mockTracerProvider(nil)
	c := NewConfig(iName, WithTracerProvider(mtp))
	expected := mtp.Tracer(
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
	assert.Nil(t, c.MergedSpanStartOptions())
}

func TestMergedSpanStartOptionsNilConifgPassedOpts(t *testing.T) {
	c := (*Config)(nil)
	sso := c.MergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
	assert.Equal(t, 1, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsEmptyConfigNoPassedOpts(t *testing.T) {
	c := NewConfig(iName)
	c.DefaultStartOpts = nil
	assert.Nil(t, c.MergedSpanStartOptions())
}

func TestMergedSpanStartOptionsPassedNoOptsWithDefaults(t *testing.T) {
	c := Config{
		DefaultStartOpts: []trace.SpanStartOption{trace.WithAttributes()},
	}
	sso := c.MergedSpanStartOptions()
	assert.Len(t, sso, 1)
	assert.Equal(t, 1, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedNoOptsNoDefaults(t *testing.T) {
	c := Config{DefaultStartOpts: nil}
	sso := c.MergedSpanStartOptions()
	assert.Len(t, sso, 0)
	assert.Equal(t, 0, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedOptsWithDefaults(t *testing.T) {
	c := Config{
		DefaultStartOpts: []trace.SpanStartOption{trace.WithAttributes()},
	}
	sso := c.MergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 2)
	assert.Equal(t, 2, cap(sso), "incorrectly sized slice")
}

func TestMergedSpanStartOptionsPassedOptsNoDefaults(t *testing.T) {
	c := Config{DefaultStartOpts: nil}
	sso := c.MergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
	assert.Equal(t, 1, cap(sso), "incorrectly sized slice")
}

func TestConfigDefaultPropagator(t *testing.T) {
	c := NewConfig(iName)
	expected := otel.GetTextMapPropagator()
	assert.Equal(t, expected, c.Propagator)
}

func TestCopy(t *testing.T) {
	prop := propagation.NewCompositeTextMapPropagator()
	// Use a non-nil propagator.
	prop = propagation.NewCompositeTextMapPropagator(prop)

	c := NewConfig(
		iName,
		WithTracerProvider(mockTracerProvider(nil)),
		WithPropagator(prop),
		WithAttributes([]attribute.KeyValue{}),
	)
	cp := c.Copy()

	assert.Equal(t, c, cp)

	// cp should be completely independent of c.

	c.instName = "different"
	assert.NotEqual(t, iName, c.instName)
	assert.Equal(t, iName, cp.instName)

	origTracer := cp.Tracer
	c.Tracer = otel.Tracer("different")
	assert.NotEqual(t, origTracer, c.Tracer)
	assert.Equal(t, origTracer, cp.Tracer)

	origProp := cp.Propagator
	c.Propagator = otel.GetTextMapPropagator()
	assert.NotEqual(t, origProp, c.Propagator)
	assert.Equal(t, origProp, cp.Propagator)

	origOpt := cp.DefaultStartOpts[0]
	// Changing the underlying array data does not change the copied data.
	c.DefaultStartOpts[0] = trace.WithNewRoot()
	assert.NotEqual(t, []trace.SpanStartOption{origOpt}, c.DefaultStartOpts)
	assert.Equal(t, []trace.SpanStartOption{origOpt}, cp.DefaultStartOpts)
}

func TestWithSpan(t *testing.T) {
	const spanName = "TestWithSpan span"
	opts := []trace.SpanStartOption{
		trace.WithAttributes(),
		trace.WithAttributes(attribute.Bool("set", true)),
	}
	spanRecorder := make(map[string]*mockSpan)
	c := NewConfig(iName, WithTracerProvider(mockTracerProvider(spanRecorder)))

	expectedErr := errors.New("TestWithSpan error")
	var called bool
	err := c.WithSpan(context.Background(), spanName, func(c context.Context) error {
		called = true
		return expectedErr
	}, opts...)
	assert.ErrorIs(t, err, expectedErr)
	assert.True(t, called, "WithSpan did not call passed func")

	require.Contains(t, spanRecorder, spanName)
	span := spanRecorder[spanName]

	assert.Equal(t, spanName, span.Name)
	assert.Equal(t, opts, span.StartOpts)

	require.Len(t, span.RecordedErrs, 1)
	assert.ErrorIs(t, span.RecordedErrs[0], expectedErr)

	require.Len(t, span.Statuses, 1)
	assert.Equal(t, span.Statuses[0], status{Code: codes.Error, Description: expectedErr.Error()})

	assert.True(t, span.Ended, "mockSpan not ended by WithSpan")
}
