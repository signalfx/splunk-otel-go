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

package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestNewWrapperFuncNilRoundTripper(t *testing.T) {
	rt := NewWrapperFunc()(nil)
	require.IsType(t, &roundTripper{}, rt)
	wrapped := rt.(*roundTripper)
	assert.Same(t, http.DefaultTransport, wrapped.RoundTripper)
}

func TestRequestToSpanNameUnrecognized(t *testing.T) {
	path := "/unrecognized"
	r, err := http.NewRequest("GET", path, http.NoBody) //nolint: noctx  // Unused request does not need context.
	require.NoError(t, err)

	expected := "HTTP GET"
	assert.Equalf(t, expected, name(r), "path: %q", path)
}

func TestRequestPathToSpanName(t *testing.T) {
	tests := []struct {
		path string
		name string
	}{
		{
			path: "/api/v1/componentstatuses",
			name: "componentstatuses",
		},
		{
			path: "/api/v1/componentstatuses/NAME",
			name: "componentstatuses/{name}",
		},
		{
			path: "/api/v1/configmaps",
			name: "configmaps",
		},
		{
			path: "/api/v1/namespaces/default/bindings",
			name: "namespaces/{namespace}/bindings",
		},
		{
			path: "/api/v1/namespaces/someothernamespace/configmaps",
			name: "namespaces/{namespace}/configmaps",
		},
		{
			path: "/api/v1/namespaces/default/configmaps/some-config-map",
			name: "namespaces/{namespace}/configmaps/{name}",
		},
		{
			path: "/api/v1/namespaces/default/persistentvolumeclaims/pvc-abcd/status",
			name: "namespaces/{namespace}/persistentvolumeclaims/{name}/status",
		},
		{
			path: "/api/v1/namespaces/default/pods/pod-1234/proxy",
			name: "namespaces/{namespace}/pods/{name}/proxy",
		},
		{
			path: "/api/v1/namespaces/default/pods/pod-5678/proxy/some-path",
			name: "namespaces/{namespace}/pods/{name}/proxy/{path}",
		},
		{
			path: "/api/v1/watch/configmaps",
			name: "watch/configmaps",
		},
		{
			path: "/api/v1/watch/namespaces",
			name: "watch/namespaces",
		},
		{
			path: "/api/v1/watch/namespaces/default/configmaps",
			name: "watch/namespaces/{namespace}/configmaps",
		},
		{
			path: "/api/v1/watch/namespaces/someothernamespace/configmaps/another-name",
			name: "watch/namespaces/{namespace}/configmaps/{name}",
		},
	}

	for _, test := range tests {
		r, err := http.NewRequest("GET", test.path, http.NoBody) //nolint: noctx  // Unused request does not need context.
		require.NoError(t, err)

		expected := "HTTP GET " + test.name
		assert.Equalf(t, expected, name(r), "path: %q", test.path)
	}
}

func TestRequestMethodToSpanName(t *testing.T) {
	tests := []struct {
		method string
		name   string
	}{
		{
			method: http.MethodGet,
			name:   "HTTP GET",
		},
		{
			method: http.MethodHead,
			name:   "HTTP HEAD",
		},
		{
			method: http.MethodPost,
			name:   "HTTP POST",
		},
		{
			method: http.MethodPut,
			name:   "HTTP PUT",
		},
		{
			method: http.MethodPatch,
			name:   "HTTP PATCH",
		},
		{
			method: http.MethodDelete,
			name:   "HTTP DELETE",
		},
		{
			method: http.MethodConnect,
			name:   "HTTP CONNECT",
		},
		{
			method: http.MethodOptions,
			name:   "HTTP OPTIONS",
		},
		{
			method: http.MethodTrace,
			name:   "HTTP TRACE",
		},
	}

	for _, test := range tests {
		r, err := http.NewRequest(test.method, "http://localhost/", http.NoBody) //nolint: noctx  // Unused request does not need context.
		require.NoError(t, err)
		assert.Equalf(t, test.name, name(r), "method: %q", test.method)
	}
}

