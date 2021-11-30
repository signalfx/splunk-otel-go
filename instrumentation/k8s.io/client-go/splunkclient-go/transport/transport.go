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

// Package transport provides a Kubernetes client wrapper function that traces
// the Kubernetes client operations.
package transport

import (
	"net/http"

	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/internal/config"
	"k8s.io/client-go/transport"
)

func NewWrapperFunc(opts ...config.Option) transport.WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}

		wrapped := roundTripper{
			RoundTripper: rt,
			cfg:          config.NewConfig(opts...),
		}

		return wrapped
	}
}

// roundTripper wraps an http.RoundTripper requests with a span.
type roundTripper struct {
	http.RoundTripper

	cfg *config.Config
}
