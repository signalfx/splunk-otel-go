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
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
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

	// Logging level to set when using the default logger.
	otelLogLevelKey = "OTEL_LOG_LEVEL"

	// splunkMetricsEndpointKey defines the endpoint Splunk specific metrics
	// are sent. This is not currently supported.
	splunkMetricsEndpointKey = "SPLUNK_METRICS_ENDPOINT"

	// splunkRealmKey defines the Splunk realm to build an endpoint from.
	splunkRealmKey = "SPLUNK_REALM"
)

// Default configuration values.
const (
	defaultAccessToken   = ""
	defaultPropagator    = "tracecontext,baggage"
	defaultTraceExporter = "otlp"
	defaultLogLevel      = "info"

	defaultJaegerEndpoint = "http://127.0.0.1:9080/v1/trace"

	realmEndpointFormat     = "https://ingest.%s.signalfx.com/v2/trace"
	otlpRealmEndpointFormat = "ingest.%s.signalfx.com:443"
)

type exporterConfig struct {
	AccessToken string
}

// config is the configuration used to create and operate an SDK.
type config struct {
	Logger     logr.Logger
	Propagator propagation.TextMapPropagator
	SpanLimits *trace.SpanLimits

	ExportConfig      *exporterConfig
	TraceExporterFunc traceExporterFunc
}

// newConfig returns a validated config with Splunk defaults.
func newConfig(opts ...Option) *config {
	c := &config{
		Logger:     logger(zapConfig(envOr(otelLogLevelKey, defaultLogLevel))),
		SpanLimits: newSpanLimits(),
		ExportConfig: &exporterConfig{
			AccessToken: envOr(accessTokenKey, defaultAccessToken),
		},
	}

	for _, o := range opts {
		o.apply(c)
	}

	// Apply default field values if they were not set.
	if c.Propagator == nil {
		c.setPropagator(envOr(otelPropagatorsKey, defaultPropagator))
	}
	if c.TraceExporterFunc == nil {
		key := envOr(otelTracesExporterKey, defaultTraceExporter)
		tef, ok := exporters[key]
		if !ok {
			err := fmt.Errorf("invalid exporter: %q", key)
			c.Logger.Error(err, "using default trace exporter: otlp")

			tef = exporters[defaultTraceExporter]
		}
		c.TraceExporterFunc = tef
	}

	return c
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

func (c *config) setPropagator(name string) {
	var props []propagation.TextMapPropagator
	for _, part := range strings.Split(name, ",") {
		factory, ok := propagators[part]
		if !ok {
			// Skip invalid data.
			c.Logger.Info("invalid propagator name", "name", part)
			continue
		}

		p := factory()
		if p == nonePropagator {
			// Make sure the disablement of the global propagator does not get
			// lost as a composite below.
			c.Propagator = p
			return
		}
		props = append(props, p)
	}

	switch len(props) {
	case 0:
		// Default to "tracecontext,baggage".
		c.Propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	case 1:
		c.Propagator = props[0]
	default:
		c.Propagator = propagation.NewCompositeTextMapPropagator(props...)
	}
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
