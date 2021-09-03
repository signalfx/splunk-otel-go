package splunksql

import (
	"context"
	"database/sql/driver"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/config"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

// otelTx is a traced version of sql.Tx
type otelTx struct {
	tx     driver.Tx
	config config.Config
	ctx    context.Context
}

var _ driver.Tx = (*otelTx)(nil)

func newTx(ctx context.Context, tx driver.Tx, c config.Config) *otelTx {
	return &otelTx{ctx: ctx, tx: tx, config: c}
}

// Commit traces the call to the wrapped Tx.Commit method.
func (t *otelTx) Commit() error {
	return t.config.WithSpan(t.ctx, moniker.Commit, func(ctx context.Context) error {
		return t.tx.Commit()
	})
}

// Rollback traces the call to the wrapped Tx.Rollback method.
func (t *otelTx) Rollback() error {
	return t.config.WithSpan(t.ctx, moniker.Rollback, func(ctx context.Context) error {
		return t.tx.Rollback()
	})
}
