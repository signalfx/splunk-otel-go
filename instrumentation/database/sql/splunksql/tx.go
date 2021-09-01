package splunksql

import (
	"context"
	"database/sql/driver"
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
	return t.config.withClientSpan(t.ctx, commitSpan, func(ctx context.Context) error {
		return t.tx.Commit()
	})
}

// Rollback traces the call to the wrapped Tx.Rollback method.
func (t *otelTx) Rollback() error {
	return t.config.withClientSpan(t.ctx, rollbackSpan, func(ctx context.Context) error {
		return t.tx.Rollback()
	})
}
