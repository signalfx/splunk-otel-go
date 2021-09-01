package test

import (
	"database/sql/driver"
	"sync/atomic"
)

type MockTx struct {
	CommitN, RollbackN uint64
}

var _ driver.Tx = (*MockTx)(nil)

func NewMockTx() *MockTx {
	return &MockTx{}
}

func (t *MockTx) Commit() error {
	atomic.AddUint64(&t.CommitN, 1)
	return nil
}

func (t *MockTx) Rollback() error {
	atomic.AddUint64(&t.RollbackN, 1)
	return nil
}
