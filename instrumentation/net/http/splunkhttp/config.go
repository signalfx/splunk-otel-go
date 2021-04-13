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
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Environmental variables used for configuration.
const (
	envVarTraceResponseHeaderEnabled = "SPLUNK_TRACE_RESPONSE_HEADER_ENABLED" // Adds `Server-Timing` header to HTTP responses
)

// WithOTelOpts is use to pass the OpenTelemetry SDK options.
func WithOTelOpts(opts ...otelhttp.Option) Option {
	return optionFunc(func(cfg *config) {
		cfg.OTelOpts = opts
	})
}

// WithTraceResponseHeader enables or disables the TraceResponseHeaderMiddleware.
//
// The default is to enable the TraceResponseHeaderMiddleware if this option is not passed.
// Additionally, the SPLUNK_TRACE_RESPONSE_HEADER_ENABLED environment variable
// can be set to TRUE or FALSE to specify this option. This option value will be
// given precedence if both it and the environment variable are set.
func WithTraceResponseHeader(v bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.TraceResponseHeaderEnabled = v
	})
}

// Option is used for setting optional config properties.
type Option interface {
	apply(*config)
}

// config represents the available configuration options.
type config struct {
	OTelOpts                   []otelhttp.Option
	TraceResponseHeaderEnabled bool
}

// optionFunc provides a convenience wrapper for simple Options
// that can be represented as functions.
type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// newConfig creates a new config struct and applies opts to it.
func newConfig(opts ...Option) *config {
	traceResponseHeaderEnabled := true
	if v := os.Getenv(envVarTraceResponseHeaderEnabled); strings.EqualFold(v, "false") {
		traceResponseHeaderEnabled = false
	}

	c := &config{
		TraceResponseHeaderEnabled: traceResponseHeaderEnabled,
	}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}
