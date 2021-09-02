package test

import (
	"context"
	"database/sql/driver"
)

type MockConnector struct {
	driver *MockDriver
}

var _ driver.Connector = (*MockConnector)(nil)

func NewMockConnector(d *MockDriver) driver.Connector {
	return &MockConnector{driver: d}
}

func (c *MockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.newConnFunc(), nil
}

func (c *MockConnector) Driver() driver.Driver {
	return c.driver
}
