// Package splunksql provides functions to trace the database/sql package
// (https://golang.org/pkg/database/sql) using the OpenTelemetry API. It will
// automatically augment operations such as connections, statements and
// transactions with tracing.
package splunksql

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

// Register makes a database driver available by the provided name. If
// Register is called twice with the same name or if driver is nil, it panics.
func Register(name string, driver driver.Driver) {
	// Wrap the sql.Register function to perserve the API from
	// signalfx-go-tracing.
	sql.Register(name, driver)
}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a database
// name and connection information. The returned database uses a traced driver
// for all connections. Register must first be called for this to succeed,
// otherwise an error is returned.
func Open(driverName, dataSourceName string, opts ...Option) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	opts = append(opts, withDataSource(driverName, dataSourceName))
	d := newDriver(db.Driver(), newConfig(opts...))

	if driverCtx, ok := d.(driver.DriverContext); ok {
		connector, err := driverCtx.OpenConnector(dataSourceName)
		if err != nil {
			return nil, err
		}
		return sql.OpenDB(connector), nil
	}

	return sql.OpenDB(dsnConnector{dsn: dataSourceName, driver: d}), nil
}

type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}
