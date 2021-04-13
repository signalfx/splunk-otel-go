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

package splunkhttp_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/oteltest"

	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
)

func ExampleTraceResponseHeaderMiddleware() {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello world") //nolint:errcheck
	})
	handler = splunkhttp.NewHandler(handler, "server", splunkhttp.WithOTelOpts(otelhttp.WithTracerProvider(oteltest.NewTracerProvider())))

	ts := httptest.NewServer(handler)
	defer ts.Close()
	resp, _ := ts.Client().Get(ts.URL) //nolint

	fmt.Println(resp.Header.Get("Access-Control-Expose-Headers"))
	fmt.Println(resp.Header.Get("Server-Timing"))

	// Output:
	// Server-Timing
	// traceparent;desc="00-00000000000000020000000000000000-0000000000000002-01"
}
