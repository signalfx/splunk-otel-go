package splunksql

import (
	"context"
	"database/sql/driver"

	"go.opentelemetry.io/otel/trace"
)

// otelRows traces driver.Rows functionality.
type otelRows struct {
	driver.Rows

	span   trace.Span
	config config
}

// Compile-time check otelRows implements driver.Rows.
var _ driver.Rows = (*otelRows)(nil)

func newRows(ctx context.Context, rows driver.Rows, c config) *otelRows {
	_, span := c.Tracer(ctx).Start(ctx, "Rows", trace.WithSpanKind(trace.SpanKindClient))
	return &otelRows{
		Rows:   rows,
		span:   span,
		config: c,
	}
}

func (r otelRows) Close() error {
	defer func() {
		if r.span != nil {
			r.span.End()
		}
	}()

	err := r.Rows.Close()
	handleErr(r.span, err)
	return err
}

func (r otelRows) Next(dest []driver.Value) error {
	defer func() {
		if r.span != nil {
			r.span.AddEvent("Next")
		}
	}()

	err := r.Rows.Next(dest)
	handleErr(r.span, err)
	return err
}
