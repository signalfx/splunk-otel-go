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

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"k8s.io/client-go/transport"
)

func NewWrapperFunc(opts ...otelhttp.Option) transport.WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return wrapRoundTripper(rt, opts...)
	}
}

func wrapRoundTripper(rt http.RoundTripper, opts ...otelhttp.Option) http.RoundTripper {
	defaults := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(nameFormatter),
	}

	return otelhttp.NewTransport(rt, append(defaults, opts...)...)
}

func nameFormatter(operation string, req *http.Request) string {
	return ""
}
