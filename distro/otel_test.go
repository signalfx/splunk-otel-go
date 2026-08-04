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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonglil/buflogr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	otelt "go.opentelemetry.io/otel/trace"
	clpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	cmpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	ctpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	comm "go.opentelemetry.io/proto/otlp/common/v1"
	lpb "go.opentelemetry.io/proto/otlp/logs/v1"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	rpb "go.opentelemetry.io/proto/otlp/resource/v1"
	tpb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/signalfx/splunk-otel-go/distro"
)

const (
	spanName   = "test span"
	metricName = "test_instrument"
	logBody    = "test_log_body"
	token      = "secret token"
)

func TestMain(m *testing.M) {
	// Do not use the default exporters.
	cleanup := setenv("OTEL_TRACES_EXPORTER", "none")
	defer cleanup()
	cleanup = setenv("OTEL_METRICS_EXPORTER", "none")
	defer cleanup()
	cleanup = setenv("OTEL_LOGS_EXPORTER", "none")
	defer cleanup()

	goleak.VerifyTestMain(m)
}

func TestRunJaegerExporter(t *testing.T) {
	assertBase := func(t *testing.T, req *http.Request) {
		assert.Equal(t, "application/x-thrift", req.Header.Get("Content-type"))
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, req *http.Request)
	}{
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT case-insensitive",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Setenv("OTEL_TRACES_EXPORTER", "Jaeger-Thrift-Splunk")
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
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
			tc.setupFn(t, srv.URL)

			emitSpan(t)

			tc.assertFn(t, <-reqCh)
		})
	}
}

func TestRunJaegerExporterTLS(t *testing.T) {
	reqCh, hFunc := reqHander()
	srv := httptest.NewUnstartedServer(hFunc)
	t.Cleanup(srv.Close)
	srv.TLS = serverTLSConfig(t)
	srv.StartTLS()
	t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
	t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", srv.URL)

	emitSpan(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := <-reqCh
	assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"))
	assert.True(t, got.TLS.HandshakeComplete, "did not perform TLS exchange")
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

	t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")

	emitSpan(t)

	got := <-reqCh
	assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"))
}

func TestRunOTLPHTTPProtobufExporter(t *testing.T) {
	assertBase := func(t *testing.T, req *http.Request) {
		assert.Equal(t, "application/x-protobuf", req.Header.Get("Content-Type"))
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, req *http.Request)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT", //nolint:goconst // Repeated case labels keep the signal-specific test tables easy to scan.
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT case-insensitive", //nolint:goconst // Repeated case labels keep the signal-specific test tables easy to scan.
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "Http/Protobuf")
				t.Setenv("OTEL_TRACES_EXPORTER", "OTLP")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT with SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			},
			assertFn: func(t *testing.T, req *http.Request) {
				assertBase(t, req)
				assert.Equal(t, []string{token}, req.Header["X-Sf-Token"])
			},
		},
	}

	for _, tc := range testCases {
		if tc.assertFn == nil {
			tc.assertFn = assertBase
		}

		t.Run(tc.desc, func(t *testing.T) {
			reqCh, handler := reqHander()
			srv := httptest.NewServer(handler)
			t.Cleanup(srv.Close)

			tc.setupFn(t, srv.URL)
			emitSpan(t)

			req := <-reqCh
			tc.assertFn(t, req)
		})
	}
}

func TestRunOTLPHTTPProtobufTracesExporterTLS(t *testing.T) {
	reqCh, handler := reqHander()

	srv := httptest.NewUnstartedServer(handler)
	t.Cleanup(srv.Close)
	srv.TLS = serverTLSConfig(t)
	srv.StartTLS()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	emitSpan(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := <-reqCh
	assert.Equal(t, "application/x-protobuf", got.Header.Get("Content-type"))
	assert.True(t, got.TLS.HandshakeComplete, "did not perform TLS exchange")
}

func TestRunOTLPGRPCTracesExporter(t *testing.T) {
	assertBase := func(t *testing.T, got *spansExportRequest) {
		assertHasSpan(t, got)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, got *spansExportRequest)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT case-insensitive",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_TRACES_EXPORTER", "OTLP")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
			},
			assertFn: func(t *testing.T, got *spansExportRequest) {
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
			coll := &collector{}
			coll.Start(t)
			tc.setupFn(t, coll.Endpoint)

			emitSpan(t)

			tc.assertFn(t, coll.ExportedSpans())
		})
	}
}

