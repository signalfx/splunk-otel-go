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
	"encoding/hex"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// TraceResponseHeaderMiddleware wraps the passed handler, functioning like middleware.
// It adds trace context in traceparent form (https://www.w3.org/TR/trace-context/#traceparent-header)
// as Server-Timing header (https://www.w3.org/TR/server-timing/)
// to the HTTP response.
func TraceResponseHeaderMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
			w.Header().Add("Access-Control-Expose-Headers", "Server-Timing")

			traceID := spanCtx.TraceID()
			hexTraceID := hex.EncodeToString(traceID[:])
			spanID := spanCtx.SpanID()
			hexSpanID := hex.EncodeToString(spanID[:])
			traceParent := "traceparent;desc=\"00-" + hexTraceID + "-" + hexSpanID + "-01\""
			w.Header().Add("Server-Timing", traceParent)
		}

		handler.ServeHTTP(w, r)
	})
}
