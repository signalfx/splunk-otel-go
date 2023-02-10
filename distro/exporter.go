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
	"net/http"
	"os"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type traceExporterFunc func(*exporterConfig) (trace.SpanExporter, error)

// exporters maps environment variable values to trace exporter creation
// functions.
var exporters = map[string]traceExporterFunc{
	// OTLP gRPC exporter.
	"otlp": newOTLPTracesExporter,
	// Jaeger thrift exporter.
	"jaeger-thrift-splunk": newJaegerThriftExporter,
	// None, explicitly do not set an exporter.
	"none": nil,
}

func newOTLPTracesExporter(c *exporterConfig) (trace.SpanExporter, error) {
	var opts []otlptracegrpc.Option

	endpoint := otlpTracesEndpoint()
	if endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	if c.AccessToken != "" {
		opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
			"X-Sf-Token": c.AccessToken,
		}))
	}

	if c.TLSConfig != nil {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlptracegrpc.WithTLSCredentials(tlsCreds))
	} else if noneEnvVarSet(otelExporterOTLPEndpointKey, otelExporterOTLPTracesEndpointKey, splunkRealmKey) {
		// Assume that the default endpoint (local collector) is non-TLS.
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

// otlpTracesEndpoint returns the endpoint to use for the OTLP gRPC exporter.
func otlpTracesEndpoint() string {
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

// noneEnvVarSet returns true if none of provided env vars is set.
func noneEnvVarSet(envs ...string) bool {
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return false
		}
	}
	return true
}

// notNone returns if s is not empty or set to none.
func notNone(s string) bool {
	return s != "" && s != "none"
}
