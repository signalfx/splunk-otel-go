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

//go:build cgo && linux

package test

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	splunkredigo "github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/option"
	splunkredis "github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/redis"
)

var addr string

const db = 15

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("Skipping running heavy integration test in short mode.")
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	resource, err := pool.Run("redis", "6", nil)
	if err != nil {
		log.Fatalf("Could not create redis container: %v", err)
	}

	addr = getHostPort(resource, "6379/tcp")

	// Wait for the Redis to come up using an exponential-backoff retry.
	if err = pool.Retry(func() error {
		conn, e := redis.Dial(
			"tcp", addr, redis.DialConnectTimeout(time.Second),
		)
		defer func() { _ = conn.Close() }()
		if e != nil {
			return e
		}

		_, e = conn.Do("SELECT", db)
		return e
	}); err != nil {
		log.Fatalf("Could not connect to redis server: %s", err)
	}

	code := m.Run()

	// Run sequentially becauase os.Exit will skip a defer.
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func getHostPort(resource *dockertest.Resource, id string) string {
	dockerURL := os.Getenv("DOCKER_HOST")
	if dockerURL == "" {
		return resource.GetHostPort(id)
	}
	u, err := url.Parse(dockerURL)
	if err != nil {
		panic(err)
	}
	return u.Hostname() + ":" + resource.GetPort(id)
}

func TestConnCreation(t *testing.T) {
	tests := []struct {
		name    string
		factory func(...interface{}) (redis.Conn, error)
	}{
		{
			name: "Dial",
			factory: func(o ...interface{}) (redis.Conn, error) {
				opts := append([]interface{}{
					redis.DialDatabase(db),
					option.WithAttributes([]attribute.KeyValue{
						semconv.DBRedisDBIndexKey.Int(db),
					}),
				}, o...)
				return splunkredis.Dial("tcp", addr, opts...)
			},
		},
		{
			name: "DialWithTimeout",
			factory: func(o ...interface{}) (redis.Conn, error) {
				tout := time.Minute
				opts := append([]interface{}{
					redis.DialConnectTimeout(tout),
					redis.DialReadTimeout(tout),
					redis.DialWriteTimeout(tout),
					redis.DialDatabase(db),
					option.WithAttributes([]attribute.KeyValue{
						semconv.DBRedisDBIndexKey.Int(db),
					}),
				}, o...)
				return splunkredis.Dial("tcp", addr, opts...)
			},
		},
		{
			name: "DialContext",
			factory: func(o ...interface{}) (redis.Conn, error) {
				opts := append([]interface{}{
					redis.DialDatabase(db),
					option.WithAttributes([]attribute.KeyValue{
						semconv.DBRedisDBIndexKey.Int(db),
					}),
				}, o...)
				return splunkredis.DialContext(context.Background(), "tcp", addr, opts...)
			},
		},
		{
			name: "DialURL",
			factory: func(o ...interface{}) (redis.Conn, error) {
				u := fmt.Sprintf("redis://%s/%d", addr, db)
				return splunkredis.DialURL(u, o...)
			},
		},
		{
			name: "DialURLContext",
			factory: func(o ...interface{}) (redis.Conn, error) {
				u := fmt.Sprintf("redis://%s/%d", addr, db)
				return splunkredis.DialURLContext(context.Background(), u, o...)
			},
		},
	}

	for _, test := range tests {
		sr := tracetest.NewSpanRecorder()
		tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
		tracer := tp.Tracer("TestConnCreation/" + test.name)
		t.Run(test.name, func(t *testing.T) {
			conn, err := test.factory(option.WithTracerProvider(tp))
			require.NoError(t, err)

			parentSpanName := "parent"
			ctx, parent := tracer.Start(context.Background(), parentSpanName)
			n := testConn(ctx, t, conn)
			parent.End()

			require.NoError(t, tp.Shutdown(context.Background()))
			assertSpans(t, n, parentSpanName, sr.Ended())
		})
	}
}

func testConn(ctx context.Context, t *testing.T, conn redis.Conn) int {
	args := []interface{}{"vehicle", "truck", ctx}

	n := 1
	_, err := conn.Do("SET", args...)
	require.NoError(t, err)

	if connCtx, ok := conn.(redis.ConnWithContext); ok {
		n++
		_, err = connCtx.DoContext(ctx, "SET", args[:len(args)-1]...)
		require.NoError(t, err)
	}

	if connTime, ok := conn.(redis.ConnWithTimeout); ok {
		n++
		_, err = connTime.DoWithTimeout(time.Minute, "SET", args...)
		require.NoError(t, err)
	}
	return n
}

func assertSpans(t *testing.T, n int, parentSpanName string, spans []trace.ReadOnlySpan) {
	var parent trace.ReadOnlySpan
	children := make([]trace.ReadOnlySpan, 0, n)
	for _, s := range spans {
		if s.Name() == parentSpanName {
			require.Nil(t, parent, "multiple parent spans")
			parent = s
			continue
		}
		children = append(children, s)
	}

	require.Len(t, children, n, "wrong number of children spans")

	host, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	for _, child := range children {
		assert.Equal(t, "SET", child.Name())
		assert.Equal(t, traceapi.SpanKindClient, child.SpanKind())
		assert.Equal(t, splunkredigo.Version(), child.InstrumentationScope().Version)
		assert.Equal(t, parent.SpanContext().TraceID(), child.SpanContext().TraceID())

		attrs := child.Attributes()
		assert.Contains(t, attrs, semconv.DBSystemRedis)
		assert.Contains(t, attrs, semconv.DBOperationKey.String("SET vehicle truck"))
		assert.Contains(t, attrs, semconv.NetTransportTCP)
		assert.Contains(t, attrs, semconv.NetPeerPortKey.Int(port))
		if ip := net.ParseIP(host); ip != nil {
			assert.Containsf(t, attrs, semconv.NetSockPeerAddrKey.String(ip.String()), "address %q", addr)
		} else {
			assert.Containsf(t, attrs, semconv.NetPeerNameKey.String(host), "address %q", addr)
		}

		assert.Contains(t, attrs, semconv.DBRedisDBIndexKey.Int(db))
	}
}
