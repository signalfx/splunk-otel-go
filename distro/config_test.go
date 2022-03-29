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
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
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
	{
		Name: "WithSpanLimits",
		ValueFunc: func(c *config) interface{} {
			return c.SpanLimits
		},
		DefaultValue: &trace.SpanLimits{
			AttributeValueLengthLimit:   12000,
			AttributeCountLimit:         -1,
			EventCountLimit:             -1,
			LinkCountLimit:              1000,
			AttributePerEventCountLimit: -1,
			AttributePerLinkCountLimit:  -1,
		},
		EnvironmentTests: []KeyValue{
			{Key: attributeValueLengthKey, Value: "10"},
			{Key: spanAttributeValueLengthKey, Value: "10"},
			{Key: attributeCountKey, Value: "10"},
			{Key: spanAttributeCountKey, Value: "10"},
			{Key: spanEventCountKey, Value: "10"},
			{Key: spanEventAttributeCountKey, Value: "10"},
			{Key: spanLinkCountKey, Value: "10"},
			{Key: spanLinkAttributeCountKey, Value: "10"},
		},
		OptionTests: []OptionTest{
			{
				Name: "valid override",
				Options: []Option{
					WithSpanLimits(trace.SpanLimits{
						AttributeValueLengthLimit:   100,
						AttributeCountLimit:         100,
						EventCountLimit:             100,
						LinkCountLimit:              100,
						AttributePerEventCountLimit: 100,
						AttributePerLinkCountLimit:  100,
					}),
				},
				AssertionFunc: func(t *testing.T, c *config) {
					assert.Equal(t, expectedSL(100, 100, 100, 100, 100, 100), c.SpanLimits)
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
