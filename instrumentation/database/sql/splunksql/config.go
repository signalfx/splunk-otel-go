package splunksql

import (
	"context"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

// config contains configuration options.
type config struct {
	TracerProvider trace.TracerProvider
}

func newConfig(options ...Option) config {
	var c config
	for _, o := range options {
		o.apply(&c)
	}
	if c.TracerProvider == nil {
		c.TracerProvider = otel.GetTracerProvider()
	}
	return c
}

// Return an OTel Tracer from the appropriate TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// Tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c config) Tracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	return c.TracerProvider.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
}

type Option interface {
	apply(*config)
}

type tracerProviderOption struct {
	tp trace.TracerProvider
}

func (o tracerProviderOption) apply(c *config) {
	c.TracerProvider = o.tp
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return tracerProviderOption{tp: tp}
}
