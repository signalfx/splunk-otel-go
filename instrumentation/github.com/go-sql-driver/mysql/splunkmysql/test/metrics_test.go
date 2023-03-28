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

package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
)

func TestMetrics(t *testing.T) {
	ctx := context.Background()

	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))

	db, err := splunksql.Open("mysql", dsn, splunksql.WithMeterProvider(meterProvider))
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, db.Close())
		assert.NoError(t, meterProvider.Shutdown(ctx))
	})

	require.NoError(t, db.Ping())

	_, err = db.Exec(createStmt)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)
	stmtIns, err := tx.Prepare(insertStmt)
	t.Cleanup(func() { assert.NoError(t, stmtIns.Close()) })
	require.NoError(t, err)
	for i := 0; i < 25; i++ {
		_, err = stmtIns.Exec(i, (i * i))
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())
	require.NoError(t, stmtIns.Close())

	var sqNum int
	stmtOut, err := db.Prepare(queryStmt)
	t.Cleanup(func() { assert.NoError(t, stmtOut.Close()) })
	require.NoError(t, err)
	require.NoError(t, stmtOut.QueryRow(13).Scan(&sqNum))
	assert.Equal(t, 13*13, sqNum, "failed to query square of 13")

	// Directly do the query.
	require.NoError(t, db.QueryRow(queryStmt, 1).Scan(&sqNum))
	assert.Equal(t, 1, sqNum, "failed to query square of 1")

	rm := metricdata.ResourceMetrics{}
	err = reader.Collect(ctx, &rm)
	require.NoError(t, err)
	require.Len(t, rm.ScopeMetrics, 1, "should export metrics")
	metrics := rm.ScopeMetrics[0]

	t.Logf("%+v", metrics.Metrics[0])
	assertMetrics(t, metrics, "db.client.connections.usage", func(m metricdata.Metrics) {
		assert.Equal(t, "{connection}", m.Unit)
	})
}

func assertMetrics(t *testing.T, got metricdata.ScopeMetrics, name string, fn func(metricdata.Metrics)) {
	t.Helper()
	for _, m := range got.Metrics {
		if m.Name == name {
			fn(m)
			return
		}
	}

	// Not found. Generate assertion failure.
	var gotMetrics []string
	for _, m := range got.Metrics {
		gotMetrics = append(gotMetrics, m.Name)
	}
	assert.Failf(t, "should contain metric", "want: %v, got: %v", name, gotMetrics)
}
