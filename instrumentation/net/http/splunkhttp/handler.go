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
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	EnvVarServerTimingEnabled = "SPLUNK_CONTEXT_SERVER_TIMING_ENABLED"
)

// NewHandler wraps the passed handler in a span named after the operation and with any provided Options.
// This will also enable all the Splunk specific defaults for HTTP tracing.
func NewHandler(handler http.Handler, operation string, opts ...Option) http.Handler {
	cfg := newConfig(opts...)

	if cfg.ServerTimingEnabled {
		handler = ServerTimingMiddleware(handler)
	}

	handler = otelhttp.NewHandler(handler, "server", cfg.OtelOpts...)
	return handler
}

// WithOtelOpts is use to pass the OpenTelemetry SDK options.
func WithOtelOpts(opts ...otelhttp.Option) Option {
	return OptionFunc(func(cfg *config) {
		cfg.OtelOpts = opts
	})
}

// WithServerTiming enabled or disabled the ServerTimingMiddleware.
func WithServerTiming(v bool) Option {
	return OptionFunc(func(cfg *config) {
		cfg.ServerTimingEnabled = v
	})
}

// Option Interface used for setting *optional* config properties
type Option interface {
	Apply(*config)
}

// config represents the available configuration options.
type config struct {
	OtelOpts            []otelhttp.Option
	ServerTimingEnabled bool
}

// OptionFunc provides a convenience wrapper for simple Options
// that can be represented as functions.
type OptionFunc func(*config)

func (o OptionFunc) Apply(c *config) {
	o(c)
}

// newConfig creates a new config struct and applies opts to it.
func newConfig(opts ...Option) *config {
	serverTimingEnabled := true
	if v := os.Getenv(EnvVarServerTimingEnabled); v == "0" || strings.EqualFold(v, "false") {
		serverTimingEnabled = false
	}

	c := &config{
		ServerTimingEnabled: serverTimingEnabled,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}
