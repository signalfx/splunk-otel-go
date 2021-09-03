package splunksql

import (
	"context"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
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

// tracer returns an OTel tracer from the appropriate TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c config) tracer(ctx context.Context) trace.Tracer {
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

// withSpan wraps the function f with a span.
func (c config) withSpan(ctx context.Context, name moniker.Span, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	// From the specification: span kind MUST always be CLIENT.
	opts = append(opts, trace.WithSpanKind(trace.SpanKindClient))

	var (
		err  error
		span trace.Span
	)
	ctx, span = c.tracer(ctx).Start(ctx, name.String(), opts...)
	defer func() {
		handleErr(span, err)
		span.End()
	}()

	err = f(ctx)
	return err
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
