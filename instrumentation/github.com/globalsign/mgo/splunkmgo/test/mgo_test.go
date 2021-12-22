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
// +build cgo,linux

package test

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/globalsign/mgo/splunkmgo"
)

var addr string

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

	resource, err := pool.Run("mongo", "5.0", nil)
	if err != nil {
		log.Fatalf("Could not create mongo container: %v", err)
	}

	// If run with docker-machine the hostname needs to be set.
	u, err := url.Parse(pool.Client.Endpoint())
	if err != nil {
		log.Fatalf("Could not parse endpoint: %s", pool.Client.Endpoint())
	}
	hostname := u.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}
	addr = net.JoinHostPort(hostname, resource.GetPort("27017/tcp"))

	// Wait for MongoDB to come up using an exponential-backoff retry.
	if err = pool.Retry(func() error {
		session, e := mgo.Dial(addr)
		if e != nil {
			return e
		}
		defer session.Close()

		return session.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to mongo server: %s", err)
	}

	code := m.Run()

	// Run sequentially becauase os.Exit will skip a defer.
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// fixtures returns a SpanRecorder and TracerProvider. The SpanRecorder is
// registered with the returned TracerProvider as a SpanProcessor. It is the
// responsibility of the caller to shut down the TracerProvider.
func fixtures() (*tracetest.SpanRecorder, *trace.TracerProvider) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	return sr, tp
}

func TestSessionPing(t *testing.T) {
	sr, tp := fixtures()

	session, err := splunkmgo.Dial(addr, splunkmgo.WithTracerProvider(tp))
	require.NoError(t, err)
	t.Cleanup(session.Close)
	require.NoError(t, session.Ping())

	require.NoError(t, tp.Shutdown(context.Background()))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, "ping", span.Name())

	attrs := span.Attributes()
	assert.Contains(t, attrs, semconv.DBSystemMongoDB)
	assert.Contains(t, attrs, semconv.DBNameKey.String("admin"))
}

func TestSessionDefaultDB(t *testing.T) {
	sr, tp := fixtures()

	session, err := splunkmgo.Dial(addr, splunkmgo.WithTracerProvider(tp))
	require.NoError(t, err)
	t.Cleanup(session.Close)
	// Passing an empty string should default to a "test" Database since no
	// explicit default was defined for the session.
	require.NoError(t, session.DB("").Run("ping", nil))

	require.NoError(t, tp.Shutdown(context.Background()))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	attrs := spans[0].Attributes()
	assert.Contains(t, attrs, semconv.DBNameKey.String("test"))
}

func TestSessionWithDefaultDB(t *testing.T) {
	sr, tp := fixtures()

	info, err := mgo.ParseURL(addr)
	require.NoError(t, err)

	const defaultDB = "TestSessionWithDefaultDB"
	info.Database = defaultDB

	session, err := splunkmgo.DialWithInfo(info, splunkmgo.WithTracerProvider(tp))
	require.NoError(t, err)
	t.Cleanup(session.Close)
	// Passing an empty string should default to the default Database defined
	// above when creating the session.
	require.NoError(t, session.DB("").Run("ping", nil))

	require.NoError(t, tp.Shutdown(context.Background()))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	attrs := spans[0].Attributes()
	assert.Contains(t, attrs, semconv.DBNameKey.String(defaultDB))
}
