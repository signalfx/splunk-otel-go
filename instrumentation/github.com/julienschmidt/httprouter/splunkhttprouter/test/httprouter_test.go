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
Package test provides end-to-end testing of the splunkhttprouter
instrumentation with the default SDK.

This package is in a separate module from the instrumentation it tests to
isolate the dependency of the default SDK and not impose this as a transitive
dependency for users.
*/
package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	//nolint:staticcheck // Deprecated package, but still used here.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func Error(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusInternalServerError)
}

func Hello(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func newTestServer(tp *trace.TracerProvider) http.Handler {
	router := splunkhttprouter.New(otelhttp.WithTracerProvider(tp))
	router.GET("/error", Error)
	router.GET("/hello/:name", Hello)

	return router
}

func newFixtures(t *testing.T) (*tracetest.SpanRecorder, http.Handler) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(context.Background())) })
	return sr, newTestServer(tp)
}

func TestRouter(t *testing.T) {
	tests := []struct {
		method    string
		target    string
		assertion func(*testing.T, string, string, trace.ReadOnlySpan)
	}{
		{
			method:    http.MethodGet,
			target:    "/hello/bob",
			assertion: assertSpan("/hello/:name", codes.Unset),
		},
		{
			method:    http.MethodGet,
			target:    "/hello/bob/",
			assertion: assertSpan("/hello/:name", codes.Unset),
		},
		{
			method:    http.MethodGet,
			target:    "/error",
			assertion: assertSpan("/error", codes.Error),
		},
	}

	for _, test := range tests {
		sr, r := newFixtures(t)
		req := httptest.NewRequest(test.method, test.target, http.NoBody)
		r.ServeHTTP(httptest.NewRecorder(), req)
		t.Run(test.method+" "+test.target, func(t *testing.T) {
			require.Len(t, sr.Ended(), 1)
			test.assertion(t, test.method, test.target, sr.Ended()[0])
		})
	}
}

func assertSpan(path string, otelCode codes.Code) func(*testing.T, string, string, trace.ReadOnlySpan) {
	return func(t *testing.T, method, _ string, span trace.ReadOnlySpan) {
		name := "HTTP " + method
		if path != "" {
			name += " " + path
		}
		assert.Equal(t, name, span.Name())

		status := span.Status()
		assert.Equal(t, otelCode, status.Code)
		assert.Equal(t, "", status.Description)
	}
}
