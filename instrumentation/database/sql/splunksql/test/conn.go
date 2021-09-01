package test

import (
	"context"
	"database/sql/driver"
	"sync/atomic"
)

type MockConn struct {
	PrepareN        uint64
	CloseN          uint64
	BeginN          uint64
	PingN           uint64
	ExecN           uint64
	ExecContextN    uint64
	QueryN          uint64
	QueryContextN   uint64
	PrepareContextN uint64
	BeginTxN        uint64
	ResetSessionN   uint64
}

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

func NewMockConn() *MockConn {
	return &MockConn{}
}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	atomic.AddUint64(&c.PrepareN, 1)
	return NewMockStmt(query), nil
}

func (c *MockConn) Close() error {
	atomic.AddUint64(&c.CloseN, 1)
	return nil
}

func (c *MockConn) Begin() (driver.Tx, error) {
	atomic.AddUint64(&c.BeginN, 1)
	return NewMockTx(), nil
}

func (c *MockConn) Ping(ctx context.Context) error {
	atomic.AddUint64(&c.PingN, 1)
	return nil
}

func (c *MockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	atomic.AddUint64(&c.ExecN, 1)
	return nil, nil
}

func (c *MockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	atomic.AddUint64(&c.ExecContextN, 1)
	return nil, nil
}

func (c *MockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	atomic.AddUint64(&c.QueryN, 1)
	return NewMockRows(), nil
}

func (c *MockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	atomic.AddUint64(&c.QueryContextN, 1)
	return NewMockRows(), nil
}

func (c *MockConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	atomic.AddUint64(&c.PrepareContextN, 1)
	return NewMockStmt(query), nil
}

func (c *MockConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	atomic.AddUint64(&c.BeginTxN, 1)
	return NewMockTx(), nil
}

func (c *MockConn) ResetSession(ctx context.Context) error {
	atomic.AddUint64(&c.ResetSessionN, 1)
	return nil
}

func (c *MockConn) Reset() {
	c.PrepareN = 0
	c.CloseN = 0
	c.BeginN = 0
	c.PingN = 0
	c.ExecN = 0
	c.ExecContextN = 0
	c.QueryN = 0
	c.QueryContextN = 0
	c.PrepareContextN = 0
	c.BeginTxN = 0
	c.ResetSessionN = 0
}
