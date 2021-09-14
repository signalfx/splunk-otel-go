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

// Package splunkpgx provides instrumentation for the github.com/jackc/pgx
// package when using database/sql.
//
// To use this package, replace any blank identified imports of the
// github.com/jackc/pgx package with an import of this package and
// use the splunksql.Open function as a replacement for any sql.Open function
// use. For example, if your code looks like this to start.
//
//     import (
//     	"database/sql"
//     	_ "github.com/jackc/pgx"
//     )
//     // ...
//     db, err := sql.Open("pgx", "postgres://localhost:5432/dbname")
//     // ...
//
// Update to this.
//
//     import (
//     	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx"
//     	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
//     )
//     // ...
//     db, err := splunksql.Open("pgx", "postgres://localhost:5432/dbname")
//     // ...
package splunkpgx

import (
	"net/url"
	"strings"

	"github.com/jackc/pgx"
	// Make sure to import this so the instrumented driver is registered.
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/dbsystem"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/transport"
)

func init() { // nolint: gochecknoinits
	splunksql.Register("pgx", splunksql.InstrumentationConfig{
		DBSystem:  dbsystem.PostgreSQL,
		DSNParser: DSNParser,
	})
}

// DSNParser parses the data source connection name for a connection to a
// Postgres database using the github.com/jackc/pgx client package.
func DSNParser(dataSourceName string) (splunksql.ConnectionConfig, error) {
	var connCfg splunksql.ConnectionConfig
	c, err := pgx.ParseConnectionString(dataSourceName)
	if err != nil {
		return connCfg, err
	}

	connCfg.Name = c.Database
	connCfg.User = c.User
	if c.Host != "" {
		connCfg.Host = c.Host
	} else {
		connCfg.Host = "localhost"
	}
	if strings.HasPrefix(connCfg.Host, "/") {
		connCfg.Transport = transport.Unix
	} else {
		connCfg.Transport = transport.TCP
		if c.Port > 0 {
			connCfg.Port = int(c.Port)
		} else {
			connCfg.Port = 5432
		}
	}

	if c.Password == "" {
		connCfg.ConnectionString = dataSourceName
	} else {
		connCfg.ConnectionString = redactPassword(dataSourceName)
	}

	return connCfg, nil
}

// redactPassword returns the dsn with the password field removed.
func redactPassword(dsn string) string {
	if u, err := url.Parse(dsn); err == nil && u.Scheme != "" {
		if u.User != nil {
			u.User = url.User(u.User.Username())
		}
		return u.String()
	}

	parts := strings.Split(dsn, " ")
	width := 2
	for i := len(parts) - 1; i >= 0; i-- {
		vals := strings.SplitN(parts[i], "=", width)
		if len(vals) < width {
			continue
		}
		key := strings.Trim(vals[0], " \t\n\r\v\f")
		if key == "password" {
			parts = append(parts[:i], parts[i+1:]...)
		}
	}

	return strings.Join(parts, " ")
}
