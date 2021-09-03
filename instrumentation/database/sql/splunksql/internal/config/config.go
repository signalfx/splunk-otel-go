package config

import (
	"context"
	"database/sql/driver"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentationName is the instrumentation library identifier for a Tracer.
const InstrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

// Config contains configuration options.
type Config struct {
	TracerProvider trace.TracerProvider
}

// NewConfig returns a new Config with default values.
func NewConfig() Config {
	return Config{TracerProvider: otel.GetTracerProvider()}
}

// Tracer returns an OTel Tracer from the appropriate TracerProvider.
//
// If the passed context contains a span, the TracerProvider that created the
// Tracer that created that span will be used. Otherwise, the TracerProvider
// from c is used.
func (c Config) Tracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(
			InstrumentationName,
			trace.WithInstrumentationVersion(splunkotel.Version()),
		)
	}
	return c.TracerProvider.Tracer(
		InstrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
	)
}

// WithSpan wraps the function f with a span.
func (c Config) WithSpan(ctx context.Context, name moniker.Span, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	// From the specification: span kind MUST always be CLIENT.
	opts = append(opts, trace.WithSpanKind(trace.SpanKindClient))

	var (
		err  error
		span trace.Span
	)
	ctx, span = c.Tracer(ctx).Start(ctx, name.String(), opts...)
	defer func() {
		handleErr(span, err)
		span.End()
	}()

	err = f(ctx)
	return err
}

func handleErr(span trace.Span, err error) {
	if span == nil {
		return
	}

	switch err {
	case nil:
		// Everything Okay.
	case driver.ErrSkip:
		// Expected if method not implemented, do not record these.
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
