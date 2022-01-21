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
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	ctpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tpb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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

type exportRequest struct {
	Header metadata.MD
	Spans  []*tpb.Span
}

type collectorTraceServiceServer struct {
	ctpb.UnimplementedTraceServiceServer

	requests chan exportRequest
}

var _ ctpb.TraceServiceServer = (*collectorTraceServiceServer)(nil)

func (ctss *collectorTraceServiceServer) Export(ctx context.Context, exp *ctpb.ExportTraceServiceRequest) (*ctpb.ExportTraceServiceResponse, error) {
	rs := exp.ResourceSpans[0]
	ils := rs.GetInstrumentationLibrarySpans()[0]
	headers, _ := metadata.FromIncomingContext(ctx)

	ctss.requests <- exportRequest{Header: headers, Spans: ils.GetSpans()}

	return &ctpb.ExportTraceServiceResponse{}, nil
}

type collector struct {
	t *testing.T

	endpoint     string
	traceService *collectorTraceServiceServer
}

func newCollector(t *testing.T) *collector {
	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	coll := &collector{
		t:        t,
		endpoint: ln.Addr().String(),
		traceService: &collectorTraceServiceServer{
			requests: make(chan exportRequest, 1),
		},
	}

	srv := grpc.NewServer()
	ctpb.RegisterTraceServiceServer(srv, coll.traceService)
	go func() { _ = srv.Serve(ln) }()

	t.Cleanup(srv.GracefulStop)

	return coll
}

func TestRunOTLPExporter(t *testing.T) {
	const (
		spanName = "test span"
		token    = "secret token"
	)

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req exportRequest)
	}{
		{
			desc: "WithEndpoint",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				return distro.Run(distro.WithEndpoint(url))
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
				require.Len(t, got.Spans, 1)
				assert.Equal(t, spanName, got.Spans[0].Name)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url))
				return distro.Run()
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
				require.Len(t, got.Spans, 1)
				assert.Equal(t, spanName, got.Spans[0].Name)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", url))
				return distro.Run()
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
				require.Len(t, got.Spans, 1)
				assert.Equal(t, spanName, got.Spans[0].Name)
			},
		},
		{
			desc: "WithEndpoint and WithAccessToken",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				return distro.Run(distro.WithEndpoint(url), distro.WithAccessToken(token))
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
				assert.Equal(t, []string{token}, got.Header.Get("x-sf-token"))
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url))
				t.Cleanup(distro.Setenv("SPLUNK_ACCESS_TOKEN", token))
				return distro.Run()
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
				assert.Equal(t, []string{token}, got.Header.Get("x-sf-token"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			coll := newCollector(t)

			// Explicitly set OTLP exporter.
			t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
			sdk, err := tc.setupFn(t, coll.endpoint)
			require.NoError(t, err, "should configure tracing")

			ctx := withTestingDeadline(context.Background(), t)
			_, span := otel.Tracer(tc.desc).Start(ctx, spanName)
			span.End()

			// Flush all spans from BSP.
			require.NoError(t, sdk.Shutdown(ctx))

			select {
			case <-ctx.Done():
				require.Fail(t, "test timeout out", ctx.Err())
			case got := <-coll.traceService.requests:
				tc.assertFn(t, got)
			}
		})
	}
}

func withTestingDeadline(ctx context.Context, t *testing.T) context.Context {
	d, ok := t.Deadline()
	if !ok {
		d = time.Now().Add(10 * time.Second)
	} else {
		d = d.Add(-time.Millisecond)
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, d)
	t.Cleanup(cancel)
	return ctx
}
