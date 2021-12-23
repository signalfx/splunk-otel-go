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

package splunksql

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

func TestURLDNSParse(t *testing.T) {
	testcases := []struct {
		name        string
		dsn         string
		expectedCfg ConnectionConfig
		errStr      string
	}{
		{
			name: "not a URL",
			dsn:  `:¯\_(ツ)_/¯:`,
			errStr: (&url.Error{
				Op:  "parse",
				URL: `:¯\_(ツ)_/¯:`,
				Err: errors.New("missing protocol scheme"),
			}).Error(),
		},
		{
			name: "params",
			dsn:  "param0=val0,paramN=valN",
			expectedCfg: ConnectionConfig{
				ConnectionString: "param0=val0,paramN=valN",
			},
		},
		{
			name: "host only",
			dsn:  "http://localhost",
			expectedCfg: ConnectionConfig{
				ConnectionString: "http://localhost",
				Host:             "localhost",
			},
		},
		{
			name: "host:port",
			dsn:  "https://localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://localhost:8080",
				Host:             "localhost",
				Port:             8080,
			},
		},
		{
			name: "with user",
			dsn:  "https://bob@localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://bob@localhost:8080",
				User:             "bob",
				Host:             "localhost",
				Port:             8080,
			},
		},
		{
			name: "redact password",
			dsn:  "https://bob:pa55w0rd@localhost:8080",
			expectedCfg: ConnectionConfig{
				ConnectionString: "https://bob@localhost:8080",
				User:             "bob",
				Host:             "localhost",
				Port:             8080,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			connCfg, err := urlDSNParse(tc.dsn)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCfg, connCfg)
		})
	}
}

func TestSpanName(t *testing.T) {
	c := newTraceConfig()

	// c.DBName empty means the moniker should be used.
	m := moniker.Begin
	assert.Equal(t, m.String(), c.spanName(m))

	const dbname = "test database"
	c.DBName = dbname
	assert.Equal(t, dbname, c.spanName(m))
}
