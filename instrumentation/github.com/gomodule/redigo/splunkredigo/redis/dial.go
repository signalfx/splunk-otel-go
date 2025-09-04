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
	"net"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/semconv/v1.17.0/netconv"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/option"
)

// Dial dials into the network address and returns a traced redis.Conn. The
// set of supported options must be either of type redis.DialOption or this
// package's DialOption. The returned redis.Conn is traced.
func Dial(network, address string, options ...interface{}) (redis.Conn, error) {
	return DialContext(context.Background(), network, address, options...)
}

// DialContext connects to the Redis server at the given network and address
// using the specified options and context. The set of supported options must
// be either of type redis.DialOption or this package's DialOption. The
// returned redis.Conn is traced.
func DialContext(ctx context.Context, network, address string, options ...interface{}) (redis.Conn, error) {
	dialOpts, localOpts := parseOptions(options...)
	c, err := redis.DialContext(ctx, network, address, dialOpts...)
	if err != nil {
		return nil, err
	}

	const parsedOpts = 1
	o := make([]option.Option, len(localOpts)+parsedOpts)
	o[0] = option.WithAttributes(netAttributes(network, address))
	if len(localOpts) > 0 {
		copy(o[1:], localOpts)
	}

	return newConn(c, o...), nil
}

var pathDBRegexp = regexp.MustCompile(`/(\d*)\z`)

// DialURL connects to a Redis server at the given URL using the Redis URI
// scheme. URLs should follow the draft IANA specification for the scheme
// (https://www.iana.org/assignments/uri-schemes/prov/redis). The returned
// redis.Conn is traced.
func DialURL(rawurl string, options ...interface{}) (redis.Conn, error) {
	return DialURLContext(context.Background(), rawurl, options...)
}

// DialURLContext connects to a Redis server at the given URL using the Redis
// URI scheme. URLs should follow the draft IANA specification for the scheme
// (https://www.iana.org/assignments/uri-schemes/prov/redis). The returned
// redis.Conn is traced.
func DialURLContext(ctx context.Context, rawurl string, options ...interface{}) (redis.Conn, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	var db int
	const matchN = 2
	if match := pathDBRegexp.FindStringSubmatch(u.Path); len(match) == matchN {
		if match[1] != "" {
			db, err = strconv.Atoi(match[1])
			if err != nil {
				return nil, fmt.Errorf("invalid database: %s", u.Path[1:])
			}
		}
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		port = "6379"
	}
	if host == "" {
		host = "localhost"
	}

	dialOpts, localOpts := parseOptions(options...)

	const parsedOpts = 1
	o := make([]option.Option, len(localOpts)+parsedOpts)
	if db > 0 {
		o[0] = option.WithAttributes(append(
			netAttributes("tcp", net.JoinHostPort(host, port)),
			semconv.DBRedisDBIndexKey.Int(db),
		))
	} else {
		o[0] = option.WithAttributes(
			netAttributes("tcp", net.JoinHostPort(host, port)),
		)
	}
	if len(localOpts) > 0 {
		copy(o[1:], localOpts)
	}

	c, err := redis.DialURLContext(ctx, rawurl, dialOpts...)
	return newConn(c, o...), err
}

// parseOptions parses a set of arbitrary options (which can be of type
// redis.DialOption or the local option.Option) and returns the corresponding
// redis.DialOption set as well as a configured dialConfig.
func parseOptions(options ...interface{}) ([]redis.DialOption, []option.Option) {
	dialOpts := []redis.DialOption{}
	localOpts := []option.Option{}
	for _, opt := range options {
		switch o := opt.(type) {
		case redis.DialOption:
			dialOpts = append(dialOpts, o)
		case option.Option:
			localOpts = append(localOpts, o)
		}
	}
	return dialOpts, localOpts
}

func netAttributes(network, address string) []attribute.KeyValue {
	ip, hostname, port := splitAddress(address)

	// Guaranteed to at least return transport attribute.
	n := 1
	if ip != "" {
		n++
	}
	if hostname != "" {
		n++
	}
	if port != 0 {
		n++
	}
	attrs := make([]attribute.KeyValue, 0, n)

	attrs = append(attrs, netconv.Transport(network))
	if hostname != "" {
		attrs = append(attrs, semconv.NetPeerNameKey.String(hostname))
		if port != 0 {
			attrs = append(attrs, semconv.NetPeerPortKey.Int(port))
		}
	} else if ip != "" {
		attrs = append(attrs, semconv.NetSockPeerAddrKey.String(ip))
		if port != 0 {
			attrs = append(attrs, semconv.NetSockPeerPortKey.Int(port))
		}
	}

	return attrs
}

// splitAddress extracts the IP address, hostname and port from address. It
// handles both IPv4 and IPv6 addresses. If the host is not recognized as a
// valid IPv4 or IPv6 address, ip will be empty and hostname will contain the
// extracted hostname. Otherwise, hostname will be empty and ip will contain
// the IP address. If address does not contain a port, port will be zero.
func splitAddress(address string) (ip, hostname string, port int) {
	h, p, err := net.SplitHostPort(address)
	if err != nil {
		h, p = address, ""
	}
	if parsedIP := net.ParseIP(h); parsedIP != nil {
		ip = parsedIP.String()
	} else {
		hostname = h
	}
	const (
		base10 = 10
		bit16  = 16
	)

	if p64, err := strconv.ParseUint(p, base10, bit16); err == nil {
		port = int(p64)
	}
	return ip, hostname, port
}
