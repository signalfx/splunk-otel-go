package splunksql

import (
	"context"
	"database/sql/driver"

	"go.opentelemetry.io/otel/trace"
)

// otelTx is a traced version of sql.Tx
type otelTx struct {
	tx     driver.Tx
	config config
	ctx    context.Context
}

var _ driver.Tx = (*otelTx)(nil)

func newTx(ctx context.Context, tx driver.Tx, c config) *otelTx {
	return &otelTx{ctx: ctx, tx: tx, config: c}
}

// Commit traces the call to the wrapped Tx.Commit method.
func (t *otelTx) Commit() error {
	_, span := t.config.Tracer(t.ctx).Start(t.ctx, "Commit", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	err := t.tx.Commit()
	handleErr(span, err)
	return err
}

// Rollback traces the call to the wrapped Tx.Rollback method.
func (t *otelTx) Rollback() error {
	_, span := t.config.Tracer(t.ctx).Start(t.ctx, "Rollback", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	err := t.tx.Commit()
	handleErr(span, err)
	return err
}
