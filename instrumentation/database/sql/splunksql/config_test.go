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

package splunksql

import (
	"context"
	"errors"
	"net/url"
	"testing"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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

func TestConfigDefaultTracerProvider(t *testing.T) {
	c := newTraceConfig()
	assert.Equal(t, otel.GetTracerProvider(), c.TracerProvider)
}

func TestWithTracerProvider(t *testing.T) {
	// Default is to use the global TracerProvider. This will override that.
	tp := new(fnTracerProvider)
	c := newTraceConfig(WithTracerProvider(tp))
	assert.Same(t, tp, c.TracerProvider)
}

func TestConfigTracerFromGlobal(t *testing.T) {
	c := newTraceConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	got := c.tracer(context.Background())
	assert.Equal(t, expected, got)
}

func TestConfigTracerFromConfig(t *testing.T) {
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{}
		},
	}
	c := newTraceConfig(WithTracerProvider(tp))
	expected := tp.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	got := c.tracer(context.Background())
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
	c := newTraceConfig()
	got := c.tracer(ctx)
	expected := trace.NewNoopTracerProvider().Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
	assert.Equal(t, expected, got)
}

func TestURLDNSParse(t *testing.T) {
	testcases := []struct {
		name        string
		dsn         string
		expectedCfg ConnectionConfig
		errStr      string
	}{
		{
			name: "not a URL",
			dsn:  `:¯\_(ツ)_/¯:`,
			errStr: (&url.Error{
				Op:  "parse",
				URL: `:¯\_(ツ)_/¯:`,
				Err: errors.New("missing protocol scheme"),
			}).Error(),
		},
		{
			name: "params",
			dsn:  "param0=val0,paramN=valN",
			expectedCfg: ConnectionConfig{
				ConnectionString: "param0=val0,paramN=valN",
			},
		},
		{
			name: "host only",
			dsn:  "http://localhost",
			expectedCfg: ConnectionConfig{
				ConnectionString: "http://localhost",
				Host:             "localhost",
			},
		},
		{
			name: "host:port",
			dsn:  "https://localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://localhost:8080",
				Host:             "localhost",
				Port:             8080,
			},
		},
		{
			name: "with user",
			dsn:  "https://bob@localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://bob@localhost:8080",
				User:             "bob",
				Host:             "localhost",
				Port:             8080,
			},
		},
		{
			name: "redact password",
			dsn:  "https://bob:pa55w0rd@localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://bob@localhost:8080",
				User:             "bob",
				Host:             "localhost",
				Port:             8080,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			connCfg, err := urlDSNParse(tc.dsn)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCfg, connCfg)
		})
	}
}

func TestSpanName(t *testing.T) {
	c := newTraceConfig()

	// c.DBName empty means the moniker should be used.
	m := moniker.Begin
	assert.Equal(t, m.String(), c.spanName(m))

	const dbname = "test database"
	c.DBName = dbname
	assert.Equal(t, dbname, c.spanName(m))
}
