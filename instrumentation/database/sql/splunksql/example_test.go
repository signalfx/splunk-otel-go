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
	"fmt"
	"log"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

const (
	dbname = "orders"
	user   = "frank"
	pass   = "hotdog"
	host   = "localhost"
	port   = 9876
)

var (
	connStr   = fmt.Sprintf("custom|%s:%s|%s:%d|%s", user, pass, host, port, dbname)
	sanitized = fmt.Sprintf("custom|%s|%s:%d|%s", user, host, port, dbname)

	// semanticAttrs are the attributes OpenTelemetry requires and recommends.
	semanticAttrs = []attribute.KeyValue{
		semconv.DBSystemOtherSQL,
		semconv.DBNameKey.String(dbname),
		// Do not include passwords!
		semconv.DBConnectionStringKey.String(sanitized),
		semconv.DBUserKey.String(user),
		// Use semconv.NetPeerIPKey if connecting via an IP address. If
		// connecting via a Unix socket, use this attribute key.
		semconv.NetPeerNameKey.String(host),
		semconv.NetPeerPortKey.Int(port),
		semconv.NetTransportTCP,
	}
)

func Example() {
	// For database drivers that are already registered with the database/sql
	// package, you can pass the OpenTelemetry attributes directly to the
	// splunksql.Open function. These attributes will be attached to all spans
	// the instrumentation creates.

	db, err := splunksql.Open(
		"already-registered-driver",
		connStr,
		splunksql.WithAttributes(semanticAttrs),
	)
	if err != nil {
		log.Fatalf("Failed to open database: %#+v", err)
	}
	defer db.Close()

	// Use the traced db...
}
