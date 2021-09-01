package test

import (
	"database/sql/driver"
	"sync/atomic"
)

type MockRows struct {
	ColumnsN, CloseN, NextN uint64
}

var _ driver.Rows = (*MockRows)(nil)

func NewMockRows() *MockRows {
	return &MockRows{}
}

func (r *MockRows) Columns() []string {
	atomic.AddUint64(&r.ColumnsN, 1)
	return nil
}

func (r *MockRows) Close() error {
	atomic.AddUint64(&r.CloseN, 1)
	return nil
}

func (r *MockRows) Next(dest []driver.Value) error {
	atomic.AddUint64(&r.NextN, 1)
	return nil
}
