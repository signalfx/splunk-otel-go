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

package splunkelastic

import (
	"context"
	"io"
	"net/http"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic"

// WrapRoundTripper returns an http.RoundTripper that wraps the passed rt. All
// requests handled by the returned http.RoundTripper will be traced with the
// assuption they are being made to an Elasticsearch server using the
// gopkg.in/olivere/elastic package.
//
// If rt is nil, the http.DefaultTransport will be used instead.
func WrapRoundTripper(rt http.RoundTripper, opts ...Option) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	cfg := internal.NewConfig(instrumentationName, localToInternal(opts)...)
	cfg.DefaultStartOpts = append([]trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
	}, cfg.DefaultStartOpts...)

	return &roundTripper{RoundTripper: rt, cfg: cfg}

}

// roundTripper wraps an http.RoundTripper's requests with a span.
type roundTripper struct {
	http.RoundTripper

	cfg *internal.Config
}

var _ http.RoundTripper = (*roundTripper)(nil)

func (rt *roundTripper) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	opts := rt.cfg.MergedSpanStartOptions(
		trace.WithAttributes(semconv.HTTPClientAttributesFromHTTPRequest(r)...),
	)

	tracer := rt.cfg.ResolveTracer(r.Context())
	ctx, span := tracer.Start(r.Context(), name(r), opts...)

	// Ensure anything downstream knows about the started span.
	r = r.WithContext(ctx)
	rt.cfg.Propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))

	resp, err = rt.RoundTripper.RoundTrip(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return
	}

	span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)
	span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(resp.StatusCode))
	resp.Body = &wrappedBody{ctx: ctx, span: span, body: resp.Body}

	return
}

// name returns an appropriate span name based on the client request.
// OpenTelemetry semantic conventions require this name to be low cardinality,
// but since the Elasticsearch API is somewhat predictable we can usually
// return more than just "HTTP {METHOD}".
func name(r *http.Request) string {
	path := tokenize(r.URL.Path)
	if path == "" {
		return "HTTP " + r.Method
	}
	return "HTTP " + r.Method + " " + path
}

type wrappedBody struct {
	ctx  context.Context
	span trace.Span
	body io.ReadCloser
}

var _ io.ReadCloser = (*wrappedBody)(nil)

func (wb *wrappedBody) Read(b []byte) (int, error) {
	n, err := wb.body.Read(b)
	switch err {
	case nil:
		// nothing to do here but fall through to the return
	case io.EOF:
		wb.span.End()
	default:
		wb.span.RecordError(err)
		wb.span.SetStatus(codes.Error, err.Error())
	}

	return n, err
}

func (wb *wrappedBody) Close() error {
	wb.span.End()
	return wb.body.Close()
}
