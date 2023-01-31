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

package splunkpq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq"
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
			errStr: "missing \"=\" after \"invalid\" in connection info string\"",
		},
		{
			name: "url: tcp address",
			dsn:  "postgres://user:password@localhost:8080/testdb",
			connCfg: splunksql.ConnectionConfig{
				Name:             "testdb",
				ConnectionString: "dbname=testdb host=localhost port=8080 user=user",
				User:             "user",
				Host:             "localhost",
				Port:             8080,
				NetTransport:     splunksql.NetTransportTCP,
			},
		},
		{
			name: "params: unix socket",
			dsn:  "user=user password=password host=/tmp/pgdb dbname=testdb",
			connCfg: splunksql.ConnectionConfig{
				Name:             "testdb",
				ConnectionString: "dbname=testdb host=/tmp/pgdb port=5432 user=user",
				User:             "user",
				Host:             "/tmp/pgdb",
				Port:             5432,
				NetTransport:     splunksql.NetTransportPipe,
				NetSockFamily:    splunksql.NetSockFamilyUnix,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := splunkpq.DSNParser(tc.dsn)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.connCfg, got)
		})
	}
}
