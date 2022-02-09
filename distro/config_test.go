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

package distro

import (
	"testing"

	"github.com/go-logr/logr"
	testr "github.com/go-logr/logr/testing"
	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type KeyValue struct {
	Key, Value string
}

type OptionTest struct {
	Name          string
	Options       []Option
	AssertionFunc func(*testing.T, *config)
}

type ConfigFieldTest struct {
	Name             string
	ValueFunc        func(*config) interface{}
	DefaultValue     interface{}
	EnvironmentTests []KeyValue
	OptionTests      []OptionTest
}

var ConfigurationTests = []*ConfigFieldTest{
	{
		Name: "AccessToken",
		ValueFunc: func(c *config) interface{} {
			return c.ExportConfig.AccessToken
		},
		DefaultValue: "",
		EnvironmentTests: []KeyValue{
			{Key: accessTokenKey, Value: "secret"},
		},
		OptionTests: []OptionTest{
			{
				Name: "valid name",
				Options: []Option{
					WithAccessToken("secret"),
				},
				AssertionFunc: func(t *testing.T, c *config) {
					assert.Equal(t, "secret", c.ExportConfig.AccessToken)
				},
			},
		},
	},
	{
		Name: "Endpoint",
		ValueFunc: func(c *config) interface{} {
			return c.ExportConfig.Endpoint
		},
		DefaultValue: "",
		OptionTests: []OptionTest{
			{
				Name: "valid URL",
				Options: []Option{
					WithEndpoint("https://localhost/"),
				},
				AssertionFunc: func(t *testing.T, c *config) {
					assert.Equal(t, "https://localhost/", c.ExportConfig.Endpoint)
				},
			},
		},
	},
	{
		Name: "Propagator",
		ValueFunc: func(c *config) interface{} {
			return c.Propagator
		},
		DefaultValue: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
		EnvironmentTests: []KeyValue{
			{Key: otelPropagatorsKey, Value: "tracecontext"},
		},
		OptionTests: []OptionTest{
			{
				Name: "nil propagator",
				Options: []Option{
					WithPropagator(nil),
				},
				AssertionFunc: func(t *testing.T, c *config) {
					assert.Equal(t, nonePropagator, c.Propagator)
				},
			},
			{
				Name: "set to tracecontext",
				Options: []Option{
					WithPropagator(propagation.TraceContext{}),
				},
				AssertionFunc: func(t *testing.T, c *config) {
					assert.Equal(t, propagation.TraceContext{}, c.Propagator)
				},
			},
		},
	},
}

func newTestConfig(t *testing.T, opts ...Option) *config {
	l := testr.NewTestLogger(t)
	return newConfig(append(opts, WithLogger(l))...)
}

func TestConfig(t *testing.T) {
	for _, tc := range ConfigurationTests {
		func(t *testing.T, tc *ConfigFieldTest) {
			t.Run(tc.Name, func(t *testing.T) {
				t.Run("DefaultValue", func(t *testing.T) {
					assert.Equal(t, tc.DefaultValue, tc.ValueFunc(newTestConfig(t)))
				})

				t.Run("EnvironmentVariableOverride", func(t *testing.T) {
					testEnvironmentOverrides(t, tc)
				})

				t.Run("OptionTests", func(t *testing.T) {
					testOptions(t, tc)
				})
			})
		}(t, tc)
	}
}

func testEnvironmentOverrides(t *testing.T, tc *ConfigFieldTest) {
	for _, ev := range tc.EnvironmentTests {
		func(key, val string) {
			revert := Setenv(key, val)
			defer revert()

			// The expected type is not known, but we can check that the value
			// has changed to verify the environment variable influenced the
			// configuration.
			assert.NotEqual(
				t, tc.DefaultValue, tc.ValueFunc(newTestConfig(t)),
				"environment variable %s=%q unused", key, val,
			)
		}(ev.Key, ev.Value)
	}
}

func testOptions(t *testing.T, tc *ConfigFieldTest) {
	for _, o := range tc.OptionTests {
		func(t *testing.T, o OptionTest) {
			t.Run(o.Name, func(t *testing.T) {
				o.AssertionFunc(t, newTestConfig(t, o.Options...))
			})
		}(t, o)
	}
}

func TestSetPropagatorComposite(t *testing.T) {
	c := config{Logger: logr.Discard()}
	c.setPropagator("tracecontext,baggage,b3,b3multi,jaeger,xray,ottrace,garbage")
	assert.Equal(t, propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader)),
		b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
		jaeger.Jaeger{},
		xray.Propagator{},
		ot.OT{},
	), c.Propagator)
}

func TestSetPropagatorDefault(t *testing.T) {
	c := config{Logger: logr.Discard()}
	c.setPropagator("")
	assert.Equal(t, propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	), c.Propagator)
}

func TestSetPropagatorCompositeWithNone(t *testing.T) {
	// Assumes specification as stated:
	//
	//   "none": No automatically configured propagator.
	//
	// means if "none" is included in the value, no propagator should be
	// configured. Therefore, setPropagator needs to return just the
	// nonePropagator value to signal this behavior.
	c := config{Logger: logr.Discard()}
	c.setPropagator("tracecontext,baggage,none")
	assert.Equal(t, nonePropagator, c.Propagator)
}

type stringEncoder struct {
	zapcore.PrimitiveArrayEncoder
	got string
}

func (e *stringEncoder) AppendString(s string) {
	e.got = s
}

func (e *stringEncoder) assert(t *testing.T, want string) {
	assert.Equal(t, want, e.got, "encoded wrong level")
}

func TestZapLevelEncoder(t *testing.T) {
	levelMap := map[int8]string{
		// Not that we use it, but 5 is the "fatal" level in zap.
		5:  "error",
		4:  "error",
		3:  "error",
		2:  "error",
		1:  "error",
		0:  "warn",
		-1: "info",
	}

	enc := new(stringEncoder)
	for level, want := range levelMap {
		zapLevelEncoder(zapcore.Level(level), enc)
		enc.assert(t, want)
	}

	// Debug should be for all verbosity between -2 and the end of the of the
	// zap level range (-127).
	for i := -2; i >= -127; i-- {
		zapLevelEncoder(zapcore.Level(i), enc)
		enc.assert(t, "debug")
	}
}

func TestZapLevel(t *testing.T) {
	testcases := []struct {
		in   string
		want int8
	}{
		{in: "debug", want: -127},
		{in: "info", want: -1},
		{in: "warn", want: 0},
		{in: "error", want: 1},
		{in: "invalid", want: -1},
	}

	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, zapcore.Level(tc.want), zapLevel(tc.in))
		})
	}
}

func TestLoggerFallback(t *testing.T) {
	orig := fallbackLoggerFunc
	t.Cleanup(func() { fallbackLoggerFunc = orig })

	var called bool
	fallbackLoggerFunc = func() logr.Logger {
		called = true
		return logr.Discard()
	}

	zc := zapConfig()
	// Set an invalid level so the zap logger build will error and a fallback
	// logger is returned.
	zc.Level = zap.AtomicLevel{}
	_ = logger(zc)
	assert.True(t, called, "fallback logger not created")
}

func TestFallbackLoggerFunc(t *testing.T) {
	sink := fallbackLoggerFunc().GetSink()
	assert.Implementsf(t, (*stdr.Underlier)(nil), sink, "invalid logr.LogSink type %T", sink)
}
