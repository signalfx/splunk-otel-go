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

package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func TestNetAttributes(t *testing.T) {
	networks := map[string]attribute.KeyValue{
		"tcp":        semconv.NetTransportTCP,
		"tcp4":       semconv.NetTransportTCP,
		"tcp6":       semconv.NetTransportTCP,
		"udp":        semconv.NetTransportUDP,
		"udp4":       semconv.NetTransportUDP,
		"udp6":       semconv.NetTransportUDP,
		"ip":         semconv.NetTransportIP,
		"ip4":        semconv.NetTransportIP,
		"ip6":        semconv.NetTransportIP,
		"unix":       semconv.NetTransportUnix,
		"unixgram":   semconv.NetTransportUnix,
		"unixpacket": semconv.NetTransportUnix,
		"redis":      semconv.NetTransportOther,
		"rediss":     semconv.NetTransportOther,
		"":           semconv.NetTransportOther,
	}

	addresses := map[string][]attribute.KeyValue{
		"": {},
		"localhost": {
			semconv.NetPeerNameKey.String("localhost"),
		},
		"localhost:80": {
			semconv.NetPeerNameKey.String("localhost"),
			semconv.NetPeerPortKey.Int(80),
		},
		"127.0.0.1": {
			semconv.NetPeerIPKey.String("127.0.0.1"),
		},
		"127.0.0.1:80": {
			semconv.NetPeerIPKey.String("127.0.0.1"),
			semconv.NetPeerPortKey.Int(80),
		},
	}

	for net, netAttr := range networks {
		for addr, addrAttrs := range addresses {
			want := append([]attribute.KeyValue{netAttr}, addrAttrs...)
			got := netAttributes(net, addr)
			assert.ElementsMatch(t, want, got)
		}
	}
}
