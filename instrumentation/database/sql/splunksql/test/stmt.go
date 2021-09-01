package test

import (
	"context"
	"database/sql/driver"
	"sync/atomic"
)

type MockStmt struct {
	query string

	CloseN        uint64
	NumInputN     uint64
	ExecN         uint64
	ExecContextN  uint64
	QueryN        uint64
	QueryContextN uint64
}

var (
	_ driver.Stmt             = (*MockStmt)(nil)
	_ driver.StmtExecContext  = (*MockStmt)(nil)
	_ driver.StmtQueryContext = (*MockStmt)(nil)
)

func NewMockStmt(query string) *MockStmt {
	return &MockStmt{query: query}
}

func (s *MockStmt) Close() error {
	atomic.AddUint64(&s.CloseN, 1)
	return nil
}

func (s *MockStmt) NumInput() int {
	atomic.AddUint64(&s.NumInputN, 1)
	return 0
}

func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	atomic.AddUint64(&s.ExecN, 1)
	return nil, nil
}

func (s *MockStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	atomic.AddUint64(&s.ExecContextN, 1)
	return nil, nil
}

func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	atomic.AddUint64(&s.QueryN, 1)
	return NewMockRows(), nil
}

func (s *MockStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	atomic.AddUint64(&s.QueryContextN, 1)
	return NewMockRows(), nil
}
