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

package splunkmysql

import (
	"net"
	"strconv"

	// Make sure to import this so the instrumented driver is registered.
	"github.com/go-sql-driver/mysql"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/dbsystem"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/transport"
)

func init() {
	splunksql.Register("mysql", splunksql.InstrumentationConfig{
		DBSystem: dbsystem.MySQL,
		DSNParser: func(dsn string) (splunksql.ConnectionConfig, error) {
			var connCfg splunksql.ConnectionConfig
			cfg, err := mysql.ParseDSN(dsn)
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
				switch cfg.Net {
				case "pipe":
					connCfg.Transport = transport.Pipe
				case "unix", "socket":
					connCfg.Transport = transport.Unix
				case "memory":
					connCfg.Transport = transport.InProc
				case "tcp":
					connCfg.Transport = transport.TCP
				}

				host, port, err := net.SplitHostPort(cfg.Addr)
				if err == nil {
					connCfg.Host = host
					if p, err := strconv.Atoi(port); err == nil {
						connCfg.Port = p
					}
				}
			}

			return connCfg, nil
		},
	})
}
