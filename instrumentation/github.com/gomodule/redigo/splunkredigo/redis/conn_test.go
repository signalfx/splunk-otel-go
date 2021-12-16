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
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type myStr bool

func (b myStr) String() string {
	return strconv.FormatBool(bool(b))
}

func TestParams(t *testing.T) {
	tests := []struct {
		cmd        string
		args       []interface{}
		wantName   string
		wantConfig trace.SpanConfig
	}{
		{
			cmd:      "",
			wantName: "redigo.Conn.Flush",
			wantConfig: trace.NewSpanStartConfig(
				trace.WithAttributes(semconv.DBSystemRedis),
				trace.WithSpanKind(trace.SpanKindClient),
			),
		},
		{
			cmd:      "SET",
			wantName: "SET",
			wantConfig: trace.NewSpanStartConfig(
				trace.WithAttributes(
					semconv.DBSystemRedis,
					semconv.DBOperationKey.String("SET"),
				),
				trace.WithSpanKind(trace.SpanKindClient),
			),
		},
		{
			cmd: "SET",
			args: []interface{}{
				"zero",
				int(1),
				int8(2),
				int16(3),
				int32(4),
				int64(5),
				struct{}{}, // skipped
				myStr(true),
			},
			wantName: "SET",
			wantConfig: trace.NewSpanStartConfig(
				trace.WithAttributes(
					semconv.DBSystemRedis,
					semconv.DBOperationKey.String(
						"SET zero 1 2 3 4 5 true",
					),
				),
				trace.WithSpanKind(trace.SpanKindClient),
			),
		},
	}

	for _, test := range tests {
		gotName, gotOpts := new(otelConn).params(test.cmd, test.args...)
		assert.Equal(t, test.wantName, gotName)
		assert.Equal(t, test.wantConfig, trace.NewSpanStartConfig(gotOpts...))
	}
}

type combined struct {
	redis.ConnWithTimeout
	redis.ConnWithContext
}

func (*combined) Close() error                                   { return nil }
func (*combined) Err() error                                     { return nil }
func (*combined) Do(string, ...interface{}) (interface{}, error) { return nil, nil }
func (*combined) Send(string, ...interface{}) error              { return nil }
func (*combined) Flush() error                                   { return nil }
func (*combined) Receive() (interface{}, error)                  { return nil, nil }

func TestNewConnWithTimeoutAndConnWithContext(t *testing.T) {
	conn := newConn(new(combined))
	assert.Implements(t, (*redis.ConnWithTimeout)(nil), conn)
	assert.Implements(t, (*redis.ConnWithContext)(nil), conn)
}

func TestNewConnWithContext(t *testing.T) {
	conn := newConn(struct{ redis.ConnWithContext }{})
	assert.Implements(t, (*redis.Conn)(nil), conn)
	assert.Implements(t, (*redis.ConnWithContext)(nil), conn)
}

func TestNewConnWithTimeout(t *testing.T) {
	conn := newConn(struct{ redis.ConnWithTimeout }{})
	assert.Implements(t, (*redis.Conn)(nil), conn)
	assert.Implements(t, (*redis.ConnWithTimeout)(nil), conn)
}

func TestNewConn(t *testing.T) {
	conn := newConn(struct{ redis.Conn }{})
	assert.Implements(t, (*redis.Conn)(nil), conn)
}

type receiveWithTimeoutChecker struct {
	combined

	called bool
}

func (r *receiveWithTimeoutChecker) ReceiveWithTimeout(time.Duration) (interface{}, error) {
	r.called = true
	return nil, nil
}

func TestReceiveWithTimeoutForwarded(t *testing.T) {
	r := new(receiveWithTimeoutChecker)
	conn := newConn(r)
	_, err := conn.(redis.ConnWithTimeout).ReceiveWithTimeout(time.Minute)
	require.NoError(t, err)
	assert.True(t, r.called, "ReceiveWithTimeout not forwarded")
}

type receiveContextChecker struct {
	combined

	called bool
}

func (r *receiveContextChecker) ReceiveContext(context.Context) (interface{}, error) {
	r.called = true
	return nil, nil
}

func TestReceiveContextForwarded(t *testing.T) {
	r := new(receiveContextChecker)
	conn := newConn(r)
	_, err := conn.(redis.ConnWithContext).ReceiveContext(nil)
	require.NoError(t, err)
	assert.True(t, r.called, "ReceiveContext not forwarded")
}
