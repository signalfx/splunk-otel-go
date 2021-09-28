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

package splunksql_test

import (
	"strings"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

func ExampleRegister() {
	// For database drivers that are already registered with the database/sql
	// package, a custom connection string parser and information about the
	// driver can be registered with the splunksql package. These will be used
	// when the splunksql.Open function is called for the registered driver.

	splunksql.Register("my-registered-driver", splunksql.InstrumentationConfig{
		DBSystem: splunksql.DBSystemOtherSQL,
		DSNParser: func(dsn string) (splunksql.ConnectionConfig, error) {
			parts := strings.SplitN(dsn, "|", 3)
			name, user, host := parts[0], parts[1], parts[2]
			return splunksql.ConnectionConfig{
				// Be sure to sanitize passwords from this!
				ConnectionString: dsn,
				Name:             name,
				User:             user,
				Host:             host,
				Port:             9876,
				NetTransport:     splunksql.NetTransportIP,
			}, nil
		},
	})

	// Now when splunksql.Open is called the provided InstrumentationConfig
	// will be used to ensure the telemetry produced adheres to OpenTelemetry
	// semantic conventions. E.g.
	//
	//  db, err := splunksql.Open("my-registered-driver", connStr)
	//  if err != nil {
	//  	log.Fatalf("Failed to open database: %#+v", err)
	//  }
	//  defer db.Close()
	//  ...
}
