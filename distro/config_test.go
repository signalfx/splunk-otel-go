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
)

type keyValue struct {
	Key, Value string
}

type configFieldTest struct {
	Name             string
	ValueFunc        func(*config) interface{}
	DefaultValue     interface{}
	EnvironmentTests []keyValue
}

var configurationTests = []*configFieldTest{
	{
		Name: "AccessToken",
		ValueFunc: func(c *config) interface{} {
			return c.ExportConfig.AccessToken
		},
		DefaultValue: "",
		EnvironmentTests: []keyValue{
			{Key: accessTokenKey, Value: "secret"},
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
		EnvironmentTests: []keyValue{
			{Key: otelPropagatorsKey, Value: "tracecontext"},
		},
	},
}

func newTestConfig(t *testing.T, opts ...Option) *config {
	l := testr.NewTestLogger(t)
	return newConfig(append(opts, WithLogger(l))...)
}

func TestConfig(t *testing.T) {
	for _, tc := range configurationTests {
		func(t *testing.T, tc *configFieldTest) {
			t.Run(tc.Name, func(t *testing.T) {
				t.Run("DefaultValue", func(t *testing.T) {
					assert.Equal(t, tc.DefaultValue, tc.ValueFunc(newTestConfig(t)))
				})

				t.Run("EnvironmentVariableOverride", func(t *testing.T) {
					testEnvironmentOverrides(t, tc)
				})
			})
		}(t, tc)
	}
}

func testEnvironmentOverrides(t *testing.T, tc *configFieldTest) {
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
