package test

import (
	"context"
	"database/sql/driver"
)

type MockConn struct{}

var (
	_ driver.Pinger             = (*MockConn)(nil)
	_ driver.Execer             = (*MockConn)(nil)
	_ driver.ExecerContext      = (*MockConn)(nil)
	_ driver.Queryer            = (*MockConn)(nil)
	_ driver.QueryerContext     = (*MockConn)(nil)
	_ driver.Conn               = (*MockConn)(nil)
	_ driver.ConnPrepareContext = (*MockConn)(nil)
	_ driver.ConnBeginTx        = (*MockConn)(nil)
	_ driver.SessionResetter    = (*MockConn)(nil)
)

func NewFullMockConn() driver.Conn {
	return &MockConn{}
}

func NewSimpleMockConn() driver.Conn {
	return struct{ driver.Conn }{&MockConn{}}
}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return NewMockStmt(query), nil
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) Begin() (driver.Tx, error) {
	return NewMockTx(), nil
}

func (c *MockConn) Ping(ctx context.Context) error {
	return nil
}

func (c *MockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (c *MockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (c *MockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return NewMockRows(), nil
}

func (c *MockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return NewMockRows(), nil
}

func (c *MockConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return NewMockStmt(query), nil
}

func (c *MockConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return NewMockTx(), nil
}

func (c *MockConn) ResetSession(ctx context.Context) error {
	return nil
}
