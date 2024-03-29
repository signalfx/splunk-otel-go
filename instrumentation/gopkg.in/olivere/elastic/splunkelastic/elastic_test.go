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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type prop struct {
	propagation.TextMapPropagator

	t      *testing.T
	want   trace.SpanContext
	called bool
}

func (p *prop) Inject(ctx context.Context, _ propagation.TextMapCarrier) {
	got := trace.SpanContextFromContext(ctx)
	assert.True(p.t, p.want.Equal(got), "wrong span context")
	p.called = true
}

func TestPropagation(t *testing.T) {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01},
		SpanID:  trace.SpanID{0x01},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	p := &prop{t: t, want: sc}

	rt := WrapRoundTripper(nil, WithPropagator(p))
	req, err := http.NewRequestWithContext(ctx, "GET", "127.0.0.1", http.NoBody)
	require.NoError(t, err)
	// p.Inject will assert proper span context injection.
	resp, _ := rt.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.True(t, p.called, "did not inject span context")
}

func TestName(t *testing.T) {
	tests := []struct {
		method string
		path   string
		name   string
	}{
		{
			method: "HEAD",
			path:   "", // Test changed to "/" in name func.
			name:   "ping",
		},
		{
			method: "GET",
			path:   "/junk/path",
			name:   "HTTP GET",
		},
		{
			method: "DELETE", // Unknown method for the passed path.
			path:   "/",
			name:   "HTTP DELETE /",
		},
		{
			method: "GET",
			path:   "/_alias",
			name:   "indices.get_alias",
		},
		{
			method: "DELETE",
			path:   "/example_index",
			name:   "indices.delete example_index",
		},
		{
			method: "PUT",
			path:   "/example_index/_bulk",
			name:   "bulk example_index",
		},
	}

	for _, test := range tests {
		url := "http://localhost:9200" + test.path
		req, err := http.NewRequestWithContext(context.Background(), test.method, url, http.NoBody)
		require.NoError(t, err)
		assert.Equal(t, test.name, name(req))
	}
}
