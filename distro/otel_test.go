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
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

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

	testCert = `
-----BEGIN CERTIFICATE-----
MIID9zCCAl+gAwIBAgIQLIENzVUKhcPL/ej83SYlLDANBgkqhkiG9w0BAQsFADBN
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExETAPBgNVBAsMCHR5bGVy
QHhpMRgwFgYDVQQDDA9ta2NlcnQgdHlsZXJAeGkwHhcNMjQwNDI1MjEwODQwWhcN
MjYwNzI1MjEwODQwWjA8MScwJQYDVQQKEx5ta2NlcnQgZGV2ZWxvcG1lbnQgY2Vy
dGlmaWNhdGUxETAPBgNVBAsMCHR5bGVyQHhpMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEA1J3ecYw+MKsUJ1VBo7mI6haEDdPhzlEmsChbCjdoiiEMY0j0
sY/LNuwNXB09bN/29tL5Do3fNMtWY0VbPbYhmghsVCd3HWPUApv3aFg0xbXv/kzj
Y6TXdCf5cN14yYmUPUWHstS4bFbMIri53+etysub/MVPjVOJ9jM7TIpn7wz3GlN+
i58+ATh6ZMA/HGjyyskKJV0YmWK7BydQdZx1tuFCkN95bOdFtv1fwfR867729f8q
m623ZGLoShoS6h0++KmXHNW6u/8wewSVJPFiKVpm6i2PUAWgKXjfEJF9l4Yl/fC5
UBD1svjdZCe1lf3fP9t69tBUZ+rb6eo8o4JrPQIDAQABo2QwYjAOBgNVHQ8BAf8E
BAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwHwYDVR0jBBgwFoAU2gNuInft33MS
xs9WWFK7nNiq4mIwGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMA0GCSqGSIb3
DQEBCwUAA4IBgQBwm/3jTM/eTH6lvGflue36LnH+ue/OmxqohYLL1Z51l/df34+5
3XsmGxRMBnQB+IR00Kp0YZGAF7HRFUP4jBWLs96kocK/ZMCcxyW1/rIGIbpSnke3
ti5GiE2FbpvgdJ0t8Pu7ueQmypp+EcjIfru51sGrKF4WQep0agoGgQrHpSkRKujR
1S43xudo9sgUVxsPPogHZP+Vh1rjTrd6AWUF921Nda5x+e69e8LAlQLrmNjptTOM
GuTWc9dKWnBUSV1Nkf3QOCsUkxiXv6BOMJTKi1m4pvLoODEcqKdzwTFR2ziTu9UF
I92ULCM9Xkvu3nM5NBwnoM8COvAdb3klsk9HDFLSTRToEXCiY/j3xSGVon+JxN5z
SXCtGTMcKKOY/y3tNP2t3/AqX6YgW2PMzUq6ewq0rAts4mv16T2yQT9PRriPuNv1
ftjtC8CTT4UN26863lF6utNMmQTKGpwJkStt6Ay+f7KpP2v90C/n4+eZdhJ3hygJ
xX4bkXYzAksKstc=
-----END CERTIFICATE-----
`

	testCertKey = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDUnd5xjD4wqxQn
