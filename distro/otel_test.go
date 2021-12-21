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

package distro_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/goleak"

	"github.com/signalfx/splunk-otel-go/distro"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRunJaegerExporter(t *testing.T) {
	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req *http.Request)
	}{
		{
			desc: "WithEndpoint",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				return distro.Run(distro.WithEndpoint(url))
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"), "should send thrift formatted trace")
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				clearFn := distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Cleanup(clearFn)
				return distro.Run()
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"), "should send thrift formatted trace")
			},
		},
		{
			desc: "WithEndpoint and WithAccessToken",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				return distro.Run(distro.WithEndpoint(url), distro.WithAccessToken("my-token"))
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"), "should send thrift formatted trace")
				user, pass, ok := got.BasicAuth()
				if !ok {
					assert.Fail(t, "should have Basic Authentication headers")
					return
				}
				assert.Equal(t, "auth", user, "should have proper username")
				assert.Equal(t, "my-token", pass, "should use the provided token as passowrd")
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				clearFn := distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Cleanup(clearFn)
				clearFn = distro.Setenv("SPLUNK_ACCESS_TOKEN", "my-token")
				t.Cleanup(clearFn)
				return distro.Run()
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"), "should send thrift formatted trace")
				user, pass, ok := got.BasicAuth()
				if !ok {
					assert.Fail(t, "should have Basic Authentication headers")
					return
				}
				assert.Equal(t, "auth", user, "should have proper username")
				assert.Equal(t, "my-token", pass, "should use the provided token as passowrd")
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// HTTP endpoint where a trace is sent
			reqCh := make(chan *http.Request, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				reqCh <- r
			}))
			defer srv.Close()

			// setup tracer
			t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
			sdk, err := tc.setupFn(t, srv.URL)
			require.NoError(t, err, "should configure tracing")

			// create a sample span
			_, span := otel.Tracer("distro/otel_test").Start(ctx, "TestRun")
			span.SetAttributes(attribute.Key("ex.com/foo").String("bar"))
			span.AddEvent("working")
			span.End()

			// shutdown tracer - this should send the trace
			err = sdk.Shutdown(ctx)
			require.NoError(t, err, "should finish tracing")

			// assert that the span has been received
			select {
			case <-ctx.Done():
				require.Fail(t, "test timeout out")
			case got := <-reqCh:
				tc.assertFn(t, got)
			}
		})
	}
}
