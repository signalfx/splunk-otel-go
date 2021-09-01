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