func TestRunOTLPGRPCTracesExporterTLS(t *testing.T) {
	coll := &collector{TLS: true}
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)

	emitSpan(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := coll.ExportedSpans()
	assertHasSpan(t, got)
}

func TestRunTracesExporterDefault(t *testing.T) {
	// Start collector at default address.
	coll := &collector{Endpoint: "localhost:4317"} //nolint:goconst // Tests intentionally exercise the default collector endpoint literal.
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "")

	emitSpan(t)

	got := coll.ExportedSpans()
	assertHasSpan(t, got)
}

func TestInvalidTracesExporter(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	// Explicitly set OTLP exporter.
	t.Setenv("OTEL_TRACES_EXPORTER", "invalid value")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitSpan(t)

	// Ensure OTLP is used as the default when the OTEL_TRACES_EXPORTER value
	// is invalid.
	got := coll.ExportedSpans()
	assertHasSpan(t, got)
}

func TestTracesResource(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitSpan(t)

	got := coll.ExportedSpans()
	require.NotNil(t, got)
	assertResource(t, got.Resource.GetAttributes())
}

func TestWithIDGenerator(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitSpan(t, distro.WithIDGenerator(&testIDGenerator{}))

	got := coll.ExportedSpans()
	require.NotNil(t, got)
	assert.Contains(t, string(got.Spans[0].TraceId), "testtrace")
	assert.Contains(t, string(got.Spans[0].SpanId), "testspan")
}

func TestRunOTLPHTTPProtobufMetricsExporter(t *testing.T) {
	assertBase := func(t *testing.T, req *http.Request) {
		assert.Equal(t, "application/x-protobuf", req.Header.Get("Content-Type"))
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, req *http.Request)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT case-insensitive",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "Http/Protobuf")
				t.Setenv("OTEL_METRICS_EXPORTER", "OTLP")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT with SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", url)
				t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
			},
			assertFn: func(t *testing.T, req *http.Request) {
				assertBase(t, req)
				assert.Equal(t, []string{token}, req.Header["X-Sf-Token"])
			},
		},
	}

	for _, tc := range testCases {
		if tc.assertFn == nil {
			tc.assertFn = assertBase
		}

		t.Run(tc.desc, func(t *testing.T) {
			reqCh, handler := reqHander()
			srv := httptest.NewServer(handler)
			t.Cleanup(srv.Close)

			tc.setupFn(t, srv.URL)
			emitMetric(t)

			got := <-reqCh
			tc.assertFn(t, got)
		})
	}
}

func TestRunOTLPHTTPProtobufMetricsExporterTLS(t *testing.T) {
	reqCh, handler := reqHander()

	srv := httptest.NewUnstartedServer(handler)
	t.Cleanup(srv.Close)
	srv.TLS = serverTLSConfig(t)
	srv.StartTLS()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	emitMetric(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := <-reqCh
	assert.Equal(t, "application/x-protobuf", got.Header.Get("Content-type"))
	assert.True(t, got.TLS.HandshakeComplete, "did not perform TLS exchange")
}

func TestRunOTLPGRPCMetricsExporter(t *testing.T) {
	assertBase := func(t *testing.T, got *metricsExportRequest) {
		assertHasMetric(t, got, metricName)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, got *metricsExportRequest)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT case-insensitive",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_METRICS_EXPORTER", "OTLP")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
			},
			assertFn: func(t *testing.T, got *metricsExportRequest) {
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
			coll := &collector{}
			coll.Start(t)
			tc.setupFn(t, coll.Endpoint)

			emitMetric(t)

			tc.assertFn(t, coll.ExportedMetrics())
		})
	}
}

func TestRunOTLPGRPCMetricsExporterTLS(t *testing.T) {
	coll := &collector{TLS: true}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)

	emitMetric(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := coll.ExportedMetrics()
	assertHasMetric(t, got, metricName)
}

func TestRunMetricsExporterDefault(t *testing.T) {
	// Start collector at default address.
	// By default the metrics exporter is OTLP.
	coll := &collector{Endpoint: "localhost:4317"}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitMetric(t)

	got := coll.ExportedMetrics()
	assertHasMetric(t, got, metricName)
}

