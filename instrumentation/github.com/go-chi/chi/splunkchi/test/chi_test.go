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

/*
Package test provides end-to-end testing of the splunkchi instrumentation with
the default SDK.

This package is in a separate module from the instrumentation it tests to
isolate the dependency of the default SDK and not impose this as a transitive
dependency for users.
*/
package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi"
)

func newTestServer(tp *trace.TracerProvider) *chi.Mux {
	r := chi.NewRouter()
	r.Use(splunkchi.Middleware(splunkchi.WithTracerProvider(tp)))
	r.Route("/users", func(r chi.Router) {
		r.Route("/{user}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Put("/", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			})
		})
	})

	r.HandleFunc("/error", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	return r
}

func newFixtures(t *testing.T) (*tracetest.SpanRecorder, *chi.Mux) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })
	return sr, newTestServer(tp)
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		method    string
		target    string
		assertion func(*testing.T, string, string, trace.ReadOnlySpan)
	}{
		{
			method:    http.MethodGet,
			target:    "/users/bob",
			assertion: assertSpan("/users/{user}/", codes.Unset, http.StatusOK),
		},
		{
			method:    http.MethodPut,
			target:    "/users/bob",
			assertion: assertSpan("/users/{user}/", codes.Unset, http.StatusAccepted),
		},
		{
			method:    http.MethodGet,
			target:    "/error",
			assertion: assertSpan("/error", codes.Error, http.StatusInternalServerError),
		},
	}

	for _, test := range tests {
		sr, r := newFixtures(t)
		req := httptest.NewRequest(test.method, test.target, http.NoBody)
		r.ServeHTTP(httptest.NewRecorder(), req)
		require.Len(t, sr.Ended(), 1)
		test.assertion(t, test.method, test.target, sr.Ended()[0])
	}
}

func assertSpan(path string, otelCode codes.Code, httpCode int) func(*testing.T, string, string, trace.ReadOnlySpan) {
	return func(t *testing.T, method, _ string, span trace.ReadOnlySpan) {
		name := "HTTP " + method
		if path != "" {
			name += " " + path
		}
		assert.Equal(t, name, span.Name())
		assert.Equal(t, traceapi.SpanKindServer, span.SpanKind())
		assert.Equal(t, splunkchi.Version(), span.InstrumentationScope().Version)

		status := span.Status()
		assert.Equal(t, otelCode, status.Code)
		assert.Equal(t, "", status.Description)

		attrs := span.Attributes()
		assert.Contains(t, attrs, semconv.HTTPMethodKey.String(method))
		assert.Contains(t, attrs, semconv.HTTPRouteKey.String(path))
		assert.Contains(t, attrs, semconv.HTTPSchemeHTTP)
		assert.Contains(t, attrs, semconv.HTTPStatusCodeKey.Int(httpCode))
		assert.Contains(t, attrs, semconv.HTTPFlavorHTTP11)

		keys := make(map[attribute.Key]struct{}, len(attrs))
		for _, a := range attrs {
			keys[a.Key] = struct{}{}
		}

		// These key values are potentially dynamic. Test an attribute
		// with this key is set regardless of its value.
		wantKeys := []attribute.Key{
			semconv.NetHostNameKey,
			semconv.NetSockPeerAddrKey,
			semconv.NetSockPeerPortKey,
		}
		for _, k := range wantKeys {
			assert.Contains(t, keys, k)
		}
	}
}
