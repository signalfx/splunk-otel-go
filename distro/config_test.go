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

	testr "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
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
			return c.ExportConfig.accessToken
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
		t.Run(ev.Key, func(t *testing.T) {
			t.Setenv(ev.Key, ev.Value)

			// The expected type is not known, but we can check that the value
			// has changed to verify the environment variable influenced the
			// configuration.
			assert.NotEqual(
				t, tc.DefaultValue, tc.ValueFunc(newTestConfig(t)),
				"environment variable %s=%q unused", ev.Key, ev.Value,
			)
		})
	}
}
