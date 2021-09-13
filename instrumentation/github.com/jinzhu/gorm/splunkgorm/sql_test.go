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

package splunkgorm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/stretchr/testify/assert"
)

type mockOpen struct {
	driverName, dataSourceName string
	options                    []splunksql.Option
}

func (m *mockOpen) Open(name, dsn string, opts ...splunksql.Option) (*sql.DB, error) {
	m.driverName = name
	m.dataSourceName = dsn
	m.options = opts
	// Return something capable of implementing Ping so gorm does not panic.
	return sql.OpenDB(mockConnector{}), nil
}

type mockConnector struct {
	driver.Connector
}

func (mockConnector) Connect(context.Context) (driver.Conn, error) {
	return mockConn{}, nil
}

type mockConn struct {
	driver.Conn
}

func (mockConn) Ping(context.Context) error { return nil }

func TestOpen(t *testing.T) {
	origOpen := openFunc
	m := new(mockOpen)
	openFunc = m.Open
	defer func() { openFunc = origOpen }()

	name, dsn := "testDB", "fake://user:pass@localhost/DB"
	options := []splunksql.Option{splunksql.WithAttributes(nil)}
	db, err := Open(name, dsn, options...)
	assert.NoError(t, err)
	assert.IsType(t, &gorm.DB{}, db)
	assert.Equal(t, name, m.driverName)
	assert.Equal(t, dsn, m.dataSourceName)
	assert.Equal(t, options, m.options)
}
