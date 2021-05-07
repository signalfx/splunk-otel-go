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
	"github.com/stretchr/testify/require"
)

type KeyValue struct {
	Key, Value string
}

type OptionTest struct {
	Name          string
	Options       []Option
	AssertionFunc func(*testing.T, *config, error)
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
			return c.AccessToken
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
				AssertionFunc: func(t *testing.T, c *config, e error) {
					assert.NoError(t, e)
					assert.Equal(t, "secret", c.AccessToken)
				},
			},
		},
	},
	{
		Name: "Endpoint",
		ValueFunc: func(c *config) interface{} {
			return c.Endpoint
		},
		DefaultValue: "",
		OptionTests: []OptionTest{
			{
				Name: "valid URL",
				Options: []Option{
					WithEndpoint("https://localhost/"),
				},
				AssertionFunc: func(t *testing.T, c *config, e error) {
					assert.NoError(t, e)
					assert.Equal(t, "https://localhost/", c.Endpoint)
				},
			},
			{
				Name: "invalid URL",
				Options: []Option{
					WithEndpoint("not://a valid.URL"),
				},
				AssertionFunc: func(t *testing.T, c *config, e error) {
					assert.Error(t, e)
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
					c, err := newConfig()
					require.NoError(t, err)
					assert.Equal(t, tc.DefaultValue, tc.ValueFunc(c))
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

			c, err := newConfig()
			if !assert.NoError(t, err) {
				return
			}
			// The expected type is not known, but we can check that the value
			// has changed to verify the environment variable influenced the
			// configuration.
			assert.NotEqual(
				t, tc.DefaultValue, tc.ValueFunc(c),
				"environment variable %s=%q unused", key, val,
			)
		}(ev.Key, ev.Value)
	}
}

func testOptions(t *testing.T, tc *ConfigFieldTest) {
	for _, o := range tc.OptionTests {
		func(t *testing.T, o OptionTest) {
			t.Run(o.Name, func(t *testing.T) {
				c, err := newConfig(o.Options...)
				o.AssertionFunc(t, c, err)
			})
		}(t, o)
	}
}
