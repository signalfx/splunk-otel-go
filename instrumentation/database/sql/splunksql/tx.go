// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package splunksql

import (
	"context"
	"database/sql/driver"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

// otelTx is a traced version of sql.Tx
type otelTx struct {
	tx     driver.Tx
	config traceConfig
	ctx    context.Context
}

var _ driver.Tx = (*otelTx)(nil)

func newTx(ctx context.Context, tx driver.Tx, c traceConfig) *otelTx {
	return &otelTx{ctx: ctx, tx: tx, config: c}
}

// Commit traces the call to the wrapped Tx.Commit method.
func (t *otelTx) Commit() error {
	return t.config.withSpan(t.ctx, moniker.Commit, func(ctx context.Context) error {
		return t.tx.Commit()
	})
}

// Rollback traces the call to the wrapped Tx.Rollback method.
func (t *otelTx) Rollback() error {
	return t.config.withSpan(t.ctx, moniker.Rollback, func(ctx context.Context) error {
		return t.tx.Rollback()
	})
}
