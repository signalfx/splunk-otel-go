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

package splunkmysql_test

import (
	"testing"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
	"github.com/stretchr/testify/assert"
)

func TestDSNParser(t *testing.T) {
	testcases := []struct {
		name    string
		dsn     string
		connCfg splunksql.ConnectionConfig
		errStr  string
	}{
		{
			name:   "invalid dsn",
			dsn:    "invalid dsn",
			errStr: "invalid DSN: missing the slash separating the database name",
		},
		{
			name: "db name with defaults",
			dsn:  "/testdb",
			connCfg: splunksql.ConnectionConfig{
				Name:             "testdb",
				ConnectionString: "tcp(127.0.0.1:3306)/testdb",
				Host:             "127.0.0.1",
				Port:             3306,
				NetTransport:     splunksql.NetTransportTCP,
			},
		},
		{
			name: "tcp address",
			dsn:  "user:password@tcp(localhost:8080)/testdb",
			connCfg: splunksql.ConnectionConfig{
				Name:             "testdb",
				ConnectionString: "user@tcp(localhost:8080)/testdb",
				User:             "user",
				Host:             "localhost",
				Port:             8080,
				NetTransport:     splunksql.NetTransportTCP,
			},
		},
		{
			name: "unix socket",
			dsn:  "user:password@unix(/tmp)/testdb",
			connCfg: splunksql.ConnectionConfig{
				Name:             "testdb",
				ConnectionString: "user@unix(/tmp)/testdb",
				User:             "user",
				NetTransport:     splunksql.NetTransportUnix,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := splunkmysql.DSNParser(tc.dsn)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.connCfg, got)
		})
	}
}
