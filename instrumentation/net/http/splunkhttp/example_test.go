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
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
)

//nolint:errcheck // example usage
func Example() {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "Hello world")
	})
	handler = splunkhttp.NewHandler(handler)
	handler = otelhttp.NewHandler(handler, "server", otelhttp.WithTracerProvider(trace.NewTracerProvider()))

	ts := httptest.NewServer(handler)
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Access-Control-Expose-Headers:", resp.Header.Get("Access-Control-Expose-Headers"))
	fmt.Println("Server-Timing:", resp.Header.Get("Server-Timing"))
}
