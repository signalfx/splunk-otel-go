package test

import (
	"database/sql/driver"
	"sync/atomic"
)

type MockDriver struct {
	connector *MockConnector

	OpenN, OpenConnectorN uint64
}

func NewMockDriver() *MockDriver {
	d := &MockDriver{}
	d.connector = NewMockConnector(d)
	d.Reset()
	return d
}

var (
	_ driver.Driver        = (*MockDriver)(nil)
	_ driver.DriverContext = (*MockDriver)(nil)
)

func (d *MockDriver) Open(name string) (driver.Conn, error) {
	atomic.AddUint64(&d.OpenN, 1)
	return d.connector.conn, nil
}

func (d *MockDriver) OpenConnector(name string) (driver.Connector, error) {
	atomic.AddUint64(&d.OpenConnectorN, 1)
	return d.connector, nil
}

func (d *MockDriver) Reset() {
	d.OpenN = 0
	d.OpenConnectorN = 0

	d.connector.Reset()
}
