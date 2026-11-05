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
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

type fallbackSpanConn struct {
	driver.Conn
	stmt       *fallbackSpanStmt
	prepareErr error
	fastCalls  int
	prepares   int
}

func (c *fallbackSpanConn) Prepare(string) (driver.Stmt, error) {
	c.prepares++
	return c.stmt, c.prepareErr
}

func (c *fallbackSpanConn) Exec(string, []driver.Value) (driver.Result, error) {
	c.fastCalls++
	return nil, driver.ErrSkip
}

func (c *fallbackSpanConn) Query(string, []driver.Value) (driver.Rows, error) {
	c.fastCalls++
	return nil, driver.ErrSkip
}

func (c *fallbackSpanConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return c.Exec("", nil)
}

func (c *fallbackSpanConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return c.Query("", nil)
}

type fallbackSpanLegacyConn interface {
	connExecer
	connQuery
}

type fallbackSpanStmt struct {
	driver.Stmt
	err   error
	calls int
}

func (s *fallbackSpanStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	s.calls++
	return driver.RowsAffected(1), s.err
}

func (s *fallbackSpanStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	s.calls++
	return &fallbackSpanRows{}, s.err
}

type fallbackSpanRows struct{ done bool }

func (*fallbackSpanRows) Columns() []string { return []string{"value"} }
func (*fallbackSpanRows) Close() error      { return nil }

func (r *fallbackSpanRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = int64(42)
	return nil
}

func newFallbackSpanDB(t *testing.T, legacy bool) (*sql.DB, *fallbackSpanConn, *tracetest.SpanRecorder) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	conn := &fallbackSpanConn{
		Conn: newMockConn(nil),
		stmt: &fallbackSpanStmt{Stmt: newMockStmt(nil)},
	}
	var underlying driver.Conn = conn
	if legacy {
		underlying = struct{ fallbackSpanLegacyConn }{conn}
	}
	d := newDriver(&namedFallbackDriver{conn: underlying}, newConfig(WithTracerProvider(provider)))
	db := sql.OpenDB(&namedFallbackConnector{driver: d})
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db, conn, recorder
}

func TestSuccessfulPreparedFallbackSpans(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		for _, operation := range []moniker.Span{moniker.Exec, moniker.Query} {
			name := operation.String()
			if legacy {
				name = "legacy " + name
			}
			t.Run(name, func(t *testing.T) {
				db, conn, recorder := newFallbackSpanDB(t, legacy)
				if operation == moniker.Exec {
					result, err := db.ExecContext(context.Background(), "statement")
					require.NoError(t, err)
					affected, err := result.RowsAffected()
					require.NoError(t, err)
					require.Equal(t, int64(1), affected)
				} else {
					rows, err := db.QueryContext(context.Background(), "statement")
					require.NoError(t, err)
					defer func() { require.NoError(t, rows.Close()) }()
					require.True(t, rows.Next())
					var value int64
					require.NoError(t, rows.Scan(&value))
					require.Equal(t, int64(42), value)
					require.False(t, rows.Next())
					require.NoError(t, rows.Err())
				}
				require.Equal(t, 1, conn.fastCalls)
				require.Equal(t, 1, conn.prepares)
				require.Equal(t, 1, conn.stmt.calls)
				spans := recorder.Ended()
				require.Len(t, spans, 3)
				require.Equal(t, operation.String(), spans[0].Name())
				require.Equal(t, moniker.Prepare.String(), spans[1].Name())
				require.Equal(t, operation.String(), spans[2].Name())
				for _, span := range spans {
					require.Equal(t, codes.Unset, span.Status().Code, span.Name())
					require.Empty(t, span.Events(), span.Name())
				}
			})
		}
	}
}

func TestPreparedFallbackFailureSpans(t *testing.T) {
	for _, operation := range []moniker.Span{moniker.Exec, moniker.Query} {
		for _, phase := range []string{"prepare", "statement"} {
			t.Run(operation.String()+"/"+phase, func(t *testing.T) {
				db, conn, recorder := newFallbackSpanDB(t, false)
				wantSpans := 3
				if phase == "prepare" {
					conn.prepareErr = driver.ErrSkip
					wantSpans = 2
				} else {
					conn.stmt.err = driver.ErrSkip
				}
				var err error
				if operation == moniker.Exec {
					_, err = db.ExecContext(context.Background(), "statement")
				} else {
					var value int64
					err = db.QueryRowContext(context.Background(), "statement").Scan(&value)
				}
				require.True(t, err == driver.ErrSkip, "original fallback error must reach the caller")
				require.Equal(t, 1, conn.fastCalls)
				require.Equal(t, 1, conn.prepares)
				require.Equal(t, wantSpans-2, conn.stmt.calls)
				spans := recorder.Ended()
				require.Len(t, spans, wantSpans)
				for _, span := range spans[:len(spans)-1] {
					require.Equal(t, codes.Unset, span.Status().Code)
					require.Empty(t, span.Events())
				}
				failed := spans[len(spans)-1]
				require.Equal(t, codes.Error, failed.Status().Code)
				require.Equal(t, driver.ErrSkip.Error(), failed.Status().Description)
				require.Len(t, failed.Events(), 1)
				require.Equal(t, "exception", failed.Events()[0].Name)
			})
		}
	}
}
