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

The default configuration sets the default OpenTelemetry SDK to propagate traces
using B3 propagator and export all spans to a locally running
Splunk OpenTelemetry Connector (http://localhost:14268/api/traces).
*/
package distro

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

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

	if c.Propagator != nil && c.Propagator != nonePropagator {
		otel.SetTextMapPropagator(c.Propagator)
	}

	if c.TraceExporterFunc == nil {
		// "none" exporter configured.
		return SDK{}, nil
	}
	exp, err := c.TraceExporterFunc(c.ExportConfig)
	if err != nil {
		return SDK{}, err
	}

	traceProvider := trace.NewTracerProvider(
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
