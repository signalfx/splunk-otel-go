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
	"errors"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

var errShutdown = errors.New("SDK shutdown failure")

// Deprecated
const splunkVerAttr = "splunk.distro.version"

const (
	distroVerAttr  = "telemetry.distro.version"
	distroNameAttr = "telemetry.distro.name"
)

const tracesSamplerKey = "OTEL_TRACES_SAMPLER"

const noServiceWarn = `The service.name resource attribute is not set. Your service is unnamed and will be difficult to identify. Set your service name using the OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES environment variable. For example, OTEL_SERVICE_NAME="<YOUR_SERVICE_NAME_HERE>".`

// SDK is the Splunk distribution of the OpenTelemetry SDK.
type SDK struct {
	shutdownFuncs []shutdownFunc
}

type shutdownFunc func(context.Context) error

// Shutdown stops the SDK and releases any used resources.
func (s SDK) Shutdown(ctx context.Context) error {
	var retErr error
	// Calling shutdownFuncs sequentially for sake of simplicity.
	for _, fn := range s.shutdownFuncs {
		if err := fn(ctx); err != nil {
			// Each error can have different cause therefore we are logging them via otel.Handle.
			otel.Handle(err)
			// We are returning a sentinel error when any shutdown error happens.
			retErr = errShutdown
		}
	}
	return retErr
}

// Run configures the default OpenTelemetry SDK and installs it globally.
//
// It is the callers responsibility to shut down the returned SDK when
// complete. This ensures all resources are released and all telemetry
// flushed.
func Run(opts ...Option) (SDK, error) {
	ctx := context.Background()
	c := newConfig(opts...)

	// Unify the SDK logging with OTel.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(e error) {
		c.Logger.Error(e, "OpenTelemetry error")
	}))
	otel.SetLogger(c.Logger)

	res, err := newResource(ctx)
	if err != nil {
		return SDK{}, err
	}

	if !serviceNameDefined(res) {
		c.Logger.Info(noServiceWarn)
	}

	otel.SetTextMapPropagator(c.Propagator)

	sdk := SDK{}

	shutdownFn, err := runTraces(c, res)
	if err != nil {
		sdk.Shutdown(ctx) //nolint:errcheck // the Shutdown errors are logged
		return SDK{}, err
	}
	if shutdownFn != nil {
		sdk.shutdownFuncs = append(sdk.shutdownFuncs, shutdownFn)
	}

	shutdownFn, err = runMetrics(c, res)
	if err != nil {
		sdk.Shutdown(ctx) //nolint:errcheck // the Shutdown errors are logged
		return SDK{}, err
	}
	if shutdownFn != nil {
		sdk.shutdownFuncs = append(sdk.shutdownFuncs, shutdownFn)
	}

	shutdownFn, err = runLogs(c, res)
	if err != nil {
		sdk.Shutdown(ctx) //nolint:errcheck // the Shutdown errors are logged
		return SDK{}, err
	}
	if shutdownFn != nil {
		sdk.shutdownFuncs = append(sdk.shutdownFuncs, shutdownFn)
	}

	return sdk, nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	// SDK's default resource.
	res := resource.Default()

	// Add process, Go runtime, and container information.
	procRes, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithContainer(),
	)
	if err != nil {
		return nil, err
	}
	res, err = resource.Merge(res, procRes)
	if err != nil {
		return nil, err
	}

	// Add Splunk-specific attributes.
	splunkVerRes := resource.NewSchemaless(attribute.String(splunkVerAttr, Version()))
	res, err = resource.Merge(res, splunkVerRes)
	if err != nil {
		return nil, err
	}

	distroVerRes := resource.NewSchemaless(attribute.String(distroVerAttr, Version()))
	res, err = resource.Merge(res, distroVerRes)
	if err != nil {
		return nil, err
	}

	distroNameRes := resource.NewSchemaless(attribute.String(distroNameAttr, Name()))
	res, err = resource.Merge(res, distroNameRes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func runTraces(c *config, res *resource.Resource) (shutdownFunc, error) {
	if c.TracesExporterFunc == nil {
		c.Logger.V(1).Info("OTEL_TRACES_EXPORTER set to none: Tracing disabled")
		// "none" exporter configured.
		return nil, nil
	}

	exp, err := c.TracesExporterFunc(c.Logger, c.ExportConfig)
	if err != nil {
		return nil, err
	}
	o := []trace.TracerProviderOption{
		trace.WithResource(res),
		trace.WithRawSpanLimits(*c.SpanLimits),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exp)),
		trace.WithIDGenerator(c.IDGenerator),
	}
	if _, ok := os.LookupEnv(tracesSamplerKey); !ok {
		o = append(o, trace.WithSampler(trace.AlwaysSample()))
	}

	traceProvider := trace.NewTracerProvider(o...)
	otel.SetTracerProvider(traceProvider)

	return traceProvider.Shutdown, nil
}

func runMetrics(c *config, res *resource.Resource) (shutdownFunc, error) {
	if c.MetricsExporterFunc == nil {
		c.Logger.V(1).Info("OTEL_METRICS_EXPORTER set to none: Metrics disabled")
		// "none" exporter configured.
		return nil, nil
	}

	exp, err := c.MetricsExporterFunc(c.ExportConfig)
	if err != nil {
		return nil, err
	}

	o := []metric.Option{
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)),
	}

	provider := metric.NewMeterProvider(o...)
	otel.SetMeterProvider(provider)

	// Add runtime metrics instrumentation.
	if err := runtime.Start(); err != nil {
		return nil, err
	}

	return provider.Shutdown, nil
}

func runLogs(c *config, res *resource.Resource) (shutdownFunc, error) {
	if c.LogsExporterFunc == nil {
		c.Logger.V(1).Info("OTEL_LOGS_EXPORTER set to none: Logs disabled")
		// "none" exporter configured.
		return nil, nil
	}

	exp, err := c.LogsExporterFunc(c.ExportConfig)
	if err != nil {
		return nil, err
	}

	o := []log.LoggerProviderOption{
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exp)),
	}

	provider := log.NewLoggerProvider(o...)
	global.SetLoggerProvider(provider)

	return provider.Shutdown, nil
}

func serviceNameDefined(r *resource.Resource) bool {
	val, ok := r.Set().Value("service.name")
	return ok && val.Type() == attribute.STRING && !strings.HasPrefix(val.AsString(), "unknown_service:")
}
