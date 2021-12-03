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

// Package splunkchi provides OpenTelemetry instrumentation for the
// github.com/go-chi/chi package.
package splunkchi

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Middleware returns github.com/go-chi/chi middleware that traces served
// requests.
func Middleware(options ...Option) func(http.Handler) http.Handler {
	cfg := newConfig(options...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allows us to track the ultimate status.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			tracer := cfg.resolveTracer(r.Context())
			carrier := propagation.HeaderCarrier(r.Header)
			ctx := cfg.propagator.Extract(r.Context(), carrier)
			// The full handler chain needs to be complete before we are sure
			// what path is being requested. Delay full naming and annotation
			// until then.
			name := "HTTP " + r.Method
			ctx, span := tracer.Start(ctx, name, cfg.defaultStartOpts...)
			defer span.End()
			r = r.WithContext(ctx)

			next.ServeHTTP(ww, r)

			path := chi.RouteContext(r.Context()).RoutePattern()
			attrs := semconv.HTTPServerAttributesFromHTTPRequest("", path, r)
			attrs = append(attrs, semconv.HTTPAttributesFromHTTPStatusCode(ww.Status())...)
			attrs = append(attrs, semconv.NetAttributesFromHTTPRequest("tcp", r)...)
			attrs = append(attrs, semconv.EndUserAttributesFromHTTPRequest(r)...)
			span.SetAttributes(attrs...)

			if path != "" {
				span.SetName(name + " " + path)
			}

			code, desc := semconv.SpanStatusFromHTTPStatusCode(ww.Status())
			span.SetStatus(code, desc)
		})
	}
}
