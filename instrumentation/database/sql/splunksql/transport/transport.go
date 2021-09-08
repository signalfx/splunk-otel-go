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

// Package transport provides identifiers for communication transport
// protocols conforming to OpenTelemetry semantic conventions.
package transport // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/transport"

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Type is a communication transport protocol.
type Type attribute.KeyValue // nolint: revive

// Attribute returns t as an attribute KeyValue. If t is empty the returned
// attribute will default to a Type Other.
func (t Type) Attribute() attribute.KeyValue {
	if !t.Key.Defined() {
		return semconv.NetTransportOther
	}
	return attribute.KeyValue(t)
}

// Valid transport protocols.
var (
	TCP    = Type(semconv.NetTransportTCP)
	UDP    = Type(semconv.NetTransportUDP)
	IP     = Type(semconv.NetTransportIP)
	Unix   = Type(semconv.NetTransportUnix)
	Pipe   = Type(semconv.NetTransportPipe)
	InProc = Type(semconv.NetTransportInProc)
	Other  = Type(semconv.NetTransportOther)
)
