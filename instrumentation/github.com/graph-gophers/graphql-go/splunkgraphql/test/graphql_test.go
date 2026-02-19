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

/*
Package test provides end-to-end testing of the splunkgraphql instrumentation
with the default SDK.

This package is in a separate module from the instrumentation it tests to
isolate the dependency of the default SDK and not impose this as a transitive
dependency for users.
*/
package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql/internal"
)

const testSchema = `
	schema {
		query: Query
	}
	type Query {
		hello: String!
		helloNonTrivial: String!
	}
`

const helloWorld = "Hello, world!"

type testResolver struct{}

func (*testResolver) Hello() string                    { return helloWorld }
func (*testResolver) HelloNonTrivial() (string, error) { return helloWorld, nil }

func fixtures(t *testing.T) (*tracetest.SpanRecorder, *httptest.Server) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	t.Cleanup(func() { assert.NoError(t, tp.Shutdown(context.Background())) })

	tracer := graphql.Tracer(splunkgraphql.NewTracer(splunkgraphql.WithTracerProvider(tp)))
	schema := graphql.MustParseSchema(testSchema, new(testResolver), tracer)
	srv := httptest.NewServer(&relay.Handler{Schema: schema})
	t.Cleanup(srv.Close)

	return sr, srv
}

func TestTracerNonTrivial(t *testing.T) {
	sr, srv := fixtures(t)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, strings.NewReader(`{
		"query": "query TestQuery() { hello, helloNonTrivial }",
		"operationName": "TestQuery"
	}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, resp.Body.Close()) })
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"data":{"hello":"Hello, world!","helloNonTrivial":"Hello, world!"}}`, string(body))

	spans := sr.Ended()
	require.Len(t, spans, 3)
	assert.Equal(t, spans[2].SpanContext().TraceID(), spans[1].SpanContext().TraceID())

	s := spans[0]
	assert.Equal(t, splunkgraphql.Version(), s.InstrumentationScope().Version)
	assert.Equal(t, "GraphQL validation", s.Name())
	assert.Equal(t, traceapi.SpanKindInternal, s.SpanKind())

	s = spans[1]
	assert.Equal(t, "GraphQL field", s.Name())
	assert.Equal(t, traceapi.SpanKindServer, s.SpanKind())
	assert.Contains(t, s.Attributes(), internal.GraphQLFieldKey.String("helloNonTrivial"))
	assert.Contains(t, s.Attributes(), internal.GraphQLTypeKey.String("Query"))

	s = spans[2]
	assert.Equal(t, "GraphQL request", s.Name())
	assert.Equal(t, traceapi.SpanKindServer, s.SpanKind())
	assert.Contains(t, s.Attributes(), internal.GraphQLQueryKey.String(
		"query TestQuery() { hello, helloNonTrivial }",
	))
}

func TestTracerTrivial(t *testing.T) {
	sr, srv := fixtures(t)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, strings.NewReader(`{
		"query": "{ hello }"
	}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, resp.Body.Close()) })
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"data":{"hello":"Hello, world!"}}`, string(body))

	spans := sr.Ended()
	require.Len(t, spans, 2)

	// Trivial queries should not trace a field access.
	s0, s1 := spans[0], spans[1]
	assert.Equal(t, "GraphQL validation", s0.Name())
	assert.Equal(t, "GraphQL request", s1.Name())
	assert.Contains(t, s1.Attributes(), internal.GraphQLQueryKey.String("{ hello }"))
}
