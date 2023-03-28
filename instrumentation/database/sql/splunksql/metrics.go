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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

func registerMetrics(meter metric.Meter, db *sql.DB) (metric.Registration, error) {
	usage, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.usage",
		instrument.WithUnit("{connection}"),
		instrument.WithDescription("The number of connections that are currently in state described by the state attribute"),
	)
	if err != nil {
		return nil, err
	}

	idleMax, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.idle.max",
		instrument.WithUnit("{connection}"),
		instrument.WithDescription("The maximum number of idle open connections allowed"),
	)
	if err != nil {
		return nil, err
	}

	reg, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			stats := db.Stats()
			poolAttr := attribute.String("pool.name", "bad") // TODO: add tests and proper implementation

			o.ObserveInt64(usage, int64(stats.InUse), poolAttr, attribute.String("state", "used"))
			o.ObserveInt64(usage, int64(stats.Idle), poolAttr, attribute.String("state", "idle"))

			return nil
		},
		// usage,
		idleMax,
		// Passing only a not-used metric and it works (sic!).
		// We just need to provide any metric to make the RegisterCallback working...
		// Probably the SDK should report some error like:
		// "db.client.connections.usage was not registered, but used"
	)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

type unregisterConnector struct {
	driver.Connector

	reg metric.Registration
}

func newUnregisterConnector(c driver.Connector, reg metric.Registration) driver.Connector {
	return unregisterConnector{Connector: c, reg: reg}
}

func (c unregisterConnector) Close() error {
	if err := c.reg.Unregister(); err != nil {
		otel.Handle(err)
	}

	if closer, ok := c.Connector.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
