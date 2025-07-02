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

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/exporters/jaeger" //nolint:staticcheck // Jaeger is deprecated, but we still support it to not break existing users.
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type traceExporterFunc func(logr.Logger, *exporterConfig) (trace.SpanExporter, error)

// traceExporters maps environment variable values to trace exporter creation
// functions.
var traceExporters = map[string]traceExporterFunc{
	// OTLP gRPC exporter.
	"otlp": newOTLPTracesExporter,
	// Jaeger thrift exporter.
	"jaeger-thrift-splunk": newJaegerThriftExporter,
	// None, explicitly do not set an exporter.
	"none": nil,
}

func tracesExporter(l logr.Logger) traceExporterFunc {
	key := envOr(otelTracesExporterKey, defaultTraceExporter)
	tef, ok := traceExporters[key]
	if !ok {
		err := fmt.Errorf("invalid %s: %q", otelTracesExporterKey, key)
		l.Error(err, "using default %s: %q", otelTracesExporterKey, defaultTraceExporter)

		return traceExporters[defaultTraceExporter]
	}
	return tef
}

func newOTLPTracesExporter(l logr.Logger, c *exporterConfig) (trace.SpanExporter, error) {
	ctx := context.Background()

	splunkEndpoint := otlpRealmTracesEndpoint()
	if splunkEndpoint != "" {
		// Direct ingest to Splunk Observabilty Cloud using HTTP/protobuf.
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(splunkEndpoint),
			otlptracehttp.WithURLPath(otlpRealmTracesEndpointPath),
			otlptracehttp.WithHeaders(map[string]string{
				"X-Sf-Token": c.AccessToken,
			}),
		)
	}

	headers := make(map[string]string)
	if c.AccessToken != "" {
		headers["X-Sf-Token"] = c.AccessToken
	}
	isLocalCollector := noneEnvVarSet(otelExporterOTLPEndpointKey, otelExporterOTLPTracesEndpointKey, splunkRealmKey)
	protocol := otlpProtocol(l, otelTracesExporterOTLPProtocolKey)

	if protocol == otlpProtocolHTTPProtobuf {
		var opts []otlptracehttp.Option

		if len(headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(headers))
		}

		if c.TLSConfig != nil {
			opts = append(opts, otlptracehttp.WithTLSClientConfig(c.TLSConfig))
		} else if isLocalCollector {
			// Assume that the default endpoint (local collector) is non-TLS.
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		return otlptracehttp.New(ctx, opts...)
	}

	var opts []otlptracegrpc.Option

	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	if c.TLSConfig != nil {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlptracegrpc.WithTLSCredentials(tlsCreds))
	} else if isLocalCollector {
		// Assume that the default endpoint (local collector) is non-TLS.
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// otlpRealmTracesEndpoint returns the endpoint to use for the OTLP HTTP/protobuf traces exporter.
func otlpRealmTracesEndpoint() string {
	// Allow the exporter to interpret these environment variables directly.
	envs := []string{otelExporterOTLPEndpointKey, otelExporterOTLPTracesEndpointKey}
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return ""
		}
	}

	// Use the realm only if OTEL_EXPORTER_OTLP*_ENDPOINT are not defined.
	// Also, be sure to communicate local is false so the default behavior of
	// the OTLP HTTP/protobuf exporter (using the system CA for authentication and
	// encryption) is used.
	if realm, ok := os.LookupEnv(splunkRealmKey); ok && notNone(realm) {
		return fmt.Sprintf(otlpRealmTracesEndpointFormat, realm)
	}

	// The OTel default is the same as Splunk's (localhost:4317)
	return ""
}

// otlpRealmMetricsEndpoint returns the endpoint to use for the OTLP HTTP/protobuf metrics exporter.
func otlpRealmMetricsEndpoint() string {
	// Allow the exporter to interpret these environment variables directly.
	envs := []string{otelExporterOTLPEndpointKey, otelExporterOTLPMetricsEndpointKey}
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return ""
		}
	}

	// Use the realm only if OTEL_EXPORTER_OTLP*_ENDPOINT are not defined.
	// Also, be sure to communicate local is false so the default behavior of
	// the OTLP HTTP/protobuf exporter (using the system CA for authentication and
	// encryption) is used.
	if realm, ok := os.LookupEnv(splunkRealmKey); ok && notNone(realm) {
		return fmt.Sprintf(otlpRealmMetricsEndpointFormat, realm)
	}

	// The OTel default is the same as Splunk's (localhost:4317)
	return ""
}

func newJaegerThriftExporter(l logr.Logger, c *exporterConfig) (trace.SpanExporter, error) {
	l.Info("OTEL_TRACES_EXPORTER=jaeger-thrift-splunk is deprecated and may be removed in a future release. Use the default OTLP exporter instead, or set the SPLUNK_REALM and SPLUNK_ACCESS_TOKEN environment variables to send telemetry directly to Splunk Observability Cloud.")

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
		return fmt.Sprintf(jaegerRealmEndpointFormat, realm)
	}

	// Use Splunk specific default (locally running collector).
	return jaegerDefaultEndpoint
}