VUGjuYjqFoQN0+HOUSawKFsKN2iKIQxjSPSxj8s27A1cHT1s3/b20vkOjd80y1Zj
RVs9tiGaCGxUJ3cdY9QCm/doWDTFte/+TONjpNd0J/lw3XjJiZQ9RYey1LhsVswi
uLnf563Ky5v8xU+NU4n2MztMimfvDPcaU36Lnz4BOHpkwD8caPLKyQolXRiZYrsH
J1B1nHW24UKQ33ls50W2/V/B9Hzrvvb1/yqbrbdkYuhKGhLqHT74qZcc1bq7/zB7
BJUk8WIpWmbqLY9QBaApeN8QkX2XhiX98LlQEPWy+N1kJ7WV/d8/23r20FRn6tvp
6jyjgms9AgMBAAECggEBALvWWQ5B0zOWAPEa7imiIWaFy9aXiVeH9EEO9hdZij72
IYaZjqxMGEGGW+s0Xe8REpNPIf+MsVRoDAMZ5BjqDleR93qFf0N0zXocpoAF5WGC
cetdrEp8+GfDZSkkULDknhVflxoTunbkW+aVAongmXafWUkAXF7EHg9nhY0Cff3+
/xjHmgc/3see9gJuyVlfL4c9EBYwWbKbNX+O5cZikHlz4goaQJUPSQXGnWjZBwEG
TQEgKN4YrufHgfw6BH1r34H9HoteeIb2ITw2C53SaSenrL97H6o3xSw+PUxbjCfz
wYc0VIDdmHyP1QgNfqQVxMc7PNkf2jCc3QGaEfL72ZUCgYEA6RTbD/1ZalUSIFpC
+7CsVrNF8P3GyfpU7gTovFshdL5cBrSYfKeAXGR41O9sutOzS5g+uFrhoTqPcfJ7
DFU3T/0aB7CipRZmzKEqaJFbVNy8vtny2GX/h8gnS9frDeu018baB3UJ823ZXlKy
ngETkb56fp6yqfpotxIsfuYRdFMCgYEA6YXfNdZmqST1mz0SdE7/xcxDuBoND2Bg
D+3YqoPFvIoNYxjLAFQ65STNFbr5nwE2LpJ2zCreXFAPrx2LfQbtRyb6IQ49BAQ2
WfaIgWho/z45ggZMm46PE2WmX1vlAKLtJmQXWtb48Upw0FcUjxwZvvhAJ+GLyTnP
4mB2VOW3sC8CgYEAlQYeofSMLO5DbqgHV8E4Dx8EQvcfGJiToLNG0tc5ZtknIapq
LJkz+t7KWurSDAm7A0ROJbUFuf4lJOeyQMpjWSMsd2o6M29FPuR0pvL5ACM6EpRr
LmzYmkogxd9BjF79+1BKqqXsPNYpOqOJTHjHuFGfJEH3XbKbLSHTjTcwRb0CgYBn
yE1JGf4KEhjtYxj+g9V8TKmT+k/vk39d1PDD50hL4HbOocgmX0jFabOTgsNJfTpd
PE57slmcYSB3TQQfDqMJSjND2ZYYYN7e4IgOKx7uwPLB7JbDU3oWvshP/QErZT1M
IJOYlY2RfungTbMfXve6PY4Vq1F6nqzbCM/OL0GRlQKBgH2IGS9GhdpcHS+4klBc
sFB7sIzs0jeCXzN//LxO10DdWgXYZ1eQZLsoOedlcsINH0c2LqI0BurUAUtk+t/f
hbjvdnAd9fjGwO5d0v+CucVB4ULfLy6Av9INvGXysuGGnxKbnW1lYYHFWcXKHpaJ
LIGDl/wyWm7itS/QV3P2EA/q
-----END PRIVATE KEY-----
`

	testCA = `
