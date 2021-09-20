// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
