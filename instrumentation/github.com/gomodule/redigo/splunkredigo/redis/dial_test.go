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
	"context"
	"fmt"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func TestNetAttributes(t *testing.T) {
	networks := map[string]attribute.KeyValue{
		"tcp":        semconv.NetTransportTCP,
		"tcp4":       semconv.NetTransportTCP,
		"tcp6":       semconv.NetTransportTCP,
		"udp":        semconv.NetTransportUDP,
		"udp4":       semconv.NetTransportUDP,
		"udp6":       semconv.NetTransportUDP,
		"ip":         semconv.NetTransportOther,
		"ip4":        semconv.NetTransportOther,
		"ip6":        semconv.NetTransportOther,
		"unix":       semconv.NetTransportInProc,
		"unixgram":   semconv.NetTransportInProc,
		"unixpacket": semconv.NetTransportInProc,
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
			semconv.NetSockPeerAddrKey.String("127.0.0.1"),
		},
		"127.0.0.1:80": {
			semconv.NetSockPeerAddrKey.String("127.0.0.1"),
			semconv.NetSockPeerPortKey.Int(80),
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

func TestDialContextForwardsError(t *testing.T) {
	// This should fail because it is not going to be able to connect to a
	// Redis server and the lookup of DB 15 will fail.
	_, err := DialContext(context.Background(), "tcp", "not.valid.localhost", redis.DialDatabase(15))
	assert.Error(t, err)
}

func TestDialURLContextErrorsForInvalidURL(t *testing.T) {
	ctx := context.Background()
	u := "\not a valid URL/"
	_, err := DialURLContext(ctx, u)
	want := fmt.Sprintf("parse %q: net/url: invalid control character in URL", u)
	assert.EqualError(t, err, want)
}

func TestDialURLContextErrorsForInvalidDBPath(t *testing.T) {
	ctx := context.Background()
	db := "9999999999999999999"
	u := "redis://localhost:6379/" + db
	_, err := DialURLContext(ctx, u)
	assert.EqualError(t, err, "invalid database: "+db)
}

func TestDialURLContextAttrs(t *testing.T) {
	tests := []struct {
		name string
		u    string
		want []attribute.KeyValue
	}{
		{
			name: "db",
			u:    "redis://fake.localhost:6379/15",
			want: []attribute.KeyValue{
				semconv.NetTransportTCP,
				semconv.NetPeerNameKey.String("fake.localhost"),
				semconv.NetPeerPortKey.Int(6379),
				semconv.DBRedisDBIndexKey.Int(15),
			},
		},
		{
			name: "default address",
			u:    "",
			want: []attribute.KeyValue{
				semconv.NetTransportTCP,
				semconv.NetPeerNameKey.String("localhost"),
				semconv.NetPeerPortKey.Int(6379),
			},
		},
		{
			name: "default port",
			u:    "redis://fake.localhost",
			want: []attribute.KeyValue{
				semconv.NetTransportTCP,
				semconv.NetPeerNameKey.String("fake.localhost"),
				semconv.NetPeerPortKey.Int(6379),
			},
		},
		{
			name: "custom port",
			u:    "redis://fake.localhost:80",
			want: []attribute.KeyValue{
				semconv.NetTransportTCP,
				semconv.NetPeerNameKey.String("fake.localhost"),
				semconv.NetPeerPortKey.Int(80),
			},
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Do not check error as we only care about conn.
			conn, _ := DialURLContext(ctx, test.u)
			oConn := conn.(struct{ redis.Conn }).Conn.(*otelConn)
			sConf := trace.NewSpanStartConfig(oConn.cfg.DefaultStartOpts...)
			attrs := sConf.Attributes()

			for _, want := range test.want {
				assert.Contains(t, attrs, want)
			}
		})
	}
}