func TestRunMetricsExporterNone(t *testing.T) {
	// Start collector at default address.
	coll := &collector{Endpoint: "localhost:4317"}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitMetric(t)

	got := coll.ExportedMetrics()
	assert.Nil(t, got)
}

func TestInvalidMetricsExporter(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	// Explicitly set OTLP exporter.
	t.Setenv("OTEL_METRICS_EXPORTER", "invalid value")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitMetric(t)

	// Ensure OTLP is used as the default when the OTEL_TRACES_EXPORTER value
	// is invalid.
	got := coll.ExportedMetrics()
	require.NotNil(t, got)
	assertHasMetric(t, got, metricName)
}

func TestRuntimeMetrics(t *testing.T) {
	// Start collector at default address.
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)
	t.Setenv("OTEL_GO_X_DEPRECATED_RUNTIME_METRICS", "true")

	sdk, err := distroRun(t)
	require.NoError(t, err)

	// Flush all spans from SDK.
	require.NoError(t, sdk.Shutdown(context.Background()))

	got := coll.ExportedMetrics()
	assertHasMetric(t, got, "runtime.uptime")        // Deprectated metric.
	assertHasMetric(t, got, "go.memory.allocations") // New metric.
}

func TestMetricsResource(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitMetric(t)

	got := coll.ExportedMetrics()
	require.NotNil(t, got)
	assertResource(t, got.Resource.GetAttributes())
}

func TestRunOTLPHTTPProtobufLogsExporter(t *testing.T) {
	reqCh, handler := reqHander()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_LOGS_EXPORTER", "otlp")

	emitLogs(t)

	got := <-reqCh
	assert.Equal(t, "application/x-protobuf", got.Header.Get("Content-Type"))
}

func TestRunOTLPHTTPProtobufLogsExporterTLS(t *testing.T) {
	reqCh, handler := reqHander()

	srv := httptest.NewUnstartedServer(handler)
	t.Cleanup(srv.Close)
	srv.TLS = serverTLSConfig(t)
	srv.StartTLS()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	emitLogs(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := <-reqCh
	assert.Equal(t, "application/x-protobuf", got.Header.Get("Content-type"))
	assert.True(t, got.TLS.HandshakeComplete, "did not perform TLS exchange")
}

func TestRunOTLPGRPCLogsExporter(t *testing.T) {
	assertBase := func(t *testing.T, got *logsExportRequest) {
		assertHasLog(t, got, logBody)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string)
		assertFn func(t *testing.T, got *logsExportRequest)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT case-insensitive",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_LOGS_EXPORTER", "OTLP")
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
			setupFn: func(t *testing.T, url string) {
				t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
			},
		},
	}
	for _, tc := range testCases {
		if tc.assertFn == nil {
			tc.assertFn = assertBase
		}

		t.Run(tc.desc, func(t *testing.T) {
			coll := &collector{}
			coll.Start(t)
			tc.setupFn(t, coll.Endpoint)

			emitLogs(t)

			tc.assertFn(t, coll.ExportedLogs())
		})
	}
}

func TestRunOTLPGRPCLogsExporterTLS(t *testing.T) {
	coll := &collector{TLS: true}
	coll.Start(t)
	t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)

	emitLogs(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := coll.ExportedLogs()
	assertHasLog(t, got, logBody)
}

func TestRunLogsExporterDefault(t *testing.T) {
	// By default the logs exporter is none.
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_LOGS_EXPORTER", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitLogs(t)

	got := coll.ExportedLogs()
	assert.Nil(t, got)
}

func TestRunLogsExporterNone(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_LOGS_EXPORTER", "none")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitLogs(t)

	got := coll.ExportedLogs()
	assert.Nil(t, got)
}

func TestInvalidLogsExporter(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	// Explicitly set none exporter.
	t.Setenv("OTEL_LOGS_EXPORTER", "invalid value")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitLogs(t)

	// Ensure none is used as the default when the OTEL_LOGS_EXPORTER value
	// is invalid.
	got := coll.ExportedLogs()
	require.Nil(t, got)
}

func TestLogsResource(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_LOGS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	emitLogs(t)

	got := coll.ExportedLogs()
	require.NotNil(t, got)
	assertResource(t, got.Resource.GetAttributes())
}

