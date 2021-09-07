package splunksql

import (
	"context"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/dsn"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

// config contains configuration options.
type config struct {
	TracerProvider trace.TracerProvider

	DBName     string
	Attributes []attribute.KeyValue
}

func newConfig(options ...Option) config {
	var c config
	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
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

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		c.TracerProvider = tp
	})
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.Attributes = append(c.Attributes, attr...)
	})
}

// withDataSource returns an Option that sets database attributes required and
// recommended by the OpenTelemetry semantic conventions.
func withDataSource(driverName, dataSourceName string) Option {
	dbname, attrs, err := dsn.Parse(driverName, dataSourceName)
	if err != nil {
		// TODO: log this error.
		return nil
	}
	return optionFunc(func(c *config) {
		c.DBName = dbname
		c.Attributes = append(c.Attributes, attrs...)
	})
}
