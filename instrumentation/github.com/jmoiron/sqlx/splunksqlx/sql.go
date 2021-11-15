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

// Package splunksqlx provides instrumentation for the github.com/jmoiron/sqlx
// package.
package splunksqlx // import "github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx"

import (
	"github.com/jmoiron/sqlx"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

// openFunc allows overrides for testing.
var openFunc = splunksql.Open

// Open opens a new (traced) connection to the database using the given driver
// and source. The driver must already be registered by the driver package.
func Open(driverName, dataSourceName string, opts ...splunksql.Option) (*sqlx.DB, error) {
	db, err := openFunc(driverName, dataSourceName, opts...)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, driverName), nil
}

// MustOpen is the same as Open, but panics on error.
func MustOpen(driverName, dataSourceName string, opts ...splunksql.Option) *sqlx.DB {
	db, err := Open(driverName, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

// Connect connects to a database with a traced connection and verifies with a
// ping. The driver used to connect must already be registered by the driver
// package.
func Connect(driverName, dataSourceName string, opts ...splunksql.Option) (*sqlx.DB, error) {
	db, err := Open(driverName, dataSourceName, opts...)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// MustConnect is the same as Connect, but panics on error.
func MustConnect(driverName, dataSourceName string, opts ...splunksql.Option) *sqlx.DB {
	db, err := Connect(driverName, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
