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

type mockConn struct{}

var (
	_ driver.Pinger             = (*mockConn)(nil)
	_ driver.Execer             = (*mockConn)(nil) // nolint: staticcheck
	_ driver.ExecerContext      = (*mockConn)(nil)
	_ driver.Queryer            = (*mockConn)(nil) // nolint: staticcheck
	_ driver.QueryerContext     = (*mockConn)(nil)
	_ driver.Conn               = (*mockConn)(nil)
	_ driver.ConnPrepareContext = (*mockConn)(nil)
	_ driver.ConnBeginTx        = (*mockConn)(nil)
	_ driver.SessionResetter    = (*mockConn)(nil)
)

func newFullMockConn() driver.Conn {
	return &mockConn{}
}

func newSimpleMockConn() driver.Conn {
	return struct{ driver.Conn }{&mockConn{}}
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return newMockStmt(query), nil
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return newMockTx(), nil
}

func (c *mockConn) Ping(ctx context.Context) error {
	return nil
}

func (c *mockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (c *mockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return newMockRows(), nil
}

func (c *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return newMockRows(), nil
}

func (c *mockConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return newMockStmt(query), nil
}

func (c *mockConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return newMockTx(), nil
}

func (c *mockConn) ResetSession(ctx context.Context) error {
	return nil
}
