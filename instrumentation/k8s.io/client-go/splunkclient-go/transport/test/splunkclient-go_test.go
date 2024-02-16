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

//go:build go1.17
// +build go1.17

package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	splunkclientgo "github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/option"
	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport"
)

func request(t *testing.T, handle func(http.ResponseWriter, *http.Request)) (*tracetest.SpanRecorder, *http.Response, string) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prop := propagation.TraceContext{}

	ctx, parent := tp.Tracer("request").Start(context.Background(), "parent")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanContextFromContext(
			prop.Extract(r.Context(), propagation.HeaderCarrier(r.Header)),
		)
		assert.Equal(t, parent.SpanContext().TraceID(), span.TraceID())

		handle(w, r)
	}))
	t.Cleanup(ts.Close)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)

	tr := transport.NewWrapperFunc(
		option.WithPropagator(prop),
		option.WithTracerProvider(tp),
	)(http.DefaultTransport)

	c := http.Client{Transport: tr}
	resp, err := c.Do(r)
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	return sr, resp, ts.URL
}

func TestEndToEndWrappedTransport(t *testing.T) {
	content := []byte("Hello, world!")
	sr, resp, url := request(t, func(w http.ResponseWriter, r *http.Request) {
		n, err := w.Write(content)
		assert.NoError(t, err)
		assert.Equal(t, len(content), n)
	})

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, content, body)
	require.NoError(t, resp.Body.Close())

	require.Len(t, sr.Ended(), 1)
	span := sr.Ended()[0]
	assert.Equal(t, "HTTP GET", span.Name())
	assert.Equal(t, trace.SpanKindClient, span.SpanKind())
	assert.Equal(t, splunkclientgo.Version(), span.InstrumentationLibrary().Version)
	assert.Contains(t, span.Attributes(), semconv.HTTPMethodKey.String("GET"))
	assert.Contains(t, span.Attributes(), semconv.HTTPURLKey.String(url))
	assert.Contains(t, span.Attributes(), semconv.HTTPFlavorHTTP11)
	assert.Contains(t, span.Attributes(), semconv.HTTPStatusCodeKey.Int(200))
	assert.Equal(t, codes.Unset, span.Status().Code)
	assert.Equal(t, "", span.Status().Description)
}

func TestWrappedTransportErrorResponse(t *testing.T) {
	sr, resp, _ := request(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	require.NoError(t, resp.Body.Close())

	require.Len(t, sr.Ended(), 1)
	span := sr.Ended()[0]
	assert.Contains(t, span.Attributes(), semconv.HTTPStatusCodeKey.Int(400))
	assert.Equal(t, codes.Error, span.Status().Code)
	assert.Equal(t, "", span.Status().Description)
}
