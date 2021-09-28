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
)

type otelConnector struct {
	driver.Connector

	driver *otelDriver
}

var _ driver.Connector = (*otelConnector)(nil)

func newConnector(c driver.Connector, d *otelDriver) *otelConnector {
	return &otelConnector{Connector: c, driver: d}
}

func (c *otelConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return newConn(conn, c.driver.config), nil
}

func (c *otelConnector) Driver() driver.Driver {
	return c.driver
}
