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
	"os"
	"strings"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	jaegerexporter "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Environment variable keys that set values of the configuration.
const (
	// Access token added to exported data.
	accessTokenKey = "SPLUNK_ACCESS_TOKEN"

	// OpenTelemetry TextMapPropagator to set as global.
	otelPropagatorsKey = "OTEL_PROPAGATORS"

	// OpenTelemetry trace exporter to use.
	otelTracesExporterKey = "OTEL_TRACES_EXPORTER"

	// OpenTelemetry exporter endpoints.
	otelExporterJaegerEndpointKey     = "OTEL_EXPORTER_JAEGER_ENDPOINT"
	otelExporterOTLPEndpointKey       = "OTEL_EXPORTER_OTLP_ENDPOINT"
	otelExporterOTLPTracesEndpointKey = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"

	// FIXME: support OTEL_SPAN_LINK_COUNT_LIMIT
	// FIXME: support OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT
)

// Default configuration values.
const (
	defaultAccessToken   = ""
	defaultPropagator    = "tracecontext,baggage"
	defaultTraceExporter = "otlp"

	defaultOTLPEndpoint   = "localhost:4317"
	defaultJaegerEndpoint = "http://127.0.0.1:9080/v1/trace"
)

type exporterConfig struct {
	AccessToken string
	Endpoint    string
}

// config is the configuration used to create and operate an SDK.
type config struct {
	Propagator propagation.TextMapPropagator

	ExportConfig      *exporterConfig
	TraceExporterFunc traceExporterFunc
}

// newConfig returns a validated config with Splunk defaults.
func newConfig(opts ...Option) (*config, error) {
	c := &config{
		ExportConfig: &exporterConfig{
			AccessToken: envOr(accessTokenKey, defaultAccessToken),
		},
	}

	for _, o := range opts {
		o.apply(c)
	}

	// Apply default field values if they were not set.
	if c.Propagator == nil {
		c.Propagator = loadPropagator(
			envOr(otelPropagatorsKey, defaultPropagator),
		)
	}
	if c.TraceExporterFunc == nil {
		c.TraceExporterFunc = loadTraceExporter(
			envOr(otelTracesExporterKey, defaultTraceExporter),
		)
	}

	return c, nil
}

type nonePropagatorType struct{ propagation.TextMapPropagator }

// nonePropagator signals the disablement of setting a global
// TextMapPropagator.
var nonePropagator = nonePropagatorType{}

// propagators maps environment variable values to TextMapPropagator creation
// functions.
var propagators = map[string]func() propagation.TextMapPropagator{
	// W3C Trace Context.
	"tracecontext": func() propagation.TextMapPropagator {
		return propagation.TraceContext{}
	},
	// W3C Baggage
	"baggage": func() propagation.TextMapPropagator {
		return propagation.Baggage{}
	},
	// B3 Single
	"b3": func() propagation.TextMapPropagator {
		return b3.New(b3.WithInjectEncoding(b3.B3SingleHeader))
	},
	// B3 Multi
	"b3multi": func() propagation.TextMapPropagator {
		return b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	},
	// Jaeger
	"jaeger": func() propagation.TextMapPropagator {
		return jaeger.Jaeger{}
	},
	// AWS X-Ray.
	"xray": func() propagation.TextMapPropagator {
		return xray.Propagator{}
	},
	// OpenTracing Trace
	"ottrace": func() propagation.TextMapPropagator {
		return ot.OT{}
	},
	// None, explicitly do not set a global propagator.
	"none": func() propagation.TextMapPropagator {
		return nonePropagator
	},
}

func loadPropagator(name string) propagation.TextMapPropagator {
	var props []propagation.TextMapPropagator
	for _, part := range strings.Split(name, ",") {
		factory, ok := propagators[part]
		if !ok {
			// Skip invalid data.
			// TODO: log this.
			continue
		}

		p := factory()
		if p == nonePropagator {
			// Make sure the disablement of the global propagator does not get
			// lost as a composite below.
			return p
		}
		props = append(props, p)
	}

	switch len(props) {
	case 0:
		// Default to "tracecontext,baggage".
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	case 1:
		return props[0]
	default:
		return propagation.NewCompositeTextMapPropagator(props...)
	}
}

