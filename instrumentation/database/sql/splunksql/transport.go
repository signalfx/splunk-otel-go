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
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// NetTransport is a communication transport protocol.
type NetTransport attribute.KeyValue

// Attribute returns t as an attribute KeyValue. If t is empty or a deprecated
// value, the returned attribute will default to a NetTransportOther.
func (t NetTransport) Attribute() attribute.KeyValue {
	if !t.Key.Defined() {
		return semconv.NetTransportOther
	}
	switch t {
	case NetTransportIP, NetTransportUnix:
		return semconv.NetTransportOther
	}
	return attribute.KeyValue(t)
}

// Valid transport protocols.
var (
	NetTransportTCP    = NetTransport(semconv.NetTransportTCP)
	NetTransportUDP    = NetTransport(semconv.NetTransportUDP)
	NetTransportPipe   = NetTransport(semconv.NetTransportPipe)
	NetTransportInProc = NetTransport(semconv.NetTransportInProc)
	NetTransportOther  = NetTransport(semconv.NetTransportOther)

	// Deprecated: Use appropriate NetSockFamily* instead.
	NetTransportIP = NetTransport(semconv.NetTransportKey.String("ip"))
	// Deprecated: Use appropriate NetSockFamily* instead.
	NetTransportUnix = NetTransport(semconv.NetTransportKey.String("unix"))
)

// NetSockFamily is a protocol address family used for communication.
type NetSockFamily attribute.KeyValue

// Attribute returns t as an attribute KeyValue. If s is empty the returned
// attribute will also be an empty, undefined, KeyValue.
func (s NetSockFamily) Attribute() attribute.KeyValue {
	return attribute.KeyValue(s)
}

// Valid protocol address families.
var (
	NetSockFamilyInet  = NetSockFamily(semconv.NetSockFamilyInet)
	NetSockFamilyInet6 = NetSockFamily(semconv.NetSockFamilyInet6)
	NetSockFamilyUnix  = NetSockFamily(semconv.NetSockFamilyUnix)
)
