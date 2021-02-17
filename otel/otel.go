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
Package otel provides functionality to quickly setup the OpenTelemetry Go
implementation with useful Splunk defaults.

The default configuration will correctly set the default OpenTelemetry SDK to
propagate traces and export all spans to a locally running Splunk Smart Agent.
Information about the Splunk Smart Agent can be found here
https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-smart-agent.html
*/
package otel

import (
	"context"

	"go.opentelemetry.io/contrib/propagators/b3"
	global "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SDK contains all OpenTelemetry SDK state and provides access to this state.
type SDK struct {
	config config

	shutdownFunc func(context.Context) error
}

func (s SDK) Shutdown(ctx context.Context) error {
	return s.shutdownFunc(ctx)
}

// Run configures the default OpenTelemetry SDK and installs it globally.
func Run(opts ...Option) (SDK, error) {
	c, err := newConfig(opts...)
	if err != nil {
		return SDK{}, err
	}

	exp, err := jaeger.NewRawExporter(
		jaeger.WithCollectorEndpoint(c.Endpoint),
	)
	if err != nil {
		return SDK{}, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithConfig(trace.Config{
			DefaultSampler: trace.AlwaysSample(),
		}),
		// TODO: configure batching policy with configured values.
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exp)),
	)
	global.SetTracerProvider(traceProvider)

	// TODO: add and honor option to set additional propagators.
	global.SetTextMapPropagator(b3.B3{})

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