type metricsExporterFunc func(logr.Logger, *exporterConfig) (metric.Exporter, error)

// metricsExporters maps environment variable values to metrics exporter creation
// functions.
var metricsExporters = map[string]metricsExporterFunc{
	// OTLP gRPC exporter.
	"otlp": newOTLPMetricsExporter,
	// None, explicitly do not set an exporter.
	"none": nil,
}

func metricsExporter(l logr.Logger) metricsExporterFunc {
	key := envOr(otelMetricsExporterKey, defaultMetricsExporter)
	mef, ok := metricsExporters[key]
	if !ok {
		err := fmt.Errorf("invalid %s: %q", otelMetricsExporterKey, key)
		l.Error(err, "using default %s: %q", otelMetricsExporterKey, defaultMetricsExporter)

		return metricsExporters[defaultMetricsExporter]
	}
	return mef
}

func newOTLPMetricsExporter(l logr.Logger, c *exporterConfig) (metric.Exporter, error) {
	ctx := context.Background()

	splunkEndpoint := otlpRealmMetricsEndpoint()
	if splunkEndpoint != "" {
		// Direct ingest to Splunk Observabilty Cloud using HTTP/protobuf.
		return otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(splunkEndpoint),
			otlpmetrichttp.WithURLPath(otlpRealmMetricsEndpointPath),
			otlpmetrichttp.WithHeaders(map[string]string{
				"X-Sf-Token": c.AccessToken,
			}),
		)
	}

	headers := make(map[string]string)
	if c.AccessToken != "" {
		headers["X-Sf-Token"] = c.AccessToken
	}
	isLocalCollector := noneEnvVarSet(otelExporterOTLPEndpointKey, otelExporterOTLPMetricsEndpointKey, splunkRealmKey)
	protocol := otlpProtocol(l, otelMetricsExporterOTLPProtocolKey)

	if protocol == otlpProtocolHTTPProtobuf {
		var opts []otlpmetrichttp.Option

		if len(headers) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(headers))
		}

		if c.TLSConfig != nil {
			opts = append(opts, otlpmetrichttp.WithTLSClientConfig(c.TLSConfig))
		} else if isLocalCollector {
			// Assume that the default endpoint (local collector) is non-TLS.
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		return otlpmetrichttp.New(ctx, opts...)
	}

	var opts []otlpmetricgrpc.Option

	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}

	if c.TLSConfig != nil {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(tlsCreds))
	} else if isLocalCollector {
		// Assume that the default endpoint (local collector) is non-TLS.
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

type logsExporterFunc func(*exporterConfig) (log.Exporter, error)

// logsExporters maps environment variable values to logs exporter creation
// functions.
var logsExporters = map[string]logsExporterFunc{
	// OTLP gRPC exporter.
	"otlp": newOTLPlogExporter,
	// None, explicitly do not set an exporter.
	"none": nil,
}

func logsExporter(l logr.Logger) logsExporterFunc {
	key := envOr(otelLogsExporterKey, defaultLogsExporter)
	lef, ok := logsExporters[key]
	if !ok {
		err := fmt.Errorf("invalid %s: %q", otelLogsExporterKey, key)
		l.Error(err, "using default %s: %q", otelLogsExporterKey, defaultLogsExporter)

		return logsExporters[defaultLogsExporter]
	}
	return lef
}

func newOTLPlogExporter(c *exporterConfig) (log.Exporter, error) {
	ctx := context.Background()

	var opts []otlploggrpc.Option

	// SPLUNK_REALM is not supported, Splunk Observability ingest does not support OTLP.
	if c.AccessToken != "" {
		opts = append(opts, otlploggrpc.WithHeaders(map[string]string{
			"X-Sf-Token": c.AccessToken,
		}))
	}

	if c.TLSConfig != nil {
		tlsCreds := credentials.NewTLS(c.TLSConfig)
		opts = append(opts, otlploggrpc.WithTLSCredentials(tlsCreds))
	} else if noneEnvVarSet(otelExporterOTLPEndpointKey, otelExporterOTLPLogsEndpointKey) {
		// Assume that the default endpoint (local collector) is non-TLS.
		opts = append(opts, otlploggrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	return otlploggrpc.New(ctx, opts...)
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

//nolint:unparam // This will receive other input values in future.
func otlpProtocol(l logr.Logger, signalKey string) string {
	// Signal-specific key takes precedence.
	if v := os.Getenv(signalKey); v != "" {
		if v == otlpProtocolGRPC || v == otlpProtocolHTTPProtobuf {
			return v
		}
		err := fmt.Errorf("invalid %s: %q", signalKey, v)
		l.Error(err, "falling back to %q", otelExporterOTLPProtocolKey)
	}

	// Fallback to general OTLP protocol.
	if v := os.Getenv(otelExporterOTLPProtocolKey); v != "" {
		if v == otlpProtocolGRPC || v == otlpProtocolHTTPProtobuf {
			return v
		}
		err := fmt.Errorf("invalid %s: %q", otelExporterOTLPProtocolKey, v)
		l.Error(err, "using default %s: %q", otelExporterOTLPProtocolKey, defaultOTLPProtocol)
	}

	return defaultOTLPProtocol
}
