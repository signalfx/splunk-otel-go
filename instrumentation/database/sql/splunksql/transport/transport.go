// Package transport provides identifiers for communication transport
// protocols conforming to OpenTelemetry semantic conventions.
package transport

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Type is a communication transport protocol.
type Type attribute.KeyValue

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
