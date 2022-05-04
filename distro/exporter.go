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

package distro

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type traceExporterFunc func(*exporterConfig) (trace.SpanExporter, error)

// exporters maps environment variable values to trace exporter creation
// functions.
var exporters = map[string]traceExporterFunc{
	// OTLP gRPC exporter.
	"otlp": newOTLPExporter,
	// Jaeger thrift exporter.
	"jaeger-thrift-splunk": newJaegerThriftExporter,
	// None, explicitly do not set an exporter.
	"none": nil,
}

func newOTLPExporter(c *exporterConfig) (trace.SpanExporter, error) {
	var opts []otlptracegrpc.Option

	// Determine from the many source of endpoint configuration what to use as
	// the endpoint and if that endpoint is to a local target. This handles
	// the case where the endpoint is configured directly with an Option or
	// via environment variables. In the latter case, if the environment
	// variable used is one of OTEL_EXPORTER_OTLP_*ENDPOINT, the endpoint
	// returned here will be empty and it will default to the OTLP exporter
	// itself to correctly interpret the value and configure TLS.
	endpoint, local := otlpEndpoint(c.Endpoint)
	if endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	// Use the determination of the endpoint being for a local target or not
	// to correctly set default TLS settings for the connection.
	//
	// With its default configuration, the collector does not use TLS nor
	// authenticate connections with a certificate. This means that if the
	// endpoint is local, the default setup, do not expect this
	// authentication.
	//
	// However, if the endpoint is not local, and has the potential to
	// traverse the public internet, make sure to authenticate and encrypt the
	// connection.
	//
	// This logic does not handle when the endpoint is defined by
	// OTEL_EXPORTER_OTLP_*ENDPOINT environment variables. It is left to the
	// OTLP gRPC exporter to handle that.
	if c.TLSConfig != nil {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlptracegrpc.WithTLSCredentials(tlsCreds))
	} else if local {
		// Local collectors by default do not use TLS/SSL.
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if c.AccessToken != "" {
		opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
			"X-Sf-Token": c.AccessToken,
		}))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

// otlpEndpoint returns the endpoint to use for the OTLP gRPC exporter and if
// that endpoint is to a local target based on the configured value.
func otlpEndpoint(configured string) (endpoint string, local bool) {
	if configured != "" {
		return configured, isLocalhost(configured)
	}

	// Allow the exporter to interpret these environment variables directly.
	envs := []string{otelExporterOTLPEndpointKey, otelExporterOTLPTracesEndpointKey}
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			// The OTLP exporter logic uses an http/https prefix to set the
			// TLS configuration of the connection. Continue using that by
			// returning false for local so the configuration is left to the
			// exporter and there is no difference between this distro and
			// that.
			return "", false
		}
	}

	// Use the realm only if OTEL_EXPORTER_OTLP*_ENDPOINT are not defined.
	// Also, be sure to communicate local is false so the default behavior of
	// the OTLP gRPC exporter (using the system CA for authentication and
	// encryption) is used.
	if realm, ok := os.LookupEnv(splunkRealmKey); ok && notNone(realm) {
		return fmt.Sprintf(otlpRealmEndpointFormat, realm), false
	}

	// The OTel default is the same as Splunk's (localhost:4317). Return an
	// empty endpoint to signal the exporter default should be used and true
	// local to ensure no TLS authentication or encryption is used.
	return "", true
}

// isLocalhost returns if endpoint resolves to the current device.
func isLocalhost(endpoint string) bool {
	host := endpoint

	u, err := url.Parse(endpoint)
	if err == nil && u.Hostname() != "" {
		host = u.Hostname()
	} else {
		h, _, e := net.SplitHostPort(endpoint)
		if e == nil {
			host = h
		}
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if ip.IsLoopback() {
			return true
		}
	}

	return false
}

func newJaegerThriftExporter(c *exporterConfig) (trace.SpanExporter, error) {
	var opts []jaeger.CollectorEndpointOption

	if e := jaegerEndpoint(c.Endpoint); e != "" {
		opts = append(opts, jaeger.WithEndpoint(e))
	}

	if c.AccessToken != "" {
		opts = append(
			opts,
			jaeger.WithUsername("auth"),
			jaeger.WithPassword(c.AccessToken),
		)
	}

	if c.TLSConfig != nil {
		client := &http.Client{
			Transport: &http.Transport{TLSClientConfig: c.TLSConfig},
		}
		opts = append(opts, jaeger.WithHTTPClient(client))
	}

	return jaeger.New(
		jaeger.WithCollectorEndpoint(opts...),
	)
}

func jaegerEndpoint(configured string) string {
	if configured != "" {
		return configured
	}

	// Allow the exporter to interpret this environment variable directly.
	if _, ok := os.LookupEnv(otelExporterJaegerEndpointKey); ok {
		return ""
	}

	// Use the realm only if OTEL_EXPORTER_JAGER_ENDPOINT is not defined.
	if realm, ok := os.LookupEnv(splunkRealmKey); ok && notNone(realm) {
		return fmt.Sprintf(realmEndpointFormat, realm)
	}

	// Use Splunk specific default (locally running collector).
	return defaultJaegerEndpoint
}

// notNone returns if s is not empty or set to none.
func notNone(s string) bool {
	return s != "" && s != "none"
}
