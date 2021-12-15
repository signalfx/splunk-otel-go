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

// Package redis provides tracing functionality for the
// github.com/gomodule/redigo/redis package.
package redis

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/option"
	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo"

type otelConn struct {
	redis.Conn

	cfg internal.Config
}

var (
	_ redis.Conn            = (*otelConn)(nil)
	_ redis.ConnWithTimeout = (*otelConn)(nil)
	_ redis.ConnWithContext = (*otelConn)(nil)
)

func localToInternal(opts []option.Option) []internal.Option {
	out := make([]internal.Option, len(opts))
	for i, o := range opts {
		out[i] = internal.Option(o)
	}
	return out
}

// newConn returns a redis.Conn backed by conn that traces all operations it
// performs with OpenTelemetry.
func newConn(conn redis.Conn, opts ...option.Option) redis.Conn {
	cfg := internal.NewConfig(instrumentationName, localToInternal(opts)...)

	// Remove the functionality the underlying conn does not implement.
	if _, ok := conn.(redis.ConnWithContext); ok {
		if _, ok := conn.(redis.ConnWithTimeout); ok {
			return &otelConn{conn, *cfg}
		}
		return struct{ redis.ConnWithContext }{&otelConn{conn, *cfg}}
	}
	if _, ok := conn.(redis.ConnWithTimeout); ok {
		return struct{ redis.ConnWithTimeout }{&otelConn{conn, *cfg}}
	}
	return struct{ redis.Conn }{&otelConn{conn, *cfg}}
}

func (c *otelConn) params(commandName string, args ...interface{}) (string, []trace.SpanStartOption) {
	name := commandName
	if name == "" {
		// When the command argument to the Do method is "", then the Do
		// method will flush the output buffer. See
		// https://godoc.org/github.com/gomodule/redigo/redis#hdr-Pipelining
		name = "redigo.Conn.Flush"
	}

	const base10 = 10
	var b bytes.Buffer
	b.WriteString(commandName)
	for _, arg := range args {
		b.WriteString(" ")
		switch arg := arg.(type) {
		case string:
			b.WriteString(arg)
		case int:
			b.WriteString(strconv.Itoa(arg))
		case int32:
			b.WriteString(strconv.FormatInt(int64(arg), base10))
		case int64:
			b.WriteString(strconv.FormatInt(arg, base10))
		case fmt.Stringer:
			b.WriteString(arg.String())
		}
	}

	return name, []trace.SpanStartOption{
		trace.WithAttributes(
			semconv.DBSystemRedis,
			semconv.DBOperationKey.String(b.String()),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	}
}

// Do sends a command to the server and returns the received reply.
// This function will use the timeout which was set when the connection is created
func (c *otelConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	ctx := context.Background()
	for i, a := range args {
		if c, ok := a.(context.Context); ok {
			ctx = c
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	name, sso := c.params(commandName, args...)
	err = c.cfg.WithSpan(ctx, name, func(ctx context.Context) error {
		var e error
		reply, e = c.Conn.Do(commandName, args...)
		return e
	}, sso...)
	return
}

// DoWithTimeout sends a command to the server and returns the received reply.
// The timeout overrides the readtimeout set when dialing the connection.
func (c *otelConn) DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error) {
	ctx := context.Background()
	for i, a := range args {
		if c, ok := a.(context.Context); ok {
			ctx = c
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	name, sso := c.params(commandName, args...)
	err = c.cfg.WithSpan(ctx, name, func(ctx context.Context) error {
		var e error
		// This should not panic given the guard in WrapConn.
		reply, e = c.Conn.(redis.ConnWithTimeout).DoWithTimeout(timeout, commandName, args...)
		return e
	}, sso...)
	return
}

// ReceiveWithTimeout receives a single reply from the Redis server.
// The timeout overrides the readtimeout set when dialing the connection.
func (c *otelConn) ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error) {
	// This should not panic given the guard in WrapConn.
	return c.Conn.(redis.ConnWithTimeout).ReceiveWithTimeout(timeout)
}

// DoContext sends a command to server and returns the received reply.
// min(ctx,DialReadTimeout()) will be used as the deadline.
// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
// DialReadTimeout() timeout return err can be checked by strings.Contains(e.Error(), "io/timeout").
// ctx timeout return err context.DeadlineExceeded.
// ctx canceled return err context.Canceled.
func (c *otelConn) DoContext(ctx context.Context, commandName string, args ...interface{}) (reply interface{}, err error) {
	name, sso := c.params(commandName, args...)
	err = c.cfg.WithSpan(ctx, name, func(ctx context.Context) error {
		var e error
		// This should not panic given the guard in WrapConn.
		reply, e = c.Conn.(redis.ConnWithContext).DoContext(ctx, commandName, args...)
		return e
	}, sso...)
	return
}

// ReceiveContext receives a single reply from the Redis server.
// min(ctx,DialReadTimeout()) will be used as the deadline.
// The connection will be closed if DialReadTimeout() timeout or ctx timeout or ctx canceled when this function is running.
// DialReadTimeout() timeout return err can be checked by strings.Contains(e.Error(), "io/timeout").
// ctx timeout return err context.DeadlineExceeded.
// ctx canceled return err context.Canceled.
func (c *otelConn) ReceiveContext(ctx context.Context) (reply interface{}, err error) {
	// This should not panic given the guard in WrapConn.
	return c.Conn.(redis.ConnWithContext).ReceiveContext(ctx)
}