func TestNoServiceWarn(t *testing.T) {
	var buf bytes.Buffer

	sdk, err := distro.Run(distro.WithLogger(buflogr.NewWithBuffer(&buf)))

	require.NoError(t, sdk.Shutdown(context.Background()))
	require.NoError(t, err)
	// INFO prefix for buflogr is verbosity level 0, our warn level.
	assert.Contains(t, buf.String(), `INFO The service.name resource attribute is not set. Your service is unnamed and will be difficult to identify. Set your service name using the OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES environment variable. For example, OTEL_SERVICE_NAME="<YOUR_SERVICE_NAME_HERE>".`)
}

func TestJaegerThriftSplunkWarn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
	t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", srv.URL)

	var buf bytes.Buffer
	sdk, err := distro.Run(distro.WithLogger(buflogr.NewWithBuffer(&buf)))

	require.NoError(t, sdk.Shutdown(context.Background()))
	require.NoError(t, err)
	// INFO prefix for buflogr is verbosity level 0, our warn level.
	assert.Contains(t, buf.String(), `INFO OTEL_TRACES_EXPORTER=jaeger-thrift-splunk is deprecated and may be removed in a future release. Use the default OTLP exporter instead, or set the SPLUNK_REALM and SPLUNK_ACCESS_TOKEN environment variables to send telemetry directly to Splunk Observability Cloud.`)
}

// setenv sets the value of the environment variable named by the key.
// It returns a function that rollbacks the setting.
func setenv(key, val string) func() {
	valSnapshot, ok := os.LookupEnv(key)
	os.Setenv(key, val)
	return func() {
		if ok {
			os.Setenv(key, valSnapshot)
		} else {
			os.Unsetenv(key)
		}
	}
}

func distroRun(t *testing.T, opts ...distro.Option) (distro.SDK, error) {
	l := testr.New(t)
	return distro.Run(append(opts, distro.WithLogger(l))...)
}

func reqHander() (<-chan *http.Request, http.HandlerFunc) {
	reqCh := make(chan *http.Request, 1)
	return reqCh, func(_ http.ResponseWriter, r *http.Request) {
		reqCh <- r
	}
}

func clientTLSConfig(t *testing.T) *tls.Config {
	creds, err := testTLSCredentials()
	require.NoError(t, err, "failed to generate TLS credentials")

	return &tls.Config{
		RootCAs:    creds.rootCAs,
		MinVersion: tls.VersionTLS13,
	}
}

func serverTLSConfig(t *testing.T) *tls.Config {
	creds, err := testTLSCredentials()
	require.NoError(t, err, "failed to generate TLS credentials")

	return &tls.Config{
		Certificates: []tls.Certificate{creds.certificate},
		MinVersion:   tls.VersionTLS13,
	}
}

type testTLSCredentialSet struct {
	certificate tls.Certificate
	rootCAs     *x509.CertPool
}

var testTLSCredentials = sync.OnceValues(func() (testTLSCredentialSet, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return testTLSCredentialSet{}, err
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return testTLSCredentialSet{}, err
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		return testTLSCredentialSet{}, err
	}

	certs := x509.NewCertPool()
	certs.AddCert(certificate)
	return testTLSCredentialSet{
		certificate: tls.Certificate{
			Certificate: [][]byte{der},
			PrivateKey:  key,
			Leaf:        certificate,
		},
		rootCAs: certs,
	}, nil
})

func emitSpan(t *testing.T, opts ...distro.Option) {
	sdk, err := distroRun(t, opts...)
	require.NoError(t, err)

	ctx := context.Background()
	_, span := otel.Tracer(t.Name()).Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))
}

func emitMetric(t *testing.T, opts ...distro.Option) {
	sdk, err := distroRun(t, opts...)
	require.NoError(t, err)

	ctx := context.Background()
	cnt, err := otel.GetMeterProvider().Meter(t.Name()).Int64Counter(metricName)
	require.NoError(t, err)
	cnt.Add(ctx, 123)

	// Flush all spans from SDK.
	require.NoError(t, sdk.Shutdown(ctx))
}

func emitLogs(t *testing.T, opts ...distro.Option) {
	sdk, err := distroRun(t, opts...)
	require.NoError(t, err)

	ctx := context.Background()

	var record log.Record
	record.SetBody(log.StringValue(logBody))
	global.GetLoggerProvider().Logger(t.Name()).Emit(ctx, record)

	// Flush all spans from SDK.
	require.NoError(t, sdk.Shutdown(ctx))
}

