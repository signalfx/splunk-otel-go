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

	"github.com/stretchr/testify/require"
)

type namedFallbackDriver struct{ conn driver.Conn }

func (d *namedFallbackDriver) Open(string) (driver.Conn, error) { return d.conn, nil }

type namedFallbackConnector struct{ driver driver.Driver }

func (c *namedFallbackConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open("")
}

func (c *namedFallbackConnector) Driver() driver.Driver { return c.driver }

type namedFallbackConn struct {
	driver.Conn
	stmt     *namedFallbackStmt
	prepares int
}

func (c *namedFallbackConn) Prepare(string) (driver.Stmt, error) {
	c.prepares++
	return c.stmt, nil
}

type namedFallbackStmt struct {
	driver.Stmt
	args    []driver.NamedValue
	execs   int
	queries int
}

func (*namedFallbackStmt) NumInput() int { return 1 }

func (s *namedFallbackStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	s.args = args
	s.execs++
	return driver.RowsAffected(1), nil
}

func (s *namedFallbackStmt) QueryContext(_ context.Context, args []driver.NamedValue) (driver.Rows, error) {
	s.args = args
	s.queries++
	return namedFallbackRows{}, nil
}

type namedFallbackRows struct{}

func (namedFallbackRows) Columns() []string         { return []string{"value"} }
func (namedFallbackRows) Close() error              { return nil }
func (namedFallbackRows) Next([]driver.Value) error { return io.EOF }

func TestNamedArgumentsReachPreparedFallback(t *testing.T) {
	for _, operation := range []string{"exec", "query"} {
		t.Run(operation, func(t *testing.T) {
			stmt := &namedFallbackStmt{Stmt: newMockStmt(nil)}
			conn := &namedFallbackConn{Conn: newMockConn(nil), stmt: stmt}
			db := sql.OpenDB(&namedFallbackConnector{driver: newDriver(&namedFallbackDriver{conn: conn}, newConfig())})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			if operation == "exec" {
				_, err := db.ExecContext(context.Background(), "statement", sql.Named("arg", 1))
				require.NoError(t, err)
				require.Equal(t, 1, stmt.execs)
			} else {
				rows, err := db.QueryContext(context.Background(), "statement", sql.Named("arg", 1))
				require.NoError(t, err)
				defer func() { require.NoError(t, rows.Close()) }()
				require.NoError(t, rows.Err())
				require.Equal(t, 1, stmt.queries)
			}
			require.Equal(t, 1, conn.prepares)
			require.Equal(t, []driver.NamedValue{{Name: "arg", Ordinal: 1, Value: int64(1)}}, stmt.args)
		})
	}
}

func TestLegacyFastPathsStillRejectNamedArguments(t *testing.T) {
	conn := newMockConn(nil)
	args := []driver.NamedValue{{Name: "arg", Ordinal: 1, Value: int64(1)}}
	execer := newConn(struct{ connExecer }{conn}, newConfig())
	_, err := execer.ExecContext(context.Background(), "statement", args)
	require.ErrorContains(t, err, "Named Parameters")
	require.Zero(t, conn.execN)
	queryer := newConn(struct{ connQuery }{conn}, newConfig())
	rows, err := queryer.QueryContext(context.Background(), "statement", args)
	require.Nil(t, rows)
	require.ErrorContains(t, err, "Named Parameters")
	require.Zero(t, conn.queryN)
}
