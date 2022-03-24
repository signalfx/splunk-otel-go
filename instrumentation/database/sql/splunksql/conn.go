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
	"errors"

	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

type otelConn struct {
	driver.Conn

	config traceConfig
}

// Compile-time check otelConn implements database interfaces.
var (
	_ driver.Pinger             = (*otelConn)(nil)
	_ driver.Execer             = (*otelConn)(nil) // nolint:staticcheck // Ensure backwards support of deprecated interface.
	_ driver.ExecerContext      = (*otelConn)(nil)
	_ driver.Queryer            = (*otelConn)(nil) // nolint:staticcheck // Ensure backwards support of deprecated interface.
	_ driver.QueryerContext     = (*otelConn)(nil)
	_ driver.Conn               = (*otelConn)(nil)
	_ driver.ConnPrepareContext = (*otelConn)(nil)
	_ driver.ConnBeginTx        = (*otelConn)(nil)
	_ driver.SessionResetter    = (*otelConn)(nil)
)

func newConn(conn driver.Conn, conf traceConfig) *otelConn {
	return &otelConn{Conn: conn, config: conf}
}

// Ping traces a ping to the connected database.
func (c *otelConn) Ping(ctx context.Context) error {
	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}
	return c.config.withSpan(ctx, moniker.Ping, pinger.Ping)
}

// Exec calls the wrapped Connection Exec method if implemented.
func (c *otelConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.Conn.(driver.Execer); ok { // nolint:staticcheck // Ensure backwards support of deprecated interface.
		return execer.Exec(query, args)
	}
	return nil, driver.ErrSkip
}

// ExecContext traces the call to the wrapped Connection ExecContext method.
// If the wrapped driver does not implement this method it will fallback to
// wrapping a call to Exec.
func (c *otelConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var (
		f   func(context.Context) error
		res driver.Result
	)
	if execer, ok := c.Conn.(driver.ExecerContext); ok {
		f = func(ctx context.Context) error {
			var err error
			res, err = execer.ExecContext(ctx, query, args)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Exec.
		vArgs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}
		f = func(ctx context.Context) error {
			var err error
			res, err = c.Exec(query, vArgs)
			return err
		}
	}

	err := c.config.withSpan(ctx, moniker.Exec, f, trace.WithAttributes(semconv.DBStatementKey.String(query)))
	return res, err
}

// Query calls the wrapped Connection Query method if implemented.
func (c *otelConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.Conn.(driver.Queryer); ok { // nolint:staticcheck // Ensure backwards support of deprecated interface.
		return queryer.Query(query, args)
	}
	return nil, driver.ErrSkip
}

// QueryContext traces the call to the wrapped Connection QueryContext method.
// If the wrapped driver does not implement this method it will fallback to
// wrapping a call to Query.
func (c *otelConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var (
		f    func(context.Context) error
		rows driver.Rows
	)
	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		f = func(ctx context.Context) error {
			var err error
			rows, err = queryer.QueryContext(ctx, query, args)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Query.
		vArgs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}
		f = func(ctx context.Context) error {
			var err error
			rows, err = c.Query(query, vArgs)
			return err
		}
	}

	err := c.config.withSpan(ctx, moniker.Query, f, trace.WithAttributes(semconv.DBStatementKey.String(query)))
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// PrepareContext returns a prepared statement, bound to this traced connection.
func (c *otelConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		f    func(context.Context) error
		stmt driver.Stmt
	)
	if preparer, ok := c.Conn.(driver.ConnPrepareContext); ok {
		f = func(ctx context.Context) error {
			var err error
			stmt, err = preparer.PrepareContext(ctx, query)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Prepare.
		f = func(ctx context.Context) error {
			var err error
			stmt, err = c.Conn.Prepare(query)
			return err
		}
	}

	err := c.config.withSpan(ctx, moniker.Prepare, f, trace.WithAttributes(semconv.DBStatementKey.String(query)))
	if err != nil {
		return nil, err
	}
	return newStmt(stmt, c.config, query), nil
}

// BeginTx starts and returns a new traced transaction.
func (c *otelConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	var (
		f  func(context.Context) error
		tx driver.Tx
	)
	if transactor, ok := c.Conn.(driver.ConnBeginTx); ok {
		f = func(ctx context.Context) error {
			var err error
			tx, err = transactor.BeginTx(ctx, opts)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Begin.
		f = func(ctx context.Context) error {
			var err error
			tx, err = c.Conn.Begin() // nolint:staticcheck // Ensure backwards support of deprecated interface.
			return err
		}
	}

	err := c.config.withSpan(ctx, moniker.Begin, f)
	if err != nil {
		return nil, err
	}
	return newTx(ctx, tx, c.config), nil
}

// ResetSession traces the call to the wrapped Connection ResetSession method.
func (c *otelConn) ResetSession(ctx context.Context) error {
	resetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}

	return c.config.withSpan(ctx, moniker.Reset, resetter.ResetSession)
}

// copied from stdlib database/sql package: src/database/sql/ctxutil.go
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	vArgs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("splunksql: driver does not support the use of Named Parameters")
		}
		vArgs[n] = param.Value
	}
	return vArgs, nil
}
