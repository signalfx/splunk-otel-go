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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

func registerMetrics(db *sql.DB, meter metric.Meter, poolName string) (metric.Registration, error) {
	usage, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.usage",
		instrument.WithUnit("{connection}"),
		instrument.WithDescription("The number of connections that are currently in state described by the state attribute"),
	)
	if err != nil {
		return nil, err
	}

	max, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.max",
		instrument.WithUnit("{connection}"),
		instrument.WithDescription("The maximum number of open connections allowed"),
	)
	if err != nil {
		return nil, err
	}

	waitTime, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.wait_time",
		instrument.WithUnit("ms"),
		instrument.WithDescription("The time it took to obtain an open connection from the pool"),
	)
	if err != nil {
		return nil, err
	}

	reg, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			poolAttr := attribute.String("pool.name", poolName)
			usedAttr := attribute.String("state", "used")
			idleAttr := attribute.String("state", "idle")

			stats := db.Stats()

			o.ObserveInt64(usage, int64(stats.InUse), poolAttr, usedAttr)
			o.ObserveInt64(usage, int64(stats.Idle), poolAttr, idleAttr)
			o.ObserveInt64(max, int64(stats.MaxOpenConnections), poolAttr)
			o.ObserveInt64(waitTime, int64(stats.WaitDuration), poolAttr)

			return nil
		},
		usage,
		max,
		waitTime,
	)
	if err != nil {
		return nil, err
	}
	return reg, nil
}
