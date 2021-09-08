package splunksql

import (
	"database/sql/driver"
)

// otelDriver wraps a SQL Driver and traces all operations it performs.
type otelDriver struct {
	driver driver.Driver
	config traceConfig
}

// Compile-time check *otelDriver implements database interfaces.
var (
	_ driver.Driver        = (*otelDriver)(nil)
	_ driver.DriverContext = (*otelDriver)(nil)
)

func newDriver(d driver.Driver, c traceConfig) driver.Driver {
	if _, ok := d.(driver.DriverContext); ok {
		return &otelDriver{driver: d, config: c}
	}
	// Remove the implementation of the driver.DriverContext and only
	// implement the driver.Driver.
	return struct{ driver.Driver }{&otelDriver{driver: d, config: c}}
}

// Open returns a new traced connection to the database. The name is a string
// in a driver-specific format.
func (d *otelDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return newConn(conn, d.config), nil
}

// OpenConnector returns a new traced connector to the database. The name is a
// string in a driver-specific format.
func (d *otelDriver) OpenConnector(name string) (driver.Connector, error) {
	// This should not panic given the guard in newDriver.
	connector, err := d.driver.(driver.DriverContext).OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return newConnector(connector, d), nil
}
