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

// Package splunkmysql provides instrumentation for the
// [github.com/go-sql-driver/mysql] package.
//
// To use this package, replace any blank identified imports of the
// github.com/go-sql-driver/mysql package with an import of this package and
// use the splunksql.Open function as a replacement for any sql.Open function
// use. For example, if your code looks like this to start.
//
//	import (
//		"database/sql"
//		_ "github.com/go-sql-driver/mysql"
//	)
//	// ...
//	db, err := sql.Open("mysql", "user:password@/dbname")
//	// ...
//
// Update to this.
//
//	import (
//		_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
//		"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
//	)
//	// ...
//	db, err := splunksql.Open("mysql", "user:password@/dbname")
//	// ...
package splunkmysql

import (
	"net"
	"strconv"

	// Make sure to import this so the instrumented driver is registered.
	"github.com/go-sql-driver/mysql"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

func init() { //nolint:gochecknoinits // register db driver
	splunksql.Register("mysql", splunksql.InstrumentationConfig{
		DBSystem:  splunksql.DBSystemMySQL,
		DSNParser: DSNParser,
	})
}

// DSNParser parses the data source connection name for a connection to a
// MySQL database using the github.com/go-sql-driver/mysql client package.
func DSNParser(dataSourceName string) (splunksql.ConnectionConfig, error) {
	var connCfg splunksql.ConnectionConfig
	cfg, err := mysql.ParseDSN(dataSourceName)
	if err != nil {
		return connCfg, err
	}

	if cfg.Passwd != "" {
		// Redact credentials.
		cfg.Passwd = ""
	}

	connCfg.Name = cfg.DBName
	connCfg.ConnectionString = cfg.FormatDSN()
	connCfg.User = cfg.User

	if cfg.Net != "" {
		host, port, err := net.SplitHostPort(cfg.Addr)
		if err == nil {
			connCfg.Host = host
			if p, err := strconv.Atoi(port); err == nil {
				connCfg.Port = p
			}
		}

		// These are the only two cases the instrumented package knows about.
		switch cfg.Net {
		case "unix":
			connCfg.NetTransport = splunksql.NetTransportPipe
			connCfg.NetSockFamily = splunksql.NetSockFamilyUnix
		case "tcp":
			connCfg.NetTransport = splunksql.NetTransportTCP
			if ip := net.ParseIP(connCfg.Host); ip != nil {
				if ip.To4() != nil {
					connCfg.NetSockFamily = splunksql.NetSockFamilyInet
				} else {
					connCfg.NetSockFamily = splunksql.NetSockFamilyInet6
				}
			}
		}
	}

	return connCfg, nil
}
