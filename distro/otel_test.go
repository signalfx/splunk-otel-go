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

	"github.com/signalfx/splunk-otel-go/distro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// TestRun is a smoke test that ensures that traces are sent using thrift protocol.
func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// HTTP endpoint where a trace is sent
	reqCh := make(chan *http.Request, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		reqCh <- r
	}))
	defer srv.Close()

	// setup tracer
	sdk, err := distro.Run(distro.WithEndpoint(srv.URL))
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
		assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"), "should send thrift formatted trace")
	}
}
