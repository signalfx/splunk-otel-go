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

// Package splunkhttprouter provides OpenTelemetry instrumentation for the
// github.com/julienschmidt/httprouter module.
//
// Deprecated: the module is not going to be released in future.
// See https://github.com/signalfx/splunk-otel-go/issues/4399 for more details.
package splunkhttprouter

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Router is a traced version of httprouter.Router.
type Router struct {
	*httprouter.Router

	handler http.Handler
}

// New returns a new router augmented with tracing.
func New(opts ...otelhttp.Option) *Router {
	r := &Router{Router: httprouter.New()}

	const nInternalOpts = 1
	o := make([]otelhttp.Option, len(opts)+nInternalOpts)
	// Put this first so it can be overridden by the user.
	o[0] = otelhttp.WithSpanNameFormatter(r.name)
	if len(opts) > 0 {
		copy(o[1:], opts)
	}
	r.handler = otelhttp.NewHandler(r.Router, "", o...)

	return r
}

// ServeHTTP serves req writing a response to w and tracing the process.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

// name is a SpanFormatter for the otelhttp instrumentation.
func (r *Router) name(_ string, req *http.Request) string {
	path := req.URL.Path
	_, params, trailing := r.Lookup(req.Method, path)
	for _, param := range params {
		path = strings.Replace(path, param.Value, ":"+param.Key, 1)
	}
	if trailing {
		path = strings.TrimSuffix(path, "/")
	}

	return "HTTP " + req.Method + " " + path
}
