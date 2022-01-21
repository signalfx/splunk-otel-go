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
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	ctpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tpb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/signalfx/splunk-otel-go/distro"
)

const (
	spanName = "test span"
	token    = "secret token"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func reqHander() (<-chan *http.Request, http.HandlerFunc) {
	reqCh := make(chan *http.Request, 1)
	return reqCh, func(rw http.ResponseWriter, r *http.Request) {
		reqCh <- r
	}
}

func TestRunJaegerExporter(t *testing.T) {
	assertBase := func(t *testing.T, req *http.Request) {
		assert.Equal(t, "application/x-thrift", req.Header.Get("Content-type"))
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req *http.Request)
	}{
		{
			desc: "WithTraceExporter",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
					jaeger.WithEndpoint(url),
				))
				require.NoError(t, err)
				return distro.Run(distro.WithTraceExporter(exp))
			},
		},
		{
			desc: "WithEndpoint",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distro.Run(distro.WithEndpoint(url))
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distro.Run()
			},
		},
		{
			desc: "WithEndpoint and WithAccessToken",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distro.Run(distro.WithEndpoint(url), distro.WithAccessToken(token))
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assertBase(t, got)
				user, pass, ok := got.BasicAuth()
				require.True(t, ok, "should have Basic Authentication headers")
				assert.Equal(t, "auth", user, "should have proper username")
				assert.Equal(t, token, pass, "should use the provided token as passowrd")
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url))
				t.Cleanup(distro.Setenv("SPLUNK_ACCESS_TOKEN", token))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distro.Run()
			},
			assertFn: func(t *testing.T, got *http.Request) {
				assertBase(t, got)
				user, pass, ok := got.BasicAuth()
				require.True(t, ok, "should have Basic Authentication headers")
				assert.Equal(t, "auth", user, "should have proper username")
				assert.Equal(t, token, pass, "should use the provided token as passowrd")
			},
		},
	}
	for _, tc := range testCases {
		if tc.assertFn == nil {
			tc.assertFn = assertBase
		}

		t.Run(tc.desc, func(t *testing.T) {
			reqCh, hFunc := reqHander()
			srv := httptest.NewServer(hFunc)
			t.Cleanup(srv.Close)

			sdk, err := tc.setupFn(t, srv.URL)
			require.NoError(t, err, "should configure tracing")

			ctx := withTestingDeadline(context.Background(), t)
			_, span := otel.Tracer(tc.desc).Start(ctx, spanName)
			span.End()

			// Flush all spans from BSP.
			require.NoError(t, sdk.Shutdown(ctx))

			select {
			case <-ctx.Done():
				require.Fail(t, "test timeout out", ctx.Err())
			case got := <-reqCh:
				tc.assertFn(t, got)
			}
		})
	}
}

func TestRunJaegerExporterDefault(t *testing.T) {
	reqCh, hFunc := reqHander()
	srv := httptest.NewUnstartedServer(hFunc)
	t.Cleanup(srv.Close)

	// Start server at default address.
	ln, err := net.Listen("tcp", "127.0.0.1:9080")
	require.NoError(t, err)
	srv.Listener = ln
	srv.Start()

	t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
	sdk, err := distro.Run()
	require.NoError(t, err)

	ctx := withTestingDeadline(context.Background(), t)
	_, span := otel.Tracer("TestRunJaegerExporterDefault").Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	select {
	case <-ctx.Done():
		require.Fail(t, "test timeout out", ctx.Err())
	case got := <-reqCh:
		assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"))
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
	endpoint     string
	srv          *grpc.Server
	traceService *collectorTraceServiceServer
}

func newCollector(t *testing.T) *collector {
	coll, err := newCollectorAt("localhost:0")
	require.NoError(t, err)
	t.Cleanup(coll.srv.GracefulStop)
	return coll
}

func newCollectorAt(address string) (*collector, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	coll := &collector{
		endpoint: ln.Addr().String(),
		traceService: &collectorTraceServiceServer{
			requests: make(chan exportRequest, 1),
		},
	}

	srv := grpc.NewServer()
	ctpb.RegisterTraceServiceServer(srv, coll.traceService)
	go func() { _ = srv.Serve(ln) }()
	coll.srv = srv

	return coll, nil
}

func TestRunOTLPExporter(t *testing.T) {
	assertBase := func(t *testing.T, req exportRequest) {
		assert.Equal(t, []string{"application/grpc"}, req.Header.Get("Content-type"))
		require.Len(t, req.Spans, 1)
		assert.Equal(t, spanName, req.Spans[0].Name)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req exportRequest)
	}{
		{
			desc: "WithTraceExporter",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				exp, err := otlptracegrpc.New(
					context.Background(),
					otlptracegrpc.WithEndpoint(url),
					otlptracegrpc.WithInsecure(),
				)
				require.NoError(t, err)
				return distro.Run(distro.WithTraceExporter(exp))
			},
		},
		{
			desc: "WithEndpoint",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distro.Run(distro.WithEndpoint(url))
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distro.Run()
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					url = "http://" + url
				}
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distro.Run()
			},
		},
		{
			desc: "WithEndpoint and WithAccessToken",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distro.Run(distro.WithEndpoint(url), distro.WithAccessToken(token))
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assertBase(t, got)
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
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distro.Run()
			},
			assertFn: func(t *testing.T, got exportRequest) {
				assertBase(t, got)
				assert.Equal(t, []string{token}, got.Header.Get("x-sf-token"))
			},
		},
	}
	for _, tc := range testCases {
		if tc.assertFn == nil {
			tc.assertFn = assertBase
		}

		t.Run(tc.desc, func(t *testing.T) {
			coll := newCollector(t)

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

func TestRunExporterDefault(t *testing.T) {
	// Start collector at default address.
	coll, err := newCollectorAt("localhost:4317")
	require.NoError(t, err, "failed to start testing collector")
	t.Cleanup(coll.srv.GracefulStop)

	sdk, err := distro.Run()
	require.NoError(t, err)

	ctx := withTestingDeadline(context.Background(), t)
	_, span := otel.Tracer("TestRunExporterDefault").Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	select {
	case <-ctx.Done():
		require.Fail(t, "test timeout out", ctx.Err())
	case got := <-coll.traceService.requests:
		assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
		require.Len(t, got.Spans, 1)
		assert.Equal(t, spanName, got.Spans[0].Name)
	}
}

func TestInvalidTraceExporter(t *testing.T) {
	coll := newCollector(t)

	// Explicitly set OTLP exporter.
	t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "invalid value"))
	sdk, err := distro.Run(distro.WithEndpoint(coll.endpoint))
	require.NoError(t, err, "should configure tracing")

	ctx := withTestingDeadline(context.Background(), t)
	_, span := otel.Tracer("TestInvalidTraceExporter").Start(ctx, "test span")
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	select {
	case <-ctx.Done():
		require.Fail(t, "test timeout out", ctx.Err())
	case got := <-coll.traceService.requests:
		// Ensure OTLP is used as the default when the OTEL_TRACES_EXPORTER
		// value is invalid.
		assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
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
