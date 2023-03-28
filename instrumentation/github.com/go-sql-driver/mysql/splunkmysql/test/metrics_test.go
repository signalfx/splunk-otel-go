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
	"encoding/json"
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
	defer func() { assert.NoError(t, meterProvider.Shutdown(ctx)) }()

	// create 1 used connection
	db, err := splunksql.Open("mysql", dsn, splunksql.WithMeterProvider(meterProvider))
	require.NoError(t, err)
	defer func() { assert.NoError(t, db.Close()) }()
	require.NoError(t, db.Ping())

	// assert
	rm := metricdata.ResourceMetrics{}
	err = reader.Collect(ctx, &rm)
	require.NoError(t, err)
	require.Len(t, rm.ScopeMetrics, 1, "should export metrics")
	metrics := rm.ScopeMetrics[0]

	assertMetrics(t, metrics, "db.client.connections.usage", func(m metricdata.Metrics) {
		assert.Equal(t, "{connection}", m.Unit)
	})

	// log all exported metrics
	d, err := json.Marshal(metrics.Metrics)
	require.NoError(t, err)
	t.Log(string(d))
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
