package test

import (
	"database/sql/driver"
)

type MockRows struct{}

var _ driver.Rows = (*MockRows)(nil)

func NewMockRows() *MockRows {
	return &MockRows{}
}

func (r *MockRows) Columns() []string              { return nil }
func (r *MockRows) Close() error                   { return nil }
func (r *MockRows) Next(dest []driver.Value) error { return nil }
