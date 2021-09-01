package test

import (
	"context"
	"database/sql/driver"
	"sync/atomic"
)

type MockConnector struct {
	driver *MockDriver
	conn   *MockConn

	ConnectN uint64
	DriverN  uint64
}

func NewMockConnector(d *MockDriver) *MockConnector {
	conn := NewMockConn()
	return &MockConnector{driver: d, conn: conn}
}

func (c *MockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	atomic.AddUint64(&c.ConnectN, 1)
	return c.conn, nil
}

func (c *MockConnector) Driver() driver.Driver {
	atomic.AddUint64(&c.DriverN, 1)
	return c.driver
}

func (c *MockConnector) Reset() {
	c.ConnectN = 0
	c.DriverN = 0

	c.conn.Reset()
}