type readCloser struct {
	readErr, closeErr error
}

var _ io.ReadCloser = readCloser{}

func (rc readCloser) Read(p []byte) (n int, err error) {
	return len(p), rc.readErr
}

func (rc readCloser) Close() error {
	return rc.closeErr
}

type span struct {
	trace.Span

	ended       bool
	recordedErr error

	statusCode codes.Code
	statusDesc string
}

func (s *span) End(...trace.SpanEndOption) {
	s.ended = true
}

func (s *span) RecordError(err error, _ ...trace.EventOption) {
	s.recordedErr = err
}

func (s *span) SetStatus(c codes.Code, d string) {
	s.statusCode, s.statusDesc = c, d
}

func TestWrappedBodyRead(t *testing.T) {
	s := new(span)
	wb := &wrappedBody{span: trace.Span(s), body: readCloser{}}

	msg := []byte("testing response")
	n, err := wb.Read(msg)
	assert.Equal(t, len(msg), n, "Read bytes not forwarded")
	assert.NoError(t, err)

	assert.False(t, s.ended, "span ended without Close")
	assert.NoError(t, s.recordedErr)
	assert.Equal(t, codes.Unset, s.statusCode)
	assert.Equal(t, "", s.statusDesc)
}

func TestWrappedBodyReadEOFError(t *testing.T) {
	s := new(span)
	wb := &wrappedBody{span: trace.Span(s), body: readCloser{readErr: io.EOF}}

	msg := []byte("testing response")
	n, err := wb.Read(msg)
	assert.Equal(t, len(msg), n, "Read bytes not forwarded")
	assert.ErrorIs(t, err, io.EOF)

	assert.True(t, s.ended, "span not ended on read completion")
	assert.NoError(t, s.recordedErr)
	assert.Equal(t, codes.Unset, s.statusCode)
	assert.Equal(t, "", s.statusDesc)
}

func TestWrappedBodyReadError(t *testing.T) {
	s := new(span)
	expectedErr := errors.New("test")
	wb := &wrappedBody{span: trace.Span(s), body: readCloser{readErr: expectedErr}}

	msg := []byte("testing response")
	n, err := wb.Read(msg)
	assert.Equal(t, len(msg), n, "Read bytes not forwarded")
	assert.ErrorIs(t, err, expectedErr)

	assert.False(t, s.ended, "span ended on non-EOF read error")
	assert.ErrorIs(t, s.recordedErr, expectedErr)
	assert.Equal(t, codes.Error, s.statusCode)
	assert.Equal(t, expectedErr.Error(), s.statusDesc)
}

func TestWrappedBodyClose(t *testing.T) {
	s := new(span)
	wb := &wrappedBody{span: trace.Span(s), body: readCloser{}}
	assert.NoError(t, wb.Close())

	assert.True(t, s.ended, "span not ended when Close called")
	assert.NoError(t, s.recordedErr)
	assert.Equal(t, codes.Unset, s.statusCode)
	assert.Equal(t, "", s.statusDesc)
}

func TestWrappedBodyCloseError(t *testing.T) {
	s := new(span)
	expectedErr := errors.New("test")
	wb := &wrappedBody{span: trace.Span(s), body: readCloser{closeErr: expectedErr}}
	assert.ErrorIs(t, wb.Close(), expectedErr)

	assert.True(t, s.ended, "span not ended when Close called")
	assert.NoError(t, s.recordedErr)
	assert.Equal(t, codes.Unset, s.statusCode)
	assert.Equal(t, "", s.statusDesc)
}

type errRoundTripper struct {
	err error
}

var _ http.RoundTripper = (*errRoundTripper)(nil)

func (e *errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestWrappedRoundTripperError(t *testing.T) {
	expected := errors.New("test error")
	tr := NewWrapperFunc()(&errRoundTripper{err: expected})
	c := http.Client{Transport: tr}
	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost", http.NoBody)
	require.NoError(t, err)
	_, err = c.Do(r) //nolint: bodyclose // do not deref a nil response
	require.ErrorIs(t, err, expected)
}
