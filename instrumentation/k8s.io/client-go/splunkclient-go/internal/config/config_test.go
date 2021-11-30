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
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

func TestConfigDefaultTracer(t *testing.T) {
	c := NewConfig()
	expected := otel.Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.Tracer)
}

func TestResolveTracerFromGlobal(t *testing.T) {
	c := NewConfig()
	expected := otel.Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	got := c.ResolveTracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestNilConfigResolvedTracer(t *testing.T) {
	c := (*Config)(nil)
	expected := otel.Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.ResolveTracer(context.Background()))
}

func TestEmptyConfigResolvedTracer(t *testing.T) {
	c := &Config{}
	expected := otel.Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, c.ResolveTracer(context.Background()))
}

func TestConfigTracerFromContext(t *testing.T) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	// This context will contain a non-recording span.
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	got := NewConfig().ResolveTracer(ctx)
	expected := trace.NewNoopTracerProvider().Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, got)
}

func TestDefaultSpanStartOptions(t *testing.T) {
	c := NewConfig()
	assert.Len(t, c.DefaultStartOpts, 0)
}

func TestMergedSpanStartOptionsNilConifg(t *testing.T) {
	c := (*Config)(nil)
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsNilConifgPassedOpts(t *testing.T) {
	c := (*Config)(nil)
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
}

func TestMergedSpanStartOptionsEmptyConfigNoPassedOpts(t *testing.T) {
	c := NewConfig()
	c.DefaultStartOpts = nil
	assert.Nil(t, c.mergedSpanStartOptions())
}

func TestMergedSpanStartOptionsPassedNoOpts(t *testing.T) {
	c := NewConfig()
	sso := c.mergedSpanStartOptions()
	assert.Len(t, sso, 0)
}

func TestMergedSpanStartOptionsPassedOpts(t *testing.T) {
	c := NewConfig()
	sso := c.mergedSpanStartOptions(trace.WithAttributes())
	assert.Len(t, sso, 1)
}
