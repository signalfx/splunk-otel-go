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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/oteltest"
)

func TestServerTimingMiddleware(t *testing.T) {
	content := []byte("Any content")
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content) //nolint:errcheck
	})
	handler = ServerTimingMiddleware(handler)
	handler = otelhttp.NewHandler(handler, "server", otelhttp.WithTracerProvider(oteltest.NewTracerProvider()))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("", "/", nil))
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "should return the same status code")
	assert.Equal(t, content, body, "should return the same content")
	assert.Contains(t, resp.Header["Access-Control-Expose-Headers"], "Server-Timing", "should set Access-Control-Expose-Headers header")
	assert.Regexp(t, "^traceparent;desc=\"00-[0-9a-f]{32}-[0-9a-f]{16}-01\"$", resp.Header.Get("Server-Timing"), "should return properly formated Server-Timing header")
}
