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

// Package splunkpq provides instrumentation for the [github.com/lib/pq]
// package when using [database/sql].
//
// To use this package, replace any blank identified imports of the
// github.com/lib/pq package with an import of this package and
// use the splunksql.Open function as a replacement for any sql.Open function
// use. For example, if your code looks like this to start.
//
//	import (
//		"database/sql"
//		_ "github.com/lib/pq"
//	)
//	// ...
//	db, err := sql.Open("postgres", "postgres://localhost:5432/dbname")
//	// ...
//
// Update to this.
//
//	import (
//		_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq"
//		"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
//	)
//	// ...
//	db, err := splunksql.Open("postgres", "postgres://localhost:5432/dbname")
//	// ...
package splunkpq

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	// Make sure to import this so the instrumented driver is registered.
	_ "github.com/lib/pq"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq/internal"
)

func init() { //nolint:gochecknoinits // register db driver
	splunksql.Register("postgres", splunksql.InstrumentationConfig{
		DBSystem:  splunksql.DBSystemPostgreSQL,
		DSNParser: DSNParser,
	})
}

// DSNParser parses the data source connection name for a connection to a
// Postgres database using the github.com/lib/pq client package.
func DSNParser(dataSourceName string) (splunksql.ConnectionConfig, error) {
	var connCfg splunksql.ConnectionConfig
	vals, err := internal.ParseDSN(dataSourceName)
	if err != nil {
		return connCfg, err
	}

	connCfg.Name = vals["dbname"]
	connCfg.User = vals["user"]
	if h, ok := vals["host"]; ok {
		connCfg.Host = h
	} else {
		connCfg.Host = "localhost"
	}
	if strings.HasPrefix(connCfg.Host, "/") {
		connCfg.NetTransport = splunksql.NetTransportPipe
		connCfg.NetSockFamily = splunksql.NetSockFamilyUnix
	} else {
		connCfg.NetTransport = splunksql.NetTransportTCP
		if ip := net.ParseIP(connCfg.Host); ip != nil {
			if ip.To4() != nil {
				connCfg.NetSockFamily = splunksql.NetSockFamilyInet
			} else {
				connCfg.NetSockFamily = splunksql.NetSockFamilyInet6
			}
		}
	}
	if pInt, err := strconv.Atoi(vals["port"]); err == nil {
		connCfg.Port = pInt
	}

	// Redact password.
	delete(vals, "password")
	parts := make([]string, 0, len(vals))
	for k, v := range vals {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	// Make this reproducible.
	sort.Strings(parts)
	connCfg.ConnectionString = strings.Join(parts, " ")

	return connCfg, nil
}
