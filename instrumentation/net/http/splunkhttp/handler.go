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

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHandler wraps the passed handler in a span named after the operation and with any provided Options.
// This will also enable all the Splunk specific defaults for HTTP tracing.
func NewHandler(handler http.Handler, operation string, opts ...otelhttp.Option) http.Handler {
	// add additional Splunk specific instrumentations
	cfg := newConfig(opts...)
	if cfg.TraceResponseHeaderEnabled {
		handler = TraceResponseHeaderMiddleware(handler)
	}

	// make sure only valid otelhttp.Option are passed to otelhttp.NewHandler
	var otelOpts []otelhttp.Option
	for _, opt := range opts {
		if _, ok := opt.(optionWrapper); ok {
			continue
		}
		otelOpts = append(otelOpts, opt)
	}
	handler = otelhttp.NewHandler(handler, operation, otelOpts...)
	return handler
}
