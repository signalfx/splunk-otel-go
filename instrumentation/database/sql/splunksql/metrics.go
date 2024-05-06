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

package splunksql // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

import (
	"context"
	"database/sql"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func registerMetrics(db *sql.DB, meter metric.Meter, poolName string) (metric.Registration, error) {
	usage, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.usage",
		metric.WithUnit("{connection}"),
		metric.WithDescription("The number of connections that are currently in state described by the state attribute"),
	)
	if err != nil {
		return nil, err
	}

	maxConn, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.max",
		metric.WithUnit("{connection}"),
		metric.WithDescription("The maximum number of open connections allowed"),
	)
	if err != nil {
		return nil, err
	}

	waitTime, err := meter.Int64ObservableUpDownCounter(
		"db.client.connections.wait_time",
		metric.WithUnit("ms"),
		metric.WithDescription("The time it took to obtain an open connection from the pool"),
	)
	if err != nil {
		return nil, err
	}

	var (
		poolAttr = attribute.String("pool.name", poolName)
		usedAttr = attribute.String("state", "used")
		idleAttr = attribute.String("state", "idle")

		poolSet  = attribute.NewSet(poolAttr)
		inUseSet = attribute.NewSet(poolAttr, usedAttr)
		idleSet  = attribute.NewSet(poolAttr, idleAttr)
	)

	reg, err := meter.RegisterCallback(
		func(_ context.Context, o metric.Observer) error {
			stats := db.Stats()

			opt := metric.WithAttributeSet(inUseSet)
			o.ObserveInt64(usage, int64(stats.InUse), opt)

			opt = metric.WithAttributeSet(idleSet)
			o.ObserveInt64(usage, int64(stats.Idle), opt)

			opt = metric.WithAttributeSet(poolSet)
			o.ObserveInt64(maxConn, int64(stats.MaxOpenConnections), opt)
			o.ObserveInt64(waitTime, int64(stats.WaitDuration), opt)

			return nil
		},
		usage,
		maxConn,
		waitTime,
	)
	if err != nil {
		return nil, err
	}
	return reg, nil
}
