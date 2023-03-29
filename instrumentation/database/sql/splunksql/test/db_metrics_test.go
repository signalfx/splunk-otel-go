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
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/sqltestutil"
)

func TestMetrics(t *testing.T) {
	testCases := []struct {
		driverName       string
		connectionString string
		wantPoolName     string
	}{
		{
			driverName:       "SanitizedConnectionString",
			connectionString: "mockDB://bob@localhost:8080/testDB",
			wantPoolName:     "mockDB://bob@localhost:8080/testDB",
		},
		{
			driverName:       "NoConnectionString",
			connectionString: "",
			wantPoolName:     "NoConnectionString",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.driverName, func(t *testing.T) {
			// instrument: register the fake driver
			driver := newSimpleMockDriver()
			connCfg := splunksql.ConnectionConfig{
				ConnectionString: tc.connectionString, // to make sure that pool.name value is sanitized via DSNParser
				Host:             "localhost",         // to avoid errors in logs
			}
			sql.Register(tc.driverName, driver)
			splunksql.Register(tc.driverName, splunksql.InstrumentationConfig{
				DSNParser: func(string) (splunksql.ConnectionConfig, error) { return connCfg, nil },
			})

			// execute test
			sqltestutil.TestMetrics(t, tc.wantPoolName, tc.driverName, "dataSourceName", func(db *sql.DB) {
				_, err := db.Exec("SELECT 1")
				require.NoError(t, err)
			})
		})
	}
}
