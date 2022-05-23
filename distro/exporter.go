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
	"net/url"
	"os"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
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

	endpoint := otlpEndpoint()
	if endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	opts = append(opts, otlptracegrpc.WithInsecure())

	if c.AccessToken != "" {
		opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
			"X-Sf-Token": c.AccessToken,
		}))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

// otlpEndpoint returns the endpoint to use for the OTLP gRPC exporter
func otlpEndpoint() string {
	// Allow the exporter to interpret these environment variables directly.
	envs := []string{otelExporterOTLPEndpointKey, otelExporterOTLPTracesEndpointKey}
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return ""
		}
	}

	// Use the realm only if OTEL_EXPORTER_OTLP*_ENDPOINT are not defined.
	// Also, be sure to communicate local is false so the default behavior of
	// the OTLP gRPC exporter (using the system CA for authentication and
	// encryption) is used.
	if realm, ok := os.LookupEnv(splunkRealmKey); ok && notNone(realm) {
		return fmt.Sprintf(otlpRealmEndpointFormat, realm)
	}

	// The OTel default is the same as Splunk's (localhost:4317)
	return ""
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

	if e := jaegerEndpoint(); e != "" {
		opts = append(opts, jaeger.WithEndpoint(e))
	}

	if c.AccessToken != "" {
		opts = append(
			opts,
			jaeger.WithUsername("auth"),
			jaeger.WithPassword(c.AccessToken),
		)
	}

	return jaeger.New(
		jaeger.WithCollectorEndpoint(opts...),
	)
}

func jaegerEndpoint() string {
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
