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
	"fmt"
	"net/url"
	"os"
)

// Environment variable keys that set values of the configuration.
const (
	accessTokenKey = "SPLUNK_ACCESS_TOKEN"
)

// config is the configuration used to create and operate an SDK.
type config struct {
	AccessToken string
	Endpoint    string
}

// newConfig returns a validated config with Splunk defaults.
func newConfig(opts ...Option) (*config, error) {
	c := &config{
		AccessToken: envOr(accessTokenKey, ""),
	}

	for _, o := range opts {
		o.apply(c)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// Validate ensures c is valid, otherwise returning an appropriate error.
func (c config) Validate() error {
	var errs []string

	if c.Endpoint != "" {
		if _, err := url.Parse(c.Endpoint); err != nil {
			errs = append(errs, "invalid endpoint: %s", err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

// envOr returns the environment variable value associated with key if it
// exists, otherwise it returns alt.
func envOr(key, alt string) string {
	v, ok := os.LookupEnv(key)
	if ok {
		return v
	}
	return alt
}

// Option sets a config setting value.
type Option interface {
	apply(*config)
}

// optionFunc is a functional option implementation for Option interface.
type optionFunc func(*config)

func (fn optionFunc) apply(c *config) {
	fn(c)
}

// WithEndpoint configures the endpoint telemetry is sent to.
// Setting empty string results in no operation.
func WithEndpoint(endpoint string) Option {
	return optionFunc(func(c *config) {
		c.Endpoint = endpoint
	})
}

// WithAccessToken configures the auth token
// allowing exporters to send data directly to a Splunk back-end.
// Setting empty string results in no operation.
func WithAccessToken(accessToken string) Option {
	return optionFunc(func(c *config) {
		c.AccessToken = accessToken
	})
}
