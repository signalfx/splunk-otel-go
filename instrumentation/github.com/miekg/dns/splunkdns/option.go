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

package splunkdns

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/miekg/dns"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns"

// Option applies options to a configuration.
type Option interface {
	internal.Option
}

func localToInternal(opts []Option) []internal.Option {
	out := make([]internal.Option, len(opts))
	for i, o := range opts {
		out[i] = internal.Option(o)
	}
	return out
}

func name(m *dns.Msg) string {
	return "DNS " + dns.OpcodeToString[m.Opcode]
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return Option(internal.WithTracerProvider(tp))
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return Option(internal.WithAttributes(attr))
}
