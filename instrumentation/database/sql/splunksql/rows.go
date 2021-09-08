package splunksql

import (
	"context"
	"database/sql/driver"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"go.opentelemetry.io/otel/trace"
)

// otelRows traces driver.Rows functionality.
type otelRows struct {
	driver.Rows

	span   trace.Span
	config traceConfig
}

// Compile-time check otelRows implements driver.Rows.
var _ driver.Rows = (*otelRows)(nil)

func newRows(ctx context.Context, rows driver.Rows, c traceConfig) *otelRows {
	_, span := c.tracer(ctx).Start(ctx, moniker.Rows.String(), trace.WithSpanKind(trace.SpanKindClient))
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
			r.span.AddEvent(moniker.Next.String())
		}
	}()

	err := r.Rows.Next(dest)
	handleErr(r.span, err)
	return err
}
