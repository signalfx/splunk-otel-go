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
	"testing"

	testr "github.com/go-logr/logr/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonglil/buflogr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/global"
	cmpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	ctpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	comm "go.opentelemetry.io/proto/otlp/common/v1"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	rpb "go.opentelemetry.io/proto/otlp/resource/v1"
	tpb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/distro"
)

const (
	spanName   = "test span"
	metricName = "test meter"
	token      = "secret token"

	testCert = `
-----BEGIN CERTIFICATE-----
MIIECTCCAnGgAwIBAgIQAXpqCAQzApLabw951TqZiTANBgkqhkiG9w0BAQsFADBN
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExETAPBgNVBAsMCHR5bGVy
QHhpMRgwFgYDVQQDDA9ta2NlcnQgdHlsZXJAeGkwHhcNMjIwMTIxMjI0NzM1WhcN
MjQwNDIxMjE0NzM1WjA8MScwJQYDVQQKEx5ta2NlcnQgZGV2ZWxvcG1lbnQgY2Vy
dGlmaWNhdGUxETAPBgNVBAsMCHR5bGVyQHhpMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAqxRz2pUB1N2eAev6p91aDxFmaBF9LECmCjGBTqhWyfW+L82m
IyyKLq2d/DLuRga3yX1ICOvUP1KUdrUO5XqkbYOv+vumISX4gKU41u6xo2hbATdh
/IkqvDG733p+eJ0GSXo2wI/4iNlj8II57dLlKfF9aPMtyxitvr709rRdkjWSGzYm
gNZws4y64GS4gp+OT5jW6SlR129QhkTMqkFBhTAW41+GgFGFJx9hMgjlQ+KGwqOF
3E8KJH/+/qTgwccd8UPT0GLo9uG+Gkmqqmk9QugEJVfxjn0Q3jtAJpxw2f33frSi
Xo2b/g7XSTv/vOEVIhluvAGkcOQL2Ire2/ZbrwIDAQABo3YwdDAOBgNVHQ8BAf8E
BAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwHwYDVR0jBBgwFoAUPYnE0BOCm4bs
89Xe6YmN4FKXkqcwLAYDVR0RBCUwI4IJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAA
AAAAAAAAAAABMA0GCSqGSIb3DQEBCwUAA4IBgQCM5NKkfO2s/FJKOtoPtvLzzHld
U49H9QvMCCY3glCydsePZLUOa0eAcV//hZlJkIiXOXsWs1Xs6SdO8rgbg30/Ta3Z
9J+T57+fBQN4lNn6s4HGlMn7KcZj1yJVHhlHk7Pn672yuYIUFpTXX+FQY1SXtyD1
f3j842e0wIC87O5Ge8DZg1kKD2SykWsoISeNvce1+6i6DrHMaQx/uq7rIYwqFSvn
uD6IQi/JxkJ7KjHiuSVLkahS4cfffVCI6udTZ8o03fo2eXxOLxQvnVlahHad/fKz
NeyUVdkXD61tCjvmkvkPSKSwOCxoR3lJYvsxGWtt5obs1f/Bs9MREHW+f2nsZO/L
wCy9c9zo5DR1W0NIv1vt/EF5zmMabb8EZBQlMj72tfkPCNkEiVCoEadDHQ3Cs7Yl
CmLIiGGhpDnL12DVgZ1HXFVtYlUzzAEVHhC12CnDZX/UhJgbjcnzOYGywpbtLcQE
tGJRLnLqRMHsAVjQuTrI5IoVW46o1UZmxdqMowE=
-----END CERTIFICATE-----
`

	testCertKey = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCrFHPalQHU3Z4B
6/qn3VoPEWZoEX0sQKYKMYFOqFbJ9b4vzaYjLIourZ38Mu5GBrfJfUgI69Q/UpR2
tQ7leqRtg6/6+6YhJfiApTjW7rGjaFsBN2H8iSq8Mbvfen54nQZJejbAj/iI2WPw
gjnt0uUp8X1o8y3LGK2+vvT2tF2SNZIbNiaA1nCzjLrgZLiCn45PmNbpKVHXb1CG
RMyqQUGFMBbjX4aAUYUnH2EyCOVD4obCo4XcTwokf/7+pODBxx3xQ9PQYuj24b4a
SaqqaT1C6AQlV/GOfRDeO0AmnHDZ/fd+tKJejZv+DtdJO/+84RUiGW68AaRw5AvY
it7b9luvAgMBAAECggEAXx8dF4DYJtoeG6Vwlds+Urhy/xQSTAOaDnreEbUJpHtN
htjKMi52mmOQPwS2dRsRHyfYYJV3SsoIoEQlhzR8riOVOnCnOzEYjbbP9rl67Yd9
ahy4D6jYqRDiC+zY2wk70KjyPnkOUFDG/5f4y6ICJyMNfsgPQyhymmckXvOEP2E+
7bESmRm5zHT7YPTeP8SHEb3uTGoPw6LB9xw8j/ugV03vm/SdewQcg7uOxrVybtda
X8q42GoQCvRxmLfde0SuQN1dJjz0Ug74A1An+IcFk96gLNOg7a+y2UnzMWOgxmBX
MYFUdoGl3mAjrgBBlYb4e++jx7YWcPNXDIIJc1yCGQKBgQDakE97hz5UPscRXR6Y
rmuFOejdi0s7YyAdQ3rMXmwVtx2t4jGZP9sBbQBA+DVKGPvl7z889qDxRZ4wMTAc
XY6sXxIOS54Fc4dGdKGtN9ItXhHyEtH9IQNaihO4LOu88VYbaspX0Ngduozricko
BX8xcOKb3OpIu6iF7Ys5oBjoUwKBgQDIYgxQrsyQVRwpGF+DSJoxjKKXIKUzk7A3
o3/2XFdWED/mh2YbLDRWlSkFajkoaXpdXAPJLkZLBm6PAquraHhAhcwV0nvZK+t1
OKmPzcX4X5yCxz8+1sEYDXynr5DrJI3FLQqhKhD9ar6ZZtSFLtiZI5D9fgklCu/4
tve0/TZjtQKBgQC4L9cblaycGE9wPZY0OwDXRCcO1H0w7ec5Yg2RPp09a5SyXaVI
rXxlZjNJjSJzcDyP2B/lwz18NhwKJtmRffJnJrMzotvnYiWE5XL+Y8VWgCkFZIDc
Hb8SxLu7gPekwYi8EDG28YO/AeAR+oqvlHpM8wG1MeWqJ6LsQnQKuvViiwKBgH+h
q/RsEhHQlBo80wFc6hGrYRhfi7n5VOFre6LgmCRSP1FHZqriEgggA7vWN8fcvzrd
0+99UPqSgzMF4XBRH18BmcdAhPADwHqud5oH2BPmWlsWK9uLj/wRAxgPhH+xjbdM
hBu5Ho87QWGWFMEr4HxSIhTEBXEZsVW6vLYEHnONAoGAdhHYICaM7YARqZVTOtV4
aT77hP/FDSN1ihmjWaf5R6pwzZHJWkY9+kFAm/M6XQmG7hlwtpbrf1ABno6KSuJ4
zj3TcsxZnqCaWx/fraIq39AnTxpl4C5bHEqZL3DM/6ATf+jjT1COSmcdtvOicYD+
vNEFg92FS2/s4hVfZmZcf4I=
-----END PRIVATE KEY-----
`

	testCA = `
-----BEGIN CERTIFICATE-----
MIIEaTCCAtGgAwIBAgIQN9JQ+3LGmMYFeIQYvoId/DANBgkqhkiG9w0BAQsFADBN
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExETAPBgNVBAsMCHR5bGVy
QHhpMRgwFgYDVQQDDA9ta2NlcnQgdHlsZXJAeGkwHhcNMjIwMTIxMjI0NjU3WhcN
MzIwMTIxMjI0NjU3WjBNMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0Ex
ETAPBgNVBAsMCHR5bGVyQHhpMRgwFgYDVQQDDA9ta2NlcnQgdHlsZXJAeGkwggGi
MA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQDYroHUMpGNy6/9RLd7ax6bDihD
6Bp4VyDecRQ5S2ClhZ/CEhTR1Hppu7PsRywM6R94TKVp15d9pwTnCUf/sqSvm+s4
/jarHrvq3QzMyCoRcI8E/WkxjFE78utShpKTzIbX95c99Ydqkxb1Ade1KvYGkFe5
Et6zELNngz1re9Bqr4ZV7PY4MCiQJOcfpzKTFYqWHz6I6wl4FmbuRpkfciOnEX2y
6t4rWSb4BPANkS17kywruoaLzrTQxEv8TOibmpVnhv+bJ8F17QCWob0p/5PPpRMY
Qsjn7daos/Oz3og5roRk3Ue6dm/GZ0aG77L6ED+IgfzxhdKoOQcsPINUHgDWrCMH
G85dQUDuM8yktrfSkalstc82dEt5Jwwvt82r5shDWrTmeQsPCtXeFYJlqwIJQTlK
dqjRUdhxZnEq085HQkOkq3me3RWcxTidqE5iJE99D0GirjjJ4xqFg041C6wqj822
ZN5TWIJ8jtTEzvl8a3gavxO4V8N/zOBgrZLvmpECAwEAAaNFMEMwDgYDVR0PAQH/
BAQDAgIEMBIGA1UdEwEB/wQIMAYBAf8CAQAwHQYDVR0OBBYEFD2JxNATgpuG7PPV
3umJjeBSl5KnMA0GCSqGSIb3DQEBCwUAA4IBgQBRDpP5rSCIJmr5cE5mnW649bVg
ZHte6qYpwEkWyJjGmp4bWBlIEbr4qMIcP2QqD/YhEn+/xKlKkR2IKQY4kpgSdt1l
hcTvc0zqa3tkC5BIfm7MCJcADUKWco2jYLASxDsc9piUYchXAy5g+j3+o2v7VJ4P
kGfLQq7lO6h+ZSn3AeaZmQHAZkHniBWZlL2pb4FdwrTtoTe9L/YoRygOfCw9RQVd
f50P3TGbWz1cGq2Ub/bKcX3R7cgIjeB+0iTf1fzrH5LT7Kp1kTu8f+iIWu5McaxO
Ykd4m0s/VFSi8gXZ811ilKPcjp0dHs3PN/XSKTERFa5tZLBcJlJtuIphAd3H1Hag
u3FYwV0w7+kQ7WUICi0CT4CIqPX9tcw0X6/PXOqNS7VwTgJW6XX+RAUHZhfnSWRu
q/4T2d6HhGLhRO0UFKkhSpjNevWSdvMRFjkh7VDIAsMAwxZiRBE6Hcs6yucho5Us
JeIxndiEik2efMFw+lME99JEArjHzIEOlM47coc=
-----END CERTIFICATE-----
`
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
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
				t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_JAEGER_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")
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

			emitSpan(t)
			// Flush all spans from BSP.
			ctx := context.Background()
			require.NoError(t, sdk.Shutdown(ctx))

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
	sdk, err := distroRun(
		t,
		distro.WithTLSConfig(clientTLSConfig(t)),
	)
	require.NoError(t, err)

	emitSpan(t)
	// Flush all spans from BSP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

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
	sdk, err := distroRun(t)
	require.NoError(t, err)

	emitSpan(t)
	// Flush all spans from BSP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := <-reqCh
	assert.Equal(t, "application/x-thrift", got.Header.Get("Content-type"))
}

func TestRunOTLPTracesExporter(t *testing.T) {
	assertBase := func(t *testing.T, req *spansExportRequest) {
		require.NotNil(t, req)
		assert.Equal(t, []string{"application/grpc"}, req.Header.Get("Content-type"))
		require.Len(t, req.Spans, 1)
		assert.Equal(t, spanName, req.Spans[0].Name)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req *spansExportRequest)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
				return distroRun(t)
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

			sdk, err := tc.setupFn(t, coll.Endpoint)
			require.NoError(t, err, "should configure tracing")

			emitSpan(t)
			// Flush all spans from BSP.
			ctx := context.Background()
			require.NoError(t, sdk.Shutdown(ctx))

			tc.assertFn(t, coll.SpansExportRequest())
		})
	}
}

func TestRunOTLPTracesExporterTLS(t *testing.T) {
	coll := &collector{TLS: true}
	coll.Start(t)

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)
	sdk, err := distroRun(
		t,
		distro.WithTLSConfig(clientTLSConfig(t)),
	)
	require.NoError(t, err)

	emitSpan(t)
	// Flush all spans from BSP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.SpansExportRequest()
	require.NotNil(t, got)
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
	require.Len(t, got.Spans, 1)
	assert.Equal(t, spanName, got.Spans[0].Name)
}

func TestRunTracesExporterDefault(t *testing.T) {
	// Start collector at default address.
	coll := &collector{Endpoint: "localhost:4317"}
	coll.Start(t)

	sdk, err := distroRun(t)
	require.NoError(t, err)

	emitSpan(t)
	// Flush all spans from BSP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.SpansExportRequest()
	require.NotNil(t, got)
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
	require.Len(t, got.Spans, 1)
	assert.Equal(t, spanName, got.Spans[0].Name)
}

func TestInvalidTracesExporter(t *testing.T) {
	coll := &collector{}
	coll.Start(t)

	// Explicitly set OTLP exporter.
	t.Setenv("OTEL_TRACES_EXPORTER", "invalid value")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)
	sdk, err := distroRun(t)
	require.NoError(t, err, "should configure tracing")

	ctx := context.Background()
	_, span := otel.Tracer("TestInvalidTraceExporter").Start(ctx, spanName)
	span.End()

	// Flush all spans from BSP.
	require.NoError(t, sdk.Shutdown(ctx))

	// Ensure OTLP is used as the default when the OTEL_TRACES_EXPORTER value
	// is invalid.
	got := coll.SpansExportRequest()
	require.NotNil(t, got)
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
}

func TestSplunkDistroVersionAttrInTracesResource(t *testing.T) {
	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)
	sdk, err := distroRun(t)
	require.NoError(t, err, "should configure tracing")

	emitSpan(t)
	// Flush all spans from BSP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.SpansExportRequest()
	require.NotNil(t, got)
	assert.Contains(t, got.Resource.GetAttributes(), &comm.KeyValue{
		Key: "splunk.distro.version",
		Value: &comm.AnyValue{
			Value: &comm.AnyValue_StringValue{
				StringValue: splunkotel.Version(),
			},
		},
	})
}

func TestRunOTLPMetricsExporter(t *testing.T) {
	t.Skip("TODO: not implemented")

	assertBase := func(t *testing.T, req *metricsExportRequest) {
		require.NotNil(t, req)
		assert.Equal(t, []string{"application/grpc"}, req.Header.Get("Content-type"))
		require.Len(t, req.Metrics, 1)
		assert.Equal(t, spanName, req.Metrics[0].Name)
	}

	testCases := []struct {
		desc     string
		setupFn  func(t *testing.T, url string) (distro.SDK, error)
		assertFn func(t *testing.T, req *metricsExportRequest)
	}{
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://"+url)
				t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
				return distroRun(t)
			},
		},
		{
			desc: "OTEL_EXPORTER_OTLP_ENDPOINT and SPLUNK_ACCESS_TOKEN",
			setupFn: func(t *testing.T, url string) (distro.SDK, error) {
				t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+url)
				t.Setenv("SPLUNK_ACCESS_TOKEN", token)
				t.Setenv("OTEL_TRACES_EXPORTER", "otlp")
				return distroRun(t)
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

			sdk, err := tc.setupFn(t, coll.Endpoint)
			require.NoError(t, err, "should configure tracing")

			emitMetric(t)
			// Flush all spans from BMP.
			ctx := context.Background()
			require.NoError(t, sdk.Shutdown(ctx))

			tc.assertFn(t, coll.MetricsExportRequest())
		})
	}
}

func TestRunOTLPMetricsExporterTLS(t *testing.T) {
	t.Skip("TODO: not implemented")

	coll := &collector{TLS: true}
	coll.Start(t)

	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://"+coll.Endpoint)
	sdk, err := distroRun(
		t,
		distro.WithTLSConfig(clientTLSConfig(t)),
	)
	require.NoError(t, err)

	emitMetric(t)
	// Flush all spans from BMP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.MetricsExportRequest()
	require.NotNil(t, got)
	assert.Equal(t, []string{"application/grpc"}, got.Header.Get("Content-type"))
	require.Len(t, got.Metrics, 1)
	assert.Equal(t, spanName, got.Metrics[0].Name)
}

func TestRunMetricsExporterDefault(t *testing.T) {
	// Start collector at default address.
	// By default the metrics exporter is NONE
	coll := &collector{Endpoint: "localhost:4317"}
	coll.Start(t)

	sdk, err := distroRun(t)
	require.NoError(t, err)

	emitMetric(t)
	// Flush all spans from BMP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.MetricsExportRequest()
	assert.Nil(t, got)
}

func TestSplunkDistroVersionAttrInMetricsResource(t *testing.T) {
	t.Skip("TODO: not implemented")

	coll := &collector{}
	coll.Start(t)
	t.Setenv("OTEL_METRICS_EXPORTER", "otlp")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://"+coll.Endpoint)
	sdk, err := distroRun(t)
	require.NoError(t, err, "should configure tracing")

	emitMetric(t)
	// Flush all spans from BMP.
	ctx := context.Background()
	require.NoError(t, sdk.Shutdown(ctx))

	got := coll.MetricsExportRequest()
	require.NotNil(t, got)
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

func distroRun(t *testing.T, opts ...distro.Option) (distro.SDK, error) {
	l := testr.NewTestLogger(t)
	return distro.Run(append(opts, distro.WithLogger(l))...)
}

func reqHander() (<-chan *http.Request, http.HandlerFunc) {
	reqCh := make(chan *http.Request, 1)
	return reqCh, func(rw http.ResponseWriter, r *http.Request) {
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

func emitSpan(t *testing.T) {
	ctx := context.Background()
	_, span := otel.Tracer(t.Name()).Start(ctx, spanName)
	span.End()
}

func emitMetric(t *testing.T) {
	ctx := context.Background()
	cnt, err := global.MeterProvider().Meter(t.Name()).Int64Counter(metricName)
	require.NoError(t, err)
	cnt.Add(ctx, 123)
}

type (
	collector struct {
		Endpoint string
		TLS      bool

		traceService   *collectorTraceServiceServer
		metricsService *collectorMetricsServiceServer
	}

	collectorTraceServiceServer struct {
		ctpb.UnimplementedTraceServiceServer

		requests chan spansExportRequest
	}

	spansExportRequest struct {
		Header   metadata.MD
		Resource *rpb.Resource
		Spans    []*tpb.Span
	}

	collectorMetricsServiceServer struct {
		cmpb.UnimplementedMetricsServiceServer

		requests chan metricsExportRequest
	}

	metricsExportRequest struct {
		Header   metadata.MD
		Resource *rpb.Resource
		Metrics  []*mpb.Metric
	}
)

func (coll *collector) Start(t *testing.T) {
	if coll.Endpoint == "" {
		coll.Endpoint = "localhost:0"
	}
	ln, err := net.Listen("tcp", coll.Endpoint)
	require.NoError(t, err)
	coll.Endpoint = ln.Addr().String() // set actual endpoint

	coll.traceService = &collectorTraceServiceServer{
		requests: make(chan spansExportRequest, 1),
	}
	coll.metricsService = &collectorMetricsServiceServer{
		requests: make(chan metricsExportRequest, 1),
	}

	var opts []grpc.ServerOption
	if coll.TLS {
		creds := credentials.NewTLS(serverTLSConfig(t))
		opts = append(opts, grpc.Creds(creds))
	}

	srv := grpc.NewServer(opts...)
	ctpb.RegisterTraceServiceServer(srv, coll.traceService)
	cmpb.RegisterMetricsServiceServer(srv, coll.metricsService)
	errCh := make(chan error, 1)
	t.Cleanup(func() {
		srv.GracefulStop()
		assert.NoError(t, <-errCh)
	})
	go func() { errCh <- srv.Serve(ln) }()
}

func (coll *collector) SpansExportRequest() *spansExportRequest {
	select {
	case got := <-coll.traceService.requests:
		return &got
	default:
		return nil
	}
}

func (coll *collector) MetricsExportRequest() *metricsExportRequest {
	select {
	case got := <-coll.metricsService.requests:
		return &got
	default:
		return nil
	}
}

func (ctss *collectorTraceServiceServer) Export(ctx context.Context, exp *ctpb.ExportTraceServiceRequest) (*ctpb.ExportTraceServiceResponse, error) {
	rs := exp.ResourceSpans[0]
	scopeSpans := rs.ScopeSpans[0]
	headers, _ := metadata.FromIncomingContext(ctx)

	ctss.requests <- spansExportRequest{
		Header:   headers,
		Resource: rs.GetResource(),
		Spans:    scopeSpans.GetSpans(),
	}

	return &ctpb.ExportTraceServiceResponse{}, nil
}

func (cmss *collectorMetricsServiceServer) Export(ctx context.Context, exp *cmpb.ExportMetricsServiceRequest) (*cmpb.ExportMetricsServiceResponse, error) {
	rs := exp.ResourceMetrics[0]
	scopeMetrics := rs.ScopeMetrics[0]
	headers, _ := metadata.FromIncomingContext(ctx)

	cmss.requests <- metricsExportRequest{
		Header:   headers,
		Resource: rs.GetResource(),
		Metrics:  scopeMetrics.GetMetrics(),
	}

	return &cmpb.ExportMetricsServiceResponse{}, nil
}
