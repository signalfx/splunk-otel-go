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
func (c config) withSpan(ctx context.Context, m moniker.Span, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	// From the specification: span kind MUST always be CLIENT.
	opts = append(opts, trace.WithSpanKind(trace.SpanKindClient))

	var (
		err  error
		span trace.Span
	)
	ctx, span = c.tracer(ctx).Start(ctx, c.spanName(m), opts...)
	defer func() {
		handleErr(span, err)
		span.End()
	}()

	err = f(ctx)
	return err
}

// spanName returns the OpenTelemetry compliant span name.
func (c config) spanName(m moniker.Span) string {
	// From the OpenTelemetry semantic conventions
	// (https://github.com/open-telemetry/opentelemetry-specification/blob/v1.6.1/specification/trace/semantic_conventions/database.md):
	//
	// > The **span name** SHOULD be set to a low cardinality value representing the statement executed on the database.
	// > It MAY be a stored procedure name (without arguments), DB statement without variable arguments, operation name, etc.
	// > Since SQL statements may have very high cardinality even without arguments, SQL spans SHOULD be named the
	// > following way, unless the statement is known to be of low cardinality:
	// > `<db.operation> <db.name>.<db.sql.table>`, provided that `db.operation` and `db.sql.table` are available.
	// > If `db.sql.table` is not available due to its semantics, the span SHOULD be named `<db.operation> <db.name>`.
	// > It is not recommended to attempt any client-side parsing of `db.statement` just to get these properties,
	// > they should only be used if the library being instrumented already provides them.
	// > When it's otherwise impossible to get any meaningful span name, `db.name` or the tech-specific database name MAY be used.
	//
	// The database/sql package does not provide the database operation nor
	// the SQL table the operation is being performed on during a call. It
	// would require client-side parsing of the statement to determine these
	// properties. Therefore, the database name is used if it is known.
	if c.DBName != "" {
		return c.DBName
	}

	// The database name is not known. Fallback to the known client-side
	// operation being performed. This will comply with the low cardinality
	// recommendation of the specification.
	return m.String()
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
