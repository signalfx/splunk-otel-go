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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/oteltest"
)

func TestServerTimingMiddleware(t *testing.T) {
	resp := responseForHandler(func(handler http.Handler) http.Handler { // nolint:bodyclose // Body is not used
		handler = ServerTimingMiddleware(handler)
		return otelhttp.NewHandler(handler, "server", otelhttp.WithTracerProvider(oteltest.NewTracerProvider()))
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode, "should return OK status code")
	assert.Contains(t, resp.Header["Access-Control-Expose-Headers"], "Server-Timing", "should set Access-Control-Expose-Headers header")
	assert.Regexp(t, "^traceparent;desc=\"00-[0-9a-f]{32}-[0-9a-f]{16}-01\"$", resp.Header.Get("Server-Timing"), "should return properly formated Server-Timing header")
}

func TestNewHandlerDefault(t *testing.T) {
	resp := responseForHandler(func(handler http.Handler) http.Handler { // nolint:bodyclose // Body is not used
		return NewHandler(handler, "server", WithOTelOpts(otelhttp.WithTracerProvider(oteltest.NewTracerProvider())))
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode, "should return OK status code")
	assert.Contains(t, resp.Header["Access-Control-Expose-Headers"], "Server-Timing", "should set Access-Control-Expose-Headers header")
	assert.Regexp(t, "^traceparent;desc=\"00-[0-9a-f]{32}-[0-9a-f]{16}-01\"$", resp.Header.Get("Server-Timing"), "should return properly formated Server-Timing header")
}

func TestNewHandlerServerTimingDisabled(t *testing.T) {
	resp := responseForHandler(func(handler http.Handler) http.Handler { // nolint:bodyclose // Body is not used
		return NewHandler(handler, "server", WithOTelOpts(otelhttp.WithTracerProvider(oteltest.NewTracerProvider())), WithServerTiming(false))
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode, "should return OK status code")
	assert.NotContains(t, resp.Header["Access-Control-Expose-Headers"], "Server-Timing", "should NOT set Access-Control-Expose-Headers header")
	assert.NotRegexp(t, "^traceparent;desc=\"00-[0-9a-f]{32}-[0-9a-f]{16}-01\"$", resp.Header.Get("Server-Timing"), "should not add traceID to Server-Timing header")
}

func responseForHandler(wrapFn func(http.Handler) http.Handler) *http.Response {
	content := []byte("Any content")
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content) //nolint:errcheck
	})
	handler = wrapFn(handler)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("", "/", nil))
	return w.Result()
}
