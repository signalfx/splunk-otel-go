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
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyCancellationConn deliberately only exposes the legacy driver methods.
type legacyCancellationConn struct {
	stmt          *mockStmt
	tx            *mockTx
	cancel        context.CancelFunc
	err           error
	execN, queryN int
}

func (c *legacyCancellationConn) Prepare(string) (driver.Stmt, error) {
	if c.cancel != nil {
		c.cancel()
	}
	return c.stmt, c.err
}

func (c *legacyCancellationConn) Begin() (driver.Tx, error) {
	if c.cancel != nil {
		c.cancel()
	}
	return c.tx, c.err
}

func (*legacyCancellationConn) Close() error { return nil }

func (c *legacyCancellationConn) Exec(string, []driver.Value) (driver.Result, error) {
	c.execN++
	return driver.RowsAffected(1), c.err
}

func (c *legacyCancellationConn) Query(string, []driver.Value) (driver.Rows, error) {
	c.queryN++
	return cancellationRows{}, c.err
}

type cancellationRows struct{}

func (cancellationRows) Columns() []string         { return nil }
func (cancellationRows) Close() error              { return nil }
func (cancellationRows) Next([]driver.Value) error { return io.EOF }

type cancellationConnector struct {
	conn driver.Conn
}

func (c cancellationConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (c cancellationConnector) Open(string) (driver.Conn, error)             { return c.conn, nil }
func (c cancellationConnector) Driver() driver.Driver                        { return c }

func TestCanceledLegacyReservedConn(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		for _, query := range []bool{false, true} {
			name := "exec"
			if query {
				name = "query"
			}
			if wrapped {
				name += "/instrumented"
			}
			t.Run(name, func(t *testing.T) {
				legacy := &legacyCancellationConn{}
				var dc driver.Conn = legacy
				if wrapped {
					dc = newConn(dc, newConfig())
				}
				db := sql.OpenDB(cancellationConnector{conn: dc})
				t.Cleanup(func() { assert.NoError(t, db.Close()) })
				conn, err := db.Conn(context.Background())
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, conn.Close()) })
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				if query {
					rows, queryErr := conn.QueryContext(ctx, "select 1")
					if rows != nil {
						t.Cleanup(func() { assert.NoError(t, rows.Close()) })
						assert.NoError(t, rows.Err())
					}
					assert.ErrorIs(t, queryErr, context.Canceled)
				} else {
					_, err = conn.ExecContext(ctx, "update t set v = 1")
					assert.ErrorIs(t, err, context.Canceled)
				}
				assert.Zero(t, legacy.execN)
				assert.Zero(t, legacy.queryN)
			})
		}
	}
}

func TestCanceledLegacyStatements(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, query := range []bool{false, true} {
		t.Run(map[bool]string{false: "exec", true: "query"}[query], func(t *testing.T) {
			legacy := newMockStmt(nil)
			stmt := newStmt(struct{ driver.Stmt }{legacy}, newConfig(), "select 1")
			if query {
				rows, err := stmt.QueryContext(ctx, nil)
				assert.Nil(t, rows)
				assert.ErrorIs(t, err, context.Canceled)
			} else {
				result, err := stmt.ExecContext(ctx, nil)
				assert.Nil(t, result)
				assert.ErrorIs(t, err, context.Canceled)
			}
			assert.Zero(t, legacy.execN)
			assert.Zero(t, legacy.queryN)
		})
	}
}

func TestLegacyCancellationKeepsNamedParameterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	legacyConn := &legacyCancellationConn{}
	conn := newConn(legacyConn, newConfig())
	legacyStmt := newMockStmt(nil)
	stmt := newStmt(struct{ driver.Stmt }{legacyStmt}, newConfig(), "select 1")
	args := []driver.NamedValue{{Name: "value", Value: 1}}
	operations := map[string]func() error{
		"conn exec":  func() error { _, err := conn.ExecContext(ctx, "", args); return err },
		"conn query": func() error { _, err := conn.QueryContext(ctx, "", args); return err }, //nolint:gocritic // Error-only driver calls cannot return live rows in these tests.
		"stmt exec":  func() error { _, err := stmt.ExecContext(ctx, args); return err },
		"stmt query": func() error { _, err := stmt.QueryContext(ctx, args); return err }, //nolint:gocritic // Error-only driver calls cannot return live rows in these tests.
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			assert.EqualError(t, operation(), "splunksql: driver does not support the use of Named Parameters")
		})
	}
	assert.Zero(t, legacyConn.execN)
	assert.Zero(t, legacyConn.queryN)
	assert.Zero(t, legacyStmt.execN)
	assert.Zero(t, legacyStmt.queryN)
}

func TestLegacyCancellationCleansUpPrepareAndBegin(t *testing.T) {
	for _, driverErr := range []error{nil, errTest} {
		t.Run(map[bool]string{false: "success", true: "driver error"}[driverErr != nil], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Cleanup failures must not replace the cancellation error.
			legacy := &legacyCancellationConn{stmt: newMockStmt(errTest), tx: newMockTx(errTest), cancel: cancel, err: driverErr}
			conn := newConn(legacy, newConfig())
			stmt, err := conn.PrepareContext(ctx, "select 1")
			require.Nil(t, stmt)
			if driverErr != nil {
				assert.ErrorIs(t, err, driverErr)
				assert.Zero(t, legacy.stmt.closeN)
			} else {
				assert.ErrorIs(t, err, context.Canceled)
				assert.Equal(t, 1, legacy.stmt.closeN)
			}

			ctx, cancel = context.WithCancel(context.Background())
			defer cancel()
			legacy.cancel = cancel
			tx, err := conn.BeginTx(ctx, driver.TxOptions{})
			require.Nil(t, tx)
			if driverErr != nil {
				assert.ErrorIs(t, err, driverErr)
				assert.Zero(t, legacy.tx.rollbackN)
			} else {
				assert.ErrorIs(t, err, context.Canceled)
				assert.Equal(t, 1, legacy.tx.rollbackN)
			}
		})
	}
}

func TestCanceledContextDelegatesToContextDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	legacy := newMockConn(errTest)
	conn := newConn(legacy, newConfig())
	stmtDriver := newMockStmt(errTest)
	stmt := newStmt(stmtDriver, newConfig(), "select 1")
	operations := map[string]func() error{
		"conn exec":  func() error { _, err := conn.ExecContext(ctx, "", nil); return err },
		"conn query": func() error { _, err := conn.QueryContext(ctx, "", nil); return err }, //nolint:gocritic // Error-only driver calls cannot return live rows in these tests.
		"prepare":    func() error { _, err := conn.PrepareContext(ctx, ""); return err },
		"begin":      func() error { _, err := conn.BeginTx(ctx, driver.TxOptions{}); return err },
		"stmt exec":  func() error { _, err := stmt.ExecContext(ctx, nil); return err },
		"stmt query": func() error { _, err := stmt.QueryContext(ctx, nil); return err }, //nolint:gocritic // Error-only driver calls cannot return live rows in these tests.
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) { assert.ErrorIs(t, operation(), errTest) })
	}
	assert.Equal(t, 1, legacy.execContextN)
	assert.Equal(t, 1, legacy.queryContextN)
	assert.Equal(t, 1, legacy.prepareContextN)
	assert.Equal(t, 1, legacy.beginTxN)
	assert.Equal(t, 1, stmtDriver.execContextN)
	assert.Equal(t, 1, stmtDriver.queryContextN)
	assert.Zero(t, legacy.execN)
	assert.Zero(t, legacy.queryN)
	assert.Zero(t, legacy.prepareN)
	assert.Zero(t, legacy.beginN)
	assert.Zero(t, stmtDriver.execN)
	assert.Zero(t, stmtDriver.queryN)
}
