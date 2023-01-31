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

// Package splunksql provides functions to trace the database/sql package
// (https://golang.org/pkg/database/sql) using the OpenTelemetry API. It will
// automatically augment operations such as connections, statements and
// transactions with tracing.
package splunksql // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]InstrumentationConfig)
)

// Register makes the InstrumentationConfig for the setup of a database
// driverwith the provided name available. If Register is called twice for the
// same name it panics.
func Register(name string, c InstrumentationConfig) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[name]; dup {
		panic("splunksql: Register called twice for " + name)
	}
	registry[name] = c
}

func retrieve(name string) InstrumentationConfig {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a database
// name and connection information. The returned database uses a traced driver
// for all connections.
func Open(driverName, dataSourceName string, opts ...Option) (*sql.DB, error) {
	// The instrumented driver needs to already have been registered with the
	// database/sql package. This is something instrumentation libraries can
	// do by importing the package containing the driver (if it correctly
	// initializes with the registration of its driver).
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	// Do not defer a call to db.Close. The underlying Connector used below
	// needs to remains open. Closing db may also close the Driver.

	// Setup any instrumentation that is registered for this driverName. If no
	// instrumentation was registered for the driver, this will give a "best
	// effort" to setup the driver.
	regOpt := withRegistrationConfig(retrieve(driverName), dataSourceName)
	// Allow users to override default instrumentation setup values by
	// appending opts to the Option returned from withRegistrationConfig. This
	// will allow similar instrumentation to be used with minor corrections
	// being applied here (e.g. using `lib/pg` instead of `pgx`).
	opts = append([]Option{regOpt}, opts...)
	d := newDriver(db.Driver(), newTraceConfig(opts...))

	var connector driver.Connector
	// Use the instrumented driver to open a connection to the database.
	if driverCtx, ok := d.(driver.DriverContext); ok {
		connector, err = driverCtx.OpenConnector(dataSourceName)
		if err != nil {
			return nil, err
		}
	} else {
		connector = newDSNConnector(d, dataSourceName)
	}
	return sql.OpenDB(newCloserConnector(connector, db)), nil
}

// dsnConnector wraps a driver to be used as a DriverContext.
type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func newDSNConnector(d driver.Driver, dsn string) driver.Connector {
	return dsnConnector{dsn: dsn, driver: d}
}

func (t dsnConnector) Connect(context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}

type closerConnector struct {
	driver.Connector

	initDB *sql.DB
}

func newCloserConnector(c driver.Connector, initDB *sql.DB) driver.Connector {
	return closerConnector{Connector: c, initDB: initDB}
}

func (c closerConnector) Close() error {
	if err := c.initDB.Close(); err != nil {
		return err
	}
	if closer, ok := c.Connector.(io.Closer); ok {
		// This may call the same underlying Connector's close method twice,
		// the first call coming from the initDB.Close. However, if this isn't
		// called here and the underlying Connector's close method is different
		// than the one from initDB we would leave things in an open state.
		// Better to ensure the operation is performed and assume the
		// underlying Close is idempotent.
		return closer.Close()
	}
	return nil
}