-----BEGIN CERTIFICATE-----
MIIEajCCAtKgAwIBAgIRAPVnMeZtEZ+ZXDBvKQcEFE4wDQYJKoZIhvcNAQELBQAw
TTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMREwDwYDVQQLDAh0eWxl
ckB4aTEYMBYGA1UEAwwPbWtjZXJ0IHR5bGVyQHhpMB4XDTIwMDgyNDE4MzQzNloX
DTMwMDgyNDE4MzQzNlowTTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENB
MREwDwYDVQQLDAh0eWxlckB4aTEYMBYGA1UEAwwPbWtjZXJ0IHR5bGVyQHhpMIIB
ojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAue009g17hNKKdgnex4Nc/mhc
AlyB/5XgRd3B29gOKnzlmsZsT+SMDSDABYfCwfrBFJQgnUI/La8zGpB0Ndq1tYNm
0Pd0qcyhyPNlVQiz2Pbxaw+5h7WEwLaYQOAHLbkb+htWwV8oz+AtWM+79Np5pmfo
x0dc2rJQnU1Eu7XBn9xA9rcNkU1Sc70D7ghDfBztAXCPn4NyQ6sp17z3WJFkoes1
Vf9+ri+MiiuA+FCEn5QHYfnN6dSExaEHDmaBa6ktC4Uqyi+VyXTWLgf3kgSzwww0
80SpBqzuoK09UKqiZQty2rQd2OvsgAWBqW3WB/1WOTfoCoZh/kwhMdQJKPXfBFmw
TlFVfKcCpbc4K+NXFTnXxS8fW8alzsZaHp/FrZxYO2/YJt6FwkIQ3Afum4PRzrOh
jISSg7++dp6umF0aIC8MxcIKxrv7qUnHy7pTKE6S9MtU3/gmMZntg6Xza2U5SwuY
E8z61i9XT3t0hzfWpt1Eqd7nVk182tpUGb++Ii3PAgMBAAGjRTBDMA4GA1UdDwEB
/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTaA24id+3fcxLG
z1ZYUruc2KriYjANBgkqhkiG9w0BAQsFAAOCAYEAQtZkGMBKgsA+rT2xm+NyVH4k
9eslmp79sGPg3fSwxwTQtvEt+c7Pdeam99NpuUpwcRW/gmXnMXeiw5CFcv7IwZlo
7ZWSFZdBJUNO4Ed4eiXnpM+2x0TeVhkA+pwHWd5LEoiV87ZlbmPzHByTt1O67t84
LuhIJ3T+SdI3FT7q6P2vlyvIeIxO/qeAP+9UrWUlkPdKWITDVG/UzAcpmsYrKlJq
+NVTYv9e07kI9ZmIu5yiXrxcg3M5mmoE8Lr1v6KhzhP00A5S62Pmb/U+m0LOnBEb
AvRbJMzvwawUrhjxaEbHP3f2mSMjtkH7tFYIwyKfTpJ/Vg6C0u5rnqu2coVPqHyz
GiS1IV8/EkWTW14APbVVZuyZ+24HJj2o5HtjasRN5tCpfpZhuZeEMNkaAJEZ2Vq0
Y88rFHHMBaV0M4EkH2YrvygPLsCgdhChCLW1tHuLTrrrDaB1wS/3e/zZIZtWoo1Y
zuJq2rX0p+GLO/JYG3yUpFCfR/BZbPWnGmNlDB55
-----END CERTIFICATE-----
`
)

func TestMain(m *testing.M) {
	// Do not use the default exporters.
	cleanup := setenv("OTEL_TRACES_EXPORTER", "none")
	defer cleanup()
	cleanup = setenv("OTEL_METRICS_EXPORTER", "none")
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

func TestRunOTLPTracesExporter(t *testing.T) {
	assertBase := func(t *testing.T, got *spansExportRequest) {
		asssertHasSpan(t, got)
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

func TestRunOTLPTracesExporterTLS(t *testing.T) {
	coll := &collector{TLS: true}
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)

	emitSpan(t, distro.WithTLSConfig(clientTLSConfig(t)))

	got := coll.ExportedSpans()
	asssertHasSpan(t, got)
}

func TestRunTracesExporterDefault(t *testing.T) {
	// Start collector at default address.
	coll := &collector{Endpoint: "localhost:4317"}
	coll.Start(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "")

	emitSpan(t)

	got := coll.ExportedSpans()
	asssertHasSpan(t, got)
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
	asssertHasSpan(t, got)
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

func TestRunOTLPMetricsExporter(t *testing.T) {
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

func TestRunOTLPMetricsExporterTLS(t *testing.T) {
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
	// By default the metrics exporter is NONE.
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)

	sdk, err := distroRun(t)
	require.NoError(t, err)

	// Flush all spans from SDK.
	require.NoError(t, sdk.Shutdown(context.Background()))

	got := coll.ExportedMetrics()
	assertHasMetric(t, got, "runtime.uptime")
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

func TestRunOTLPLogsExporterTLS(t *testing.T) {
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
	certs := x509.NewCertPool()
	require.True(t, certs.AppendCertsFromPEM([]byte(testCA)), "failed to add CA")

	return &tls.Config{
		RootCAs:    certs,
		MinVersion: tls.VersionTLS13,
	}
}

func serverTLSConfig(t *testing.T) *tls.Config {
	cert, err := tls.X509KeyPair([]byte(testCert), []byte(testCertKey))
	require.NoError(t, err, "failed to load X509 key pair")

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
}

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

func asssertHasSpan(t *testing.T, got *spansExportRequest) {
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
		Key: "splunk.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: distro.Version(),
			},
		},
	}, "should have proper splunk.distro.version value")

	assert.Contains(t, attrs, &comm.KeyValue{
		Key: "telemetry.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: distro.Version(),
			},
		},
	}, "should have proper telemetry.distro.version value")

	assert.Contains(t, attrs, &comm.KeyValue{
		Key: "telemetry.distro.name",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: distro.Name(),
			},
		},
	}, "should have proper telemetry.distro.name value")

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
