// Package splunksql provides functions to trace the database/sql package
// (https://golang.org/pkg/database/sql) using the OpenTelemetry API. It will
// automatically augment operations such as connections, statements and
// transactions with tracing.
package splunksql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

// Register makes a traced database driver available by the provided name. It
// must be called before Open, if that connection is to be traced. If Register
// is called twice with the same name or if driver is nil, it panics.
func Register(name string, driver driver.Driver, opts ...Option) {
	if driver == nil {
		panic("splunksql: Register driver is nil")
	}

	// Use a suffixed version of the passed name. This is to help distinguish
	// instrumented vs non-instrumented registrations of the same driver.
	name = tracedDriverName(name)
	// sql.Register will panic if called twice with the same driver. Preserve
	// this behavior as users should not be confused if their code mistakenly
	// tries to register a driver with different configuration about which
	// registration is used.
	sql.Register(name, newDriver(driver, newConfig(opts...)))
}

// errNotRegistered is returned when there is an attempt to open a database
// with a driver that has not previously been registered using this package.
var errNotRegistered = errors.New("splunksql: Register must be called before Open")

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a database
// name and connection information. The returned database uses a traced driver
// for all connections. Register must first be called for this to succeed,
// otherwise an error is returned.
//
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the Open
// function should be called just once. It is rarely necessary to
// close a DB.
func Open(driverName, dataSourceName string) (*sql.DB, error) {
	name := tracedDriverName(driverName)
	if !registered(name) {
		return nil, errNotRegistered
	}
	return sql.Open(name, dataSourceName)
}

// tracedDriverName returns the name of the traced version of a driver name.
func tracedDriverName(name string) string { return name + ".traced" }

// registered returns if the driver with name has already been registered.
func registered(name string) bool {
	for _, v := range sql.Drivers() {
		if name == v {
			return true
		}
	}
	return false
}
