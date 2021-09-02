package test

import (
	"database/sql/driver"
)

type MockDriver struct {
	newConnFunc func() driver.Conn
}

var (
	_ driver.Driver        = (*MockDriver)(nil)
	_ driver.DriverContext = (*MockDriver)(nil)
)

func NewFullMockDriver() driver.Driver {
	return &MockDriver{newConnFunc: NewFullMockConn}
}

func NewSimpleMockDriver() driver.Driver {
	d := &MockDriver{newConnFunc: NewSimpleMockConn}
	return struct{ driver.Driver }{d}
}

func (d *MockDriver) Open(string) (driver.Conn, error) {
	return d.newConnFunc(), nil
}

func (d *MockDriver) OpenConnector(string) (driver.Connector, error) {
	return NewMockConnector(d), nil
}