func assertHasSpan(t *testing.T, got *spansExportRequest) {
	t.Helper()

	if !assert.NotNil(t, got, "request must not be nil") {
		return
	}
	for _, m := range got.Spans {
		if m.Name == spanName {
			return
		}
	}

	// Not found. Generate assertion failure.
	var gotSpans []string
	for _, m := range got.Spans {
		gotSpans = append(gotSpans, m.Name)
	}
	assert.Failf(t, "should contain span", "want: %v, got: %v", spanName, gotSpans)
}

func assertHasMetric(t *testing.T, got *metricsExportRequest, name string) {
	t.Helper()

	if !assert.NotNil(t, got, "request must not be nil") {
		return
	}
	for _, m := range got.Metrics {
		if m.Name == name {
			return
		}
	}

	// Not found. Generate assertion failure.
	var gotMetrics []string
	for _, m := range got.Metrics {
		gotMetrics = append(gotMetrics, m.Name)
	}
	assert.Failf(t, "should contain metric", "want: %v, got: %v", name, gotMetrics)
}

func assertHasLog(t *testing.T, got *logsExportRequest, body string) {
	t.Helper()

	if !assert.NotNil(t, got, "request must not be nil") {
		return
	}
	for _, l := range got.Logs {
		if l.Body.GetStringValue() == body {
			return
		}
	}

	// Not found. Generate assertion failure.
	var gotLogs []string
	for _, l := range got.Logs {
		gotLogs = append(gotLogs, l.Body.GetStringValue())
	}
	assert.Failf(t, "should contain log", "want: %v, got: %v", body, gotLogs)
}

func assertResource(t *testing.T, attrs []*comm.KeyValue) {
	assert.Contains(t, attrs, &comm.KeyValue{
		Key: "telemetry.distro.name",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: "splunk-otel-go",
			},
		},
	}, "should have proper telemetry.distro.name value")

	assert.Contains(t, attrs, &comm.KeyValue{
		Key: "telemetry.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: distro.Version(),
			},
		},
	}, "should have proper telemetry.distro.version value")

	assert.Contains(t, attrs, &comm.KeyValue{
		Key: "splunk.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: distro.Version(),
			},
		},
	}, "should have proper splunk.distro.version value")

	var gotAttrKeys []string
	for _, attr := range attrs {
		gotAttrKeys = append(gotAttrKeys, attr.Key)
	}

	assert.Subset(t, gotAttrKeys,
		[]string{"process.pid", "process.executable.name", "process.executable.path"},
		"should contain process attributes")

	assert.Subset(t, gotAttrKeys,
		[]string{"process.runtime.name", "process.runtime.version", "process.runtime.description"},
		"should contain Go runtime attributes")
}

type (
	collector struct {
		Endpoint string
		TLS      bool

		traceService   *collectorTraceServiceServer
		metricsService *collectorMetricsServiceServer
		logsService    *collectorLogsServiceServer
		grpcSrv        *grpc.Server
	}

	collectorTraceServiceServer struct {
		ctpb.UnimplementedTraceServiceServer

		mtx  sync.Mutex
		data *spansExportRequest
	}

	spansExportRequest struct {
		Header   metadata.MD
		Resource *rpb.Resource
		Spans    []*tpb.Span
	}

	collectorMetricsServiceServer struct {
		cmpb.UnimplementedMetricsServiceServer

		mtx  sync.Mutex
		data *metricsExportRequest
	}

	metricsExportRequest struct {
		Header   metadata.MD
		Resource *rpb.Resource
		Metrics  []*mpb.Metric
	}

	collectorLogsServiceServer struct {
		clpb.UnimplementedLogsServiceServer

		mtx  sync.Mutex
		data *logsExportRequest
	}

	logsExportRequest struct {
		Header   metadata.MD
		Resource *rpb.Resource
		Logs     []*lpb.LogRecord
	}

	testIDGenerator struct{}
)

