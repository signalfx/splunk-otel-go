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
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	testr "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonglil/buflogr"
	"go.opentelemetry.io/otel"
	ctpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	comm "go.opentelemetry.io/proto/otlp/common/v1"
	rpb "go.opentelemetry.io/proto/otlp/resource/v1"
	tpb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/distro"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const (
	spanName = "test span"
	token    = "secret token"
)

func reqHander() (<-chan *http.Request, http.HandlerFunc) {
	reqCh := make(chan *http.Request, 1)
	return reqCh, func(rw http.ResponseWriter, r *http.Request) {
		reqCh <- r
	}
}

func distroRun(t *testing.T) (distro.SDK, error) {
	l := testr.NewTestLogger(t)
	return distro.Run(distro.WithLogger(l))
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
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url))
				t.Cleanup(distro.Setenv("SPLUNK_ACCESS_TOKEN", token))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk"))
				return distroRun(t)
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

			ctx := context.Background()
			_, span := otel.Tracer(tc.desc).Start(ctx, spanName)
			span.End()

			// Flush all spans from BSP.
			require.NoError(t, sdk.Shutdown(ctx))
			tc.assertFn(t, <-reqCh)
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
	sdk, err := distroRun(t)
	require.NoError(t, err)

	ctx := context.Background()
	_, span := otel.Tracer("TestRunJaegerExporterDefault").Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	got := <-reqCh
	assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"))
}

type exportRequest struct {
	Header   metadata.MD
	Resource *rpb.Resource
	Spans    []*tpb.Span
}

type collectorTraceServiceServer struct {
	ctpb.UnimplementedTraceServiceServer

	requests chan exportRequest
}

var _ ctpb.TraceServiceServer = (*collectorTraceServiceServer)(nil)

func (ctss *collectorTraceServiceServer) Export(ctx context.Context, exp *ctpb.ExportTraceServiceRequest) (*ctpb.ExportTraceServiceResponse, error) {
	rs := exp.ResourceSpans[0]
	scopeSpans := rs.ScopeSpans[0]
	headers, _ := metadata.FromIncomingContext(ctx)

	ctss.requests <- exportRequest{
		Header:   headers,
		Resource: rs.GetResource(),
		Spans:    scopeSpans.GetSpans(),
	}

	return &ctpb.ExportTraceServiceResponse{}, nil
}

type collector struct {
	endpoint     string
	srv          *grpc.Server
	traceService *collectorTraceServiceServer
}

func newCollector(t *testing.T) *collector {
	coll, errCh, err := newCollectorAt("localhost:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		coll.srv.GracefulStop()
		require.NoError(t, <-errCh)
	})
	return coll
}

func newCollectorAt(address string) (*collector, chan error, error) {
	errCh := make(chan error, 1)

	ln, err := net.Listen("tcp", address)
	if err != nil {
		errCh <- nil
		return nil, errCh, err
	}

	coll := &collector{
		endpoint: ln.Addr().String(),
		traceService: &collectorTraceServiceServer{
			requests: make(chan exportRequest, 1),
		},
	}

	srv := grpc.NewServer()
	ctpb.RegisterTraceServiceServer(srv, coll.traceService)
	go func() { errCh <- srv.Serve(ln) }()
	coll.srv = srv

	return coll, errCh, nil
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
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://"+url))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url))
				t.Cleanup(distro.Setenv("SPLUNK_ACCESS_TOKEN", token))
				t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "otlp"))
				return distroRun(t)
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

			ctx := context.Background()
			_, span := otel.Tracer(tc.desc).Start(ctx, spanName)
			span.End()

			// Flush all spans from BSP.
			require.NoError(t, sdk.Shutdown(ctx))
			tc.assertFn(t, <-coll.traceService.requests)
		})
	}
}

func TestRunExporterDefault(t *testing.T) {
	// Start collector at default address.
	coll, errCh, err := newCollectorAt("localhost:4317")
	require.NoError(t, err, "failed to start testing collector")
	t.Cleanup(func() {
		coll.srv.GracefulStop()
		require.NoError(t, <-errCh)
	})

	sdk, err := distroRun(t)
	require.NoError(t, err)

	ctx := context.Background()
	_, span := otel.Tracer("TestRunExporterDefault").Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))
	got := <-coll.traceService.requests
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
	require.Len(t, got.Spans, 1)
	assert.Equal(t, spanName, got.Spans[0].Name)
}

func TestInvalidTraceExporter(t *testing.T) {
	coll := newCollector(t)

	// Explicitly set OTLP exporter.
	t.Cleanup(distro.Setenv("OTEL_TRACES_EXPORTER", "invalid value"))
	t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.endpoint))
	sdk, err := distroRun(t)
	require.NoError(t, err, "should configure tracing")

	ctx := context.Background()
	_, span := otel.Tracer("TestInvalidTraceExporter").Start(ctx, "test span")
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	// Ensure OTLP is used as the default when the OTEL_TRACES_EXPORTER value
	// is invalid.
	got := <-coll.traceService.requests
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
}

func TestSplunkDistroVerionAttrInResource(t *testing.T) {
	coll := newCollector(t)
	t.Cleanup(distro.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.endpoint))
	sdk, err := distroRun(t)
	require.NoError(t, err, "should configure tracing")

	ctx := context.Background()
	_, span := otel.Tracer("TestInvalidTraceExporter").Start(ctx, "test span")
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	got := <-coll.traceService.requests
	assert.Contains(t, got.Resource.GetAttributes(), &comm.KeyValue{
		Key: "splunk.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: splunkotel.Version(),
			},
		},
	})
}

func TestNoServiceWarn(t *testing.T) {
	var buf bytes.Buffer
	sdk, err := distro.Run(distro.WithLogger(buflogr.NewWithBuffer(&buf)))
	require.NoError(t, sdk.Shutdown(context.Background()))
	require.NoError(t, err)
	// INFO prefix for buflogr is verbosity level 0, our warn level.
	assert.Contains(t, buf.String(), `INFO service.name attribute is not set. Your service is unnamed and might be difficult to identify. Set your service name using the OTEL_SERVICE_NAME environment variable. For example, OTEL_SERVICE_NAME="<YOUR_SERVICE_NAME_HERE>")`)
}
