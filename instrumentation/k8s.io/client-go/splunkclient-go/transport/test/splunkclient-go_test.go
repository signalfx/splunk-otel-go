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

package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/option"
	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport"
)

func TestEndToEndWrappedTransport(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prop := propagation.TraceContext{}
	content := []byte("Hello, world!")

	ctx, parent := tp.Tracer("TestEndToEndWrappedTransport").Start(context.Background(), "parent")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanContextFromContext(
			prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header)),
		)
		assert.Equal(t, parent.SpanContext().TraceID(), span.TraceID())
		_, err := w.Write(content)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)

	tr := transport.NewWrapperFunc(
		option.WithPropagator(prop),
		option.WithTracerProvider(tp),
	)(http.DefaultTransport)

	c := http.Client{Transport: tr}
	resp, err := c.Do(r)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, content, body)
	require.NoError(t, resp.Body.Close())

	require.NoError(t, tp.Shutdown(context.Background()))
	require.Len(t, sr.Ended(), 1)

	span := sr.Ended()[0]
	assert.Equal(t, "HTTP GET", span.Name())
	assert.Equal(t, trace.SpanKindClient, span.SpanKind())
	assert.Contains(t, span.Attributes(), attribute.String("http.method", "GET"))
	assert.Contains(t, span.Attributes(), attribute.String("http.url", ts.URL))
	assert.Contains(t, span.Attributes(), attribute.String("http.scheme", "http"))
	assert.Contains(t, span.Attributes(), attribute.String("http.host", strings.TrimPrefix(ts.URL, "http://")))
	assert.Contains(t, span.Attributes(), attribute.String("http.flavor", "1.1"))
	assert.Contains(t, span.Attributes(), attribute.Int("http.status_code", 200))
	assert.Equal(t, codes.Unset, span.Status().Code)
	assert.Equal(t, "", span.Status().Description)
}
