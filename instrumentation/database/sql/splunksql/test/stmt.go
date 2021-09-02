package test

import (
	"context"
	"database/sql/driver"
)

type MockStmt struct {
	query string
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
	return nil
}

func (s *MockStmt) NumInput() int {
	return 0
}

func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *MockStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return NewMockRows(), nil
}

func (s *MockStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return NewMockRows(), nil
}
