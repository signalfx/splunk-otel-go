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

package splunkhttp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigs(t *testing.T) {
	tests := []struct {
		name   string
		opts   []Option
		envs   map[string]string
		assert func(t *testing.T, c *config)
	}{
		// Defaiult
		{
			name: "Default",
			assert: func(t *testing.T, c *config) {
				assert.Nil(t, c.OTelOpts, "should not set any additional OTel options")
				assert.True(t, c.ServerTimingEnabled, "should enable ServerTiming")
			},
		},
		// ServerTiming
		{
			name: "ServerTiming WithServerTiming(false)",
			opts: []Option{
				WithServerTiming(false),
			},
			assert: func(t *testing.T, c *config) {
				assert.False(t, c.ServerTimingEnabled, "should disable ServerTiming")
			},
		},
		{
			name: "ServerTiming SPLUNK_CONTEXT_SERVER_TIMING_ENABLED=False",
			envs: map[string]string{
				"SPLUNK_CONTEXT_SERVER_TIMING_ENABLED": "False",
			},
			assert: func(t *testing.T, c *config) {
				assert.False(t, c.ServerTimingEnabled, "should disable ServerTiming")
			},
		},
		{
			name: "ServerTiming WithServerTiming(true) SPLUNK_CONTEXT_SERVER_TIMING_ENABLED=True",
			envs: map[string]string{
				"SPLUNK_CONTEXT_SERVER_TIMING_ENABLED": "False",
			},
			opts: []Option{
				WithServerTiming(true),
			},
			assert: func(t *testing.T, c *config) {
				assert.True(t, c.ServerTimingEnabled, "should enable ServerTiming, because option has higher priority than env var")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// set env var before the test and bring the old values back after the test
			for key, val := range tt.envs {
				if v, ok := os.LookupEnv(key); ok {
					defer func() { os.Setenv(key, v) }()
				} else {
					defer func() { os.Unsetenv(key) }()
				}
				os.Setenv(key, val)
			}

			cfg := newConfig(tt.opts...)
			tt.assert(t, cfg)
		})
	}
}
