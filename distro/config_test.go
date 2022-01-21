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

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
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

func TestConfig(t *testing.T) {
	for _, tc := range ConfigurationTests {
		func(t *testing.T, tc *ConfigFieldTest) {
			t.Run(tc.Name, func(t *testing.T) {
				t.Run("DefaultValue", func(t *testing.T) {
					assert.Equal(t, tc.DefaultValue, tc.ValueFunc(newConfig()))
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
				t, tc.DefaultValue, tc.ValueFunc(newConfig()),
				"environment variable %s=%q unused", key, val,
			)
		}(ev.Key, ev.Value)
	}
}

func testOptions(t *testing.T, tc *ConfigFieldTest) {
	for _, o := range tc.OptionTests {
		func(t *testing.T, o OptionTest) {
			t.Run(o.Name, func(t *testing.T) {
				o.AssertionFunc(t, newConfig(o.Options...))
			})
		}(t, o)
	}
}

func TestLoadPropagatorComposite(t *testing.T) {
	assert.Equal(t, propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader)),
		b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
		jaeger.Jaeger{},
		xray.Propagator{},
		ot.OT{},
	), loadPropagator("tracecontext,baggage,b3,b3multi,jaeger,xray,ottrace,garbage"))
}

func TestLoadPropagatorDefault(t *testing.T) {
	assert.Equal(t, propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	), loadPropagator(""))
}

func TestLoadPropagatorCompositeWithNone(t *testing.T) {
	// Assumes specification as stated:
	//
	//   "none": No automatically configured propagator.
	//
	// means if "none" is included in the value, no propagator should be
	// configured. Therefore, loadPropagator needs to return just the
	// nonePropagator value to signal this behavior.
	assert.Equal(t, nonePropagator, loadPropagator("tracecontext,baggage,none"))
}
