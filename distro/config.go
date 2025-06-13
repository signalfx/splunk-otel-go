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
	"crypto/tls"
	"os"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Environment variable keys that set values of the configuration.
const (
	// Access token added to exported data.
	accessTokenKey = "SPLUNK_ACCESS_TOKEN"

	// OpenTelemetry TextMapPropagator to set as global.
	otelPropagatorsKey = "OTEL_PROPAGATORS"

	// OpenTelemetry exporter to use.
	otelTracesExporterKey  = "OTEL_TRACES_EXPORTER"
	otelMetricsExporterKey = "OTEL_METRICS_EXPORTER"
	otelLogsExporterKey    = "OTEL_LOGS_EXPORTER"

	// OpenTelemetry exporter endpoints.
	otelExporterJaegerEndpointKey      = "OTEL_EXPORTER_JAEGER_ENDPOINT"
	otelExporterOTLPEndpointKey        = "OTEL_EXPORTER_OTLP_ENDPOINT"
	otelExporterOTLPTracesEndpointKey  = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	otelExporterOTLPMetricsEndpointKey = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"
	otelExporterOTLPLogsEndpointKey    = "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"

	// OpenTelemetry OTLP exporter protocol.
	otelExporterOTLPProtocolKey        = "OTEL_EXPORTER_OTLP_PROTOCOL"
	otelTracesExporterOTLPProtocolKey  = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"

	// Logging level to set when using the default logger.
	otelLogLevelKey = "OTEL_LOG_LEVEL"

	// splunkRealmKey defines the Splunk realm to build an endpoint from.
	splunkRealmKey = "SPLUNK_REALM"
)

// Default configuration values.
const (
	defaultAccessToken     = ""
	defaultTraceExporter   = "otlp"
	defaultMetricsExporter = "otlp"
	defaultLogsExporter    = "none"
	defaultLogLevel        = "info"
	defaultOTLPProtocol    = "grpc"

	jaegerDefaultEndpoint     = "http://127.0.0.1:9080/v1/trace"
	jaegerRealmEndpointFormat = "https://ingest.%s.signalfx.com/v2/trace"

	otlpRealmTracesEndpointFormat  = "ingest.%s.signalfx.com"
	otlpRealmTracesEndpointPath    = "/v2/trace/otlp"
	otlpRealmMetricsEndpointFormat = "ingest.%s.signalfx.com"
	otlpRealmMetricsEndpointPath   = "/v2/datapoint/otlp"
)

type exporterConfig struct {
	AccessToken string
	TLSConfig   *tls.Config
}

// config is the configuration used to create and operate an SDK.
type config struct {
	Logger      logr.Logger
	Propagator  propagation.TextMapPropagator
	SpanLimits  *trace.SpanLimits
	IDGenerator trace.IDGenerator

	ExportConfig        *exporterConfig
	TracesExporterFunc  traceExporterFunc
	MetricsExporterFunc metricsExporterFunc
	LogsExporterFunc    logsExporterFunc
}

// newConfig returns a validated config with Splunk defaults.
func newConfig(opts ...Option) *config {
	c := &config{
		Logger:     logger(zapConfig(envOr(otelLogLevelKey, defaultLogLevel))),
		Propagator: autoprop.NewTextMapPropagator(),
		SpanLimits: newSpanLimits(),
		ExportConfig: &exporterConfig{
			AccessToken: envOr(accessTokenKey, defaultAccessToken),
		},
	}
	for _, o := range opts {
		o.apply(c)
	}
	c.TracesExporterFunc = tracesExporter(c.Logger)
	c.MetricsExporterFunc = metricsExporter(c.Logger)
	c.LogsExporterFunc = logsExporter(c.Logger)
	return c
}

// envOr returns the environment variable value associated with key if it
// set and not empty, otherwise it returns alt.
func envOr(key, alt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return alt
}

// Option sets a config setting value.
type Option interface {
	apply(*config)
}

// optionFunc is a functional option implementation for Option interface.
type optionFunc func(*config)

func (fn optionFunc) apply(c *config) {
	fn(c)
}

// WithTLSConfig configures the TLS configuration used by the exporter.
//
// If this option is not provided, the exporter connection will use the default
// TLS config.
func WithTLSConfig(conf *tls.Config) Option {
	return optionFunc(func(c *config) {
		c.ExportConfig.TLSConfig = conf
	})
}

// WithLogger configures the logger used by this distro.
//
// The logr.Logger provided should be configured with a verbosity enabled to
// emit Info logs of the desired level. The following log level to verbosity
// value are used.
//   - warning: 0
//   - info: 1
//   - debug: 2+
//
// By default, a zapr.Logger configured for info logging will be used if this
// is not provided.
func WithLogger(l logr.Logger) Option {
	return optionFunc(func(c *config) {
		c.Logger = l
	})
}

// WithIDGenerator configures the generator used to generate span and trace IDs.
// If this option is not provided, the SDK's default ID generator will be used.
func WithIDGenerator(g trace.IDGenerator) Option {
	return optionFunc(func(c *config) {
		c.IDGenerator = g
	})
}
