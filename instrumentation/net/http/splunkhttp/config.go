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

// WithTraceResponseHeader enables or disables the TraceResponseHeaderMiddleware.
//
// The default is to enable the TraceResponseHeaderMiddleware if this option is not passed.
// Additionally, the SPLUNK_TRACE_RESPONSE_HEADER_ENABLED environment variable
// can be set to TRUE or FALSE to specify this option. This option value will be
// given precedence if both it and the environment variable are set.
func WithTraceResponseHeader(v bool) Option {
	return newOptionFunc(func(cfg *config) {
		cfg.TraceResponseHeaderEnabled = v
	})
}

// Option is used for setting optional config properties.
type Option interface {
	otelhttp.Option
	apply(*config)
}

// config represents the available configuration options.
type config struct {
	TraceResponseHeaderEnabled bool
}

func newOptionFunc(fn func(cfg *config)) optionFunc {
	return optionFunc{
		Option: otelhttp.WithNop(),
		fn:     fn,
	}
}

// optionFunc provides a convenience wrapper for simple Options
// that can be represented as functions.
type optionFunc struct {
	otelhttp.Option
	fn func(*config)
}

func (o optionFunc) apply(c *config) {
	o.fn(c)
}

// newConfig creates a new config struct and applies opts to it.
func newConfig(opts ...otelhttp.Option) *config {
	traceResponseHeaderEnabled := true
	if v := os.Getenv(envVarTraceResponseHeaderEnabled); strings.EqualFold(v, "false") {
		traceResponseHeaderEnabled = false
	}

	c := &config{
		TraceResponseHeaderEnabled: traceResponseHeaderEnabled,
	}
	for _, opt := range opts {
		if splunkOpt, ok := opt.(Option); ok {
			splunkOpt.apply(c)
		}
	}
	return c
}
