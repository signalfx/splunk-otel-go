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

/*
Package distro provides functionality to quickly setup the OpenTelemetry Go
implementation with useful Splunk defaults.

The default configuration sets the default OpenTelemetry SDK to propagate
traces using a W3C tracecontext and W3C baggage propagator and export all
spans to a locally running Splunk OpenTelemetry Connector.
*/
package distro

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

var distroVerAttr = attribute.String("splunk.distro.version", splunkotel.Version())

const noServiceWarn = `service.name attribute is not set. Your service is unnamed and might be difficult to identify. Set your service name using the OTEL_SERVICE_NAME environment variable. For example, OTEL_SERVICE_NAME="<YOUR_SERVICE_NAME_HERE>")`

// SDK contains all OpenTelemetry SDK state and provides access to this state.
type SDK struct {
	config config

	shutdownFunc func(context.Context) error
}

// Shutdown stops the SDK and releases any used resources.
func (s SDK) Shutdown(ctx context.Context) error {
	if s.shutdownFunc != nil {
		return s.shutdownFunc(ctx)
	}
	return nil
}

// Run configures the default OpenTelemetry SDK and installs it globally.
//
// It is the callers responsibility to shut down the returned SDK when
// complete. This ensures all resources are released and all telemetry
// flushed.
func Run(opts ...Option) (SDK, error) {
	c := newConfig(opts...)

	// Unify the SDK logging with OTel.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(e error) {
		c.Logger.Error(e, "OpenTelemetry error")
	}))
	otel.SetLogger(c.Logger)

	// SPLUNK_METRICS_ENDPOINT is currently not supported, log this fact.
	if _, ok := os.LookupEnv(splunkMetricsEndpointKey); ok {
		c.Logger.Info("SPLUNK_METRICS_ENDPOINT set; not supported by this distro")
	}

	if c.Propagator != nil && c.Propagator != nonePropagator {
		otel.SetTextMapPropagator(c.Propagator)
	}

	if c.TraceExporterFunc == nil {
		c.Logger.V(1).Info("OTEL_TRACES_EXPORTER set to none/nil: Tracing disabled")
		// "none" exporter configured.
		return SDK{}, nil
	}
	exp, err := c.TraceExporterFunc(c.ExportConfig)
	if err != nil {
		return SDK{}, err
	}

	res, err := resource.Merge(
		resource.Default(),
		// Use a schema-less Resource here, uses resource.Default's.
		resource.NewSchemaless(distroVerAttr),
	)
	if err != nil {
		return SDK{}, err
	}
	if !serviceNameDefined(res) {
		c.Logger.Info(noServiceWarn)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
		// TODO: configure batching policy with configured values.
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exp)),
	)
	otel.SetTracerProvider(traceProvider)

	return SDK{
		config: *c,
		shutdownFunc: func(ctx context.Context) error {
			if err := traceProvider.Shutdown(ctx); err != nil {
				return err
			}
			return exp.Shutdown(ctx)
		},
	}, nil
}

func serviceNameDefined(r *resource.Resource) bool {
	val, ok := r.Set().Value(semconv.ServiceNameKey)
	return ok && val.Type() == attribute.STRING && !strings.HasPrefix(val.AsString(), "unknown_service:")
}
