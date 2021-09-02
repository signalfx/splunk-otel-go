package test

import (
	"database/sql/driver"
)

type MockTx struct{}

var _ driver.Tx = (*MockTx)(nil)

func NewMockTx() *MockTx {
	return &MockTx{}
}

func (t *MockTx) Commit() error   { return nil }
func (t *MockTx) Rollback() error { return nil }