type traceExporterFunc func(*exporterConfig) (sdktrace.SpanExporter, error)

// exporters maps environment variable values to trace exporter creation
// functions.
var exporters = map[string]traceExporterFunc{
	// OTLP gRPC exporter.
	"otlp": func(c *exporterConfig) (sdktrace.SpanExporter, error) {
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

		if c.AccessToken != "" {
			opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
				"X-Sf-Token": c.AccessToken,
			}))
		}

		return otlptracegrpc.New(context.Background(), opts...)
	},
	// Jaeger thrift exporter.
	"jaeger-thrift-splunk": func(c *exporterConfig) (sdktrace.SpanExporter, error) {
		var opts []jaegerexporter.CollectorEndpointOption

		if c.Endpoint == "" {
			if endpoint := func() string {
				// Allow the exporter to use environment variables.
				if _, ok := os.LookupEnv(otelExporterJaegerEndpointKey); ok {
					return ""
				}
				return defaultJaegerEndpoint
			}(); endpoint != "" {
				opts = append(opts, jaegerexporter.WithEndpoint(endpoint))
			}
		} else {
			opts = append(opts, jaegerexporter.WithEndpoint(c.Endpoint))
		}

		if c.AccessToken != "" {
			opts = append(
				opts,
				jaegerexporter.WithUsername("auth"),
				jaegerexporter.WithPassword(c.AccessToken),
			)
		}

		return jaegerexporter.New(
			jaegerexporter.WithCollectorEndpoint(opts...),
		)
	},
	// None, explicitly do not set an exporter.
	"none": nil,
}

func loadTraceExporter(name string) traceExporterFunc {
	tef, ok := exporters[name]
	if !ok {
		return exporters[defaultTraceExporter]
	}
	return tef
}

// envOr returns the environment variable value associated with key if it
// exists, otherwise it returns alt.
func envOr(key, alt string) string {
	v, ok := os.LookupEnv(key)
	if ok {
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

// WithEndpoint configures the endpoint telemetry is sent to. Passing an empty
// string results in the default value being used.
func WithEndpoint(endpoint string) Option {
	return optionFunc(func(c *config) {
		c.ExportConfig.Endpoint = endpoint
	})
}

// WithAccessToken configures the authentication token used to authenticate
// telemetry sent directly to Splunk Observability Cloud. Passing an empty
// string results in no authentication token being used, and assumes
// authentication is handled by another system.
//
// The SPLUNK_ACCESS_TOKEN environment variable value is used if this Option
// is not provided.
func WithAccessToken(accessToken string) Option {
	return optionFunc(func(c *config) {
		c.ExportConfig.AccessToken = accessToken
	})
}

// WithTraceExporter configures the OpenTelemetry trace SpanExporter used to
// deliver telemetry. This exporter is registered with the OpenTelemetry SDK
// using a batch span processor.
//
// The OTEL_TRACES_EXPORTER environment variable value is used if this Option
// is not provided. Valid values for this environment variable are "otlp" for
// an OTLP exporter, and "jaeger-thrift-splunk" for a Splunk specific Jaeger
// thrift exporter. If this environment variable is set to "none", no exporter
// is registered and Run will return an error stating this.
//
// By default, an OTLP exporter is used if this is not provided or the
// OTEL_TRACES_EXPORTER environment variable is not set.
func WithTraceExporter(e sdktrace.SpanExporter) Option {
	return optionFunc(func(c *config) {
		c.TraceExporterFunc = func(*exporterConfig) (sdktrace.SpanExporter, error) {
			return e, nil
		}
	})
}

// WithPropagator configures the OpenTelemetry TextMapPropagator set as the
// global TextMapPropagator. Passing nil will prevent any TextMapPropagator
// from being set.
//
// The OTEL_PROPAGATORS environment variable value is used if this Option is
// not provided.
//
// By default, a tracecontext and baggage TextMapPropagator is set as the
// global TextMapPropagator if this is not provided or the OTEL_PROPAGATORS
// environment variable is not set.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		if p == nil {
			// Set to nonePropagator so when environment variable overrides
			// are applied this is distinguishable from no WithPropagator
			// option being passed.
			p = nonePropagator
		}
		c.Propagator = p
	})
}
