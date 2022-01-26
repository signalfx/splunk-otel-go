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
	"net/http"
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

	if c.Endpoint == "" {
		if endpoint := func() string {
			// Allow the exporter to use environment variables if set.
			if _, ok := os.LookupEnv(otelExporterOTLPEndpointKey); ok {
				return ""
			}
			if _, ok := os.LookupEnv(otelExporterOTLPTracesEndpointKey); ok {
				return ""
			}
			return defaultOTLPEndpoint
		}(); endpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
		}
	} else {
		opts = append(opts, otlptracegrpc.WithEndpoint(c.Endpoint))
	}

	if c.UseTLS {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlptracegrpc.WithTLSCredentials(tlsCreds))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if c.AccessToken != "" {
		opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
			"X-Sf-Token": c.AccessToken,
		}))
	}

	return otlptracegrpc.New(context.Background(), opts...)
}

func newJaegerThriftExporter(c *exporterConfig) (trace.SpanExporter, error) {
	var opts []jaeger.CollectorEndpointOption

	if c.Endpoint == "" {
		if endpoint := func() string {
			// Allow the exporter to use environment variables.
			if _, ok := os.LookupEnv(otelExporterJaegerEndpointKey); ok {
				return ""
			}
			return defaultJaegerEndpoint
		}(); endpoint != "" {
			opts = append(opts, jaeger.WithEndpoint(endpoint))
		}
	} else {
		opts = append(opts, jaeger.WithEndpoint(c.Endpoint))
	}

	if c.AccessToken != "" {
		opts = append(
			opts,
			jaeger.WithUsername("auth"),
			jaeger.WithPassword(c.AccessToken),
		)
	}

	if c.UseTLS {
		client := &http.Client{
			Transport: &http.Transport{TLSClientConfig: c.TLSConfig},
		}
		opts = append(opts, jaeger.WithHTTPClient(client))
	}

	return jaeger.New(
		jaeger.WithCollectorEndpoint(opts...),
	)
}
