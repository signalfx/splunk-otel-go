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

package test

import (
	"context"
	"database/sql/driver"
)

type mockDriver struct {
	newConnFunc func() driver.Conn
}

var (
	_ driver.Driver        = (*mockDriver)(nil)
	_ driver.DriverContext = (*mockDriver)(nil)
)

func newFullMockDriver() driver.Driver {
	return &mockDriver{newConnFunc: newFullMockConn}
}

func newSimpleMockDriver() driver.Driver {
	d := &mockDriver{newConnFunc: newSimpleMockConn}
	return struct{ driver.Driver }{d}
}

func (d *mockDriver) Open(string) (driver.Conn, error) {
	return d.newConnFunc(), nil
}

func (d *mockDriver) OpenConnector(string) (driver.Connector, error) {
	return newMockConnector(d), nil
}

type mockConnector struct {
	driver *mockDriver
}

var _ driver.Connector = (*mockConnector)(nil)

func newMockConnector(d *mockDriver) driver.Connector {
	return &mockConnector{driver: d}
}

func (c *mockConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.newConnFunc(), nil
}

func (c *mockConnector) Driver() driver.Driver {
	return c.driver
}

type mockConn struct{}

var (
	_ driver.Pinger             = (*mockConn)(nil)
	_ driver.Execer             = (*mockConn)(nil) // nolint: staticcheck // Ensure backwards support of deprecated interface.
	_ driver.ExecerContext      = (*mockConn)(nil)
	_ driver.Queryer            = (*mockConn)(nil) // nolint: staticcheck // Ensure backwards support of deprecated interface.
	_ driver.QueryerContext     = (*mockConn)(nil)
	_ driver.Conn               = (*mockConn)(nil)
	_ driver.ConnPrepareContext = (*mockConn)(nil)
	_ driver.ConnBeginTx        = (*mockConn)(nil)
	_ driver.SessionResetter    = (*mockConn)(nil)
)

func newFullMockConn() driver.Conn { return &mockConn{} }

func newSimpleMockConn() driver.Conn { return struct{ driver.Conn }{&mockConn{}} }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return newMockStmt(query), nil
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return newMockTx(), nil
}

func (c *mockConn) Ping(context.Context) error {
	return nil
}

func (c *mockConn) Exec(string, []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (c *mockConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (c *mockConn) Query(string, []driver.Value) (driver.Rows, error) {
	return newMockRows(), nil
}

func (c *mockConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return newMockRows(), nil
}

func (c *mockConn) PrepareContext(_ context.Context, query string) (driver.Stmt, error) {
	return newMockStmt(query), nil
}

func (c *mockConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return newMockTx(), nil
}

func (c *mockConn) ResetSession(context.Context) error {
	return nil
}

type mockRows struct{}

var _ driver.Rows = (*mockRows)(nil)

func newMockRows() *mockRows {
	return &mockRows{}
}

func (r *mockRows) Columns() []string         { return nil }
func (r *mockRows) Close() error              { return nil }
func (r *mockRows) Next([]driver.Value) error { return nil }

type mockStmt struct {
	query string
}

var (
	_ driver.Stmt             = (*mockStmt)(nil)
	_ driver.StmtExecContext  = (*mockStmt)(nil)
	_ driver.StmtQueryContext = (*mockStmt)(nil)
)

func newMockStmt(query string) *mockStmt {
	return &mockStmt{query: query}
}

func (s *mockStmt) Close() error {
	return nil
}

func (s *mockStmt) NumInput() int {
	return 0
}

func (s *mockStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *mockStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (s *mockStmt) Query([]driver.Value) (driver.Rows, error) {
	return newMockRows(), nil
}

func (s *mockStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	return newMockRows(), nil
}

type mockTx struct{}

var _ driver.Tx = (*mockTx)(nil)

func newMockTx() *mockTx {
	return &mockTx{}
}

func (t *mockTx) Commit() error   { return nil }
func (t *mockTx) Rollback() error { return nil }