func (coll *collector) Start(t *testing.T) {
	if coll.Endpoint == "" {
		coll.Endpoint = "localhost:0"
	}
	ln, err := net.Listen("tcp", coll.Endpoint)
	require.NoError(t, err)
	coll.Endpoint = ln.Addr().String() // set actual endpoint

	coll.traceService = &collectorTraceServiceServer{}
	coll.metricsService = &collectorMetricsServiceServer{}
	coll.logsService = &collectorLogsServiceServer{}

	var opts []grpc.ServerOption
	if coll.TLS {
		creds := credentials.NewTLS(serverTLSConfig(t))
		opts = append(opts, grpc.Creds(creds))
	}

	coll.grpcSrv = grpc.NewServer(opts...)
	ctpb.RegisterTraceServiceServer(coll.grpcSrv, coll.traceService)
	cmpb.RegisterMetricsServiceServer(coll.grpcSrv, coll.metricsService)
	clpb.RegisterLogsServiceServer(coll.grpcSrv, coll.logsService)

	errCh := make(chan error, 1)

	// Serve and then stop during cleanup.
	t.Cleanup(func() {
		coll.grpcSrv.GracefulStop()
		if err := <-errCh; err != nil && err != grpc.ErrServerStopped {
			assert.NoError(t, err)
		}
	})
	go func() { errCh <- coll.grpcSrv.Serve(ln) }()
}

func (coll *collector) ExportedSpans() *spansExportRequest {
	defer coll.traceService.mtx.Unlock()
	coll.traceService.mtx.Lock()
	return coll.traceService.data
}

func (coll *collector) ExportedMetrics() *metricsExportRequest {
	defer coll.metricsService.mtx.Unlock()
	coll.metricsService.mtx.Lock()
	return coll.metricsService.data
}

func (coll *collector) ExportedLogs() *logsExportRequest {
	defer coll.logsService.mtx.Unlock()
	coll.logsService.mtx.Lock()
	return coll.logsService.data
}

func (ctss *collectorTraceServiceServer) Export(ctx context.Context, exp *ctpb.ExportTraceServiceRequest) (*ctpb.ExportTraceServiceResponse, error) {
	rs := exp.ResourceSpans[0]

	headers, _ := metadata.FromIncomingContext(ctx)

	ctss.mtx.Lock()
	defer ctss.mtx.Unlock()
	if ctss.data == nil {
		// headers and resource should be the same. set them once
		ctss.data = &spansExportRequest{
			Header:   headers,
			Resource: rs.GetResource(),
		}
	}
	// concat all spans
	for _, scopeSpans := range rs.ScopeSpans {
		ctss.data.Spans = append(ctss.data.Spans, scopeSpans.GetSpans()...)
	}

	return &ctpb.ExportTraceServiceResponse{}, nil
}

func (clss *collectorLogsServiceServer) Export(ctx context.Context, exp *clpb.ExportLogsServiceRequest) (*clpb.ExportLogsServiceResponse, error) {
	rl := exp.ResourceLogs[0]

	headers, _ := metadata.FromIncomingContext(ctx)

	clss.mtx.Lock()
	defer clss.mtx.Unlock()
	if clss.data == nil {
		// headers and resource should be the same. set them once
		clss.data = &logsExportRequest{
			Header:   headers,
			Resource: rl.GetResource(),
		}
	}
	// concat all logs
	for _, scopeLogs := range rl.ScopeLogs {
		clss.data.Logs = append(clss.data.Logs, scopeLogs.GetLogRecords()...)
	}

	return &clpb.ExportLogsServiceResponse{}, nil
}

func (cmss *collectorMetricsServiceServer) Export(ctx context.Context, exp *cmpb.ExportMetricsServiceRequest) (*cmpb.ExportMetricsServiceResponse, error) {
	rs := exp.ResourceMetrics[0]
	headers, _ := metadata.FromIncomingContext(ctx)

	cmss.mtx.Lock()
	defer cmss.mtx.Unlock()
	if cmss.data == nil {
		// headers and resource should be the same. set them once
		cmss.data = &metricsExportRequest{
			Header:   headers,
			Resource: rs.GetResource(),
		}
	}
	// concat all metrics
	for _, scopeMetrics := range rs.ScopeMetrics {
		cmss.data.Metrics = append(cmss.data.Metrics, scopeMetrics.GetMetrics()...)
	}

	return &cmpb.ExportMetricsServiceResponse{}, nil
}

func (g *testIDGenerator) NewSpanID(_ context.Context, _ otelt.TraceID) otelt.SpanID {
	sid := otelt.SpanID{}
	copy(sid[:], "testspan")
	return sid
}

func (g *testIDGenerator) NewIDs(_ context.Context) (otelt.TraceID, otelt.SpanID) {
	tid := otelt.TraceID{}
	copy(tid[:], "testtrace")
	sid := otelt.SpanID{}
	copy(sid[:], "testspan")
	return tid, sid
}
