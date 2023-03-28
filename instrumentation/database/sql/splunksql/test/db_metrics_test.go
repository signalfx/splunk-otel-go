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
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

func TestMetrics(t *testing.T) {
	ctx := context.Background()

	driverName := "simple-driver"
	driver := newSimpleMockDriver()
	connCfg := splunksql.ConnectionConfig{
		// Do not set the Name field so monikers are used to identify
		// spans.
		ConnectionString: "mockDB://bob@localhost:8080/testDB",
		User:             "bob",
		Host:             "localhost",
		Port:             8080,
		NetTransport:     splunksql.NetTransportOther,
	}
	sql.Register(driverName, driver)
	splunksql.Register(driverName, splunksql.InstrumentationConfig{
		DSNParser: func(string) (splunksql.ConnectionConfig, error) { return connCfg, nil },
	})

	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))

	db, err := splunksql.Open(driverName, "dataSourceName",
		splunksql.WithMeterProvider(meterProvider))
	require.NoError(t, err)
	defer db.Close()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.ExecContext(ctx, "SELECT 1") // 1 active connection
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	err = reader.Collect(ctx, &rm)
	require.NoError(t, err)
	require.Len(t, rm.ScopeMetrics, 1, "should export metrics")
	metrics := rm.ScopeMetrics[0]

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
