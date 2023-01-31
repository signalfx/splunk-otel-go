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
	"context"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConnector struct {
	driver *mockDriver

	connectN int
	driverN  int
}

var _ driver.Connector = (*mockConnector)(nil)

func newMockConnector(d *mockDriver) *mockConnector {
	return &mockConnector{driver: d}
}

func (c *mockConnector) Connect(context.Context) (driver.Conn, error) {
	c.connectN++
	return c.driver.Open("")
}

func (c *mockConnector) Driver() driver.Driver {
	c.driverN++
	return c.driver
}

type errCloserConnector struct{ driver.Connector }

func (c *errCloserConnector) Close() error { return assert.AnError }

func TestUnderlyingConnectorCloser(t *testing.T) {
	c := newConnector(nil, nil)
	var err error
	assert.NotPanics(t, func() { err = c.Close() })
	assert.NoError(t, err)

	underlying := new(errCloserConnector)
	c = newConnector(underlying, nil)
	assert.ErrorIs(t, c.Close(), assert.AnError)
}
