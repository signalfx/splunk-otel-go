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

package sqltestutil // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/sqltestutil"

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	splunkotel "github.com/signalfx/splunk-otel-go"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

// TestMetrics is meant for testing metrics instrumentation for instrumentation libraries
// built on top of splunksql.
func TestMetrics(t *testing.T, wantPoolName, driverName, dataSourceName string, execFn func(*sql.DB)) {
	ctx := context.Background()
	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))
	defer func() { assert.NoError(t, meterProvider.Shutdown(ctx)) }()

	// create 1 used connection
	db, err := splunksql.Open(driverName, dataSourceName, splunksql.WithMeterProvider(meterProvider))
	require.NoError(t, err)
	defer func() { assert.NoError(t, db.Close()) }()
	execFn(db)

	// assert
	wantPoolAttr := attribute.String("pool.name", wantPoolName)
	want := metricdata.ScopeMetrics{
		Scope: instrumentation.Scope{
			Name:      "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql",
			Version:   splunkotel.Version(),
			SchemaURL: semconv.SchemaURL,
		},
		Metrics: []metricdata.Metrics{
			{
				Name:        "db.client.connections.usage",
				Unit:        "{connection}",
				Description: "The number of connections that are currently in state described by the state attribute",
				Data: metricdata.Sum[int64]{
					DataPoints: []metricdata.DataPoint[int64]{
						{
							Attributes: attribute.NewSet(
								wantPoolAttr,
								attribute.String("state", "used"),
							),
							Value: 0,
						},
						{
							Attributes: attribute.NewSet(
								wantPoolAttr,
								attribute.String("state", "idle"),
							),
							Value: 1,
						},
					},
					Temporality: metricdata.CumulativeTemporality,
					IsMonotonic: false,
				},
			},
			{
				Name:        "db.client.connections.max",
				Unit:        "{connection}",
				Description: "The maximum number of open connections allowed",
				Data: metricdata.Sum[int64]{
					DataPoints: []metricdata.DataPoint[int64]{
						{
							Attributes: attribute.NewSet(wantPoolAttr),
							Value:      0,
						},
					},
					Temporality: metricdata.CumulativeTemporality,
					IsMonotonic: false,
				},
			},
			{
				Name:        "db.client.connections.wait_time",
				Unit:        "ms",
				Description: "The time it took to obtain an open connection from the pool",
				Data: metricdata.Sum[int64]{
					DataPoints: []metricdata.DataPoint[int64]{
						{
							Attributes: attribute.NewSet(wantPoolAttr),
							Value:      0,
						},
					},
					Temporality: metricdata.CumulativeTemporality,
					IsMonotonic: false,
				},
			},
		},
	}

	rm := metricdata.ResourceMetrics{}
	err = reader.Collect(ctx, &rm)
	require.NoError(t, err)
	require.Len(t, rm.ScopeMetrics, 1, "should export metrics")
	got := rm.ScopeMetrics[0]
	metricdatatest.AssertEqual(t, want, got, metricdatatest.IgnoreTimestamp())
}
