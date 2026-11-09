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
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestFastPathSpanErrorStatus(t *testing.T) {
	const (
		exec       = "exec"
		legacyExec = "legacy exec"
	)
	for _, test := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "success", code: codes.Unset},
		{name: "skip", err: driver.ErrSkip, code: codes.Unset},
		{name: "driver error", err: errTest, code: codes.Error},
		{name: "wrapped skip is an error", err: fmt.Errorf("driver failed: %w", driver.ErrSkip), code: codes.Error},
		{name: "same text as skip is an error", err: errors.New(driver.ErrSkip.Error()), code: codes.Error},
		{name: "context canceled", err: context.Canceled, code: codes.Error},
	} {
		for _, operation := range []struct {
			name         string
			exec         int
			execContext  int
			query        int
			queryContext int
		}{
			{name: exec, execContext: 1},
			{name: "query", queryContext: 1},
			{name: legacyExec, exec: 1},
			{name: "legacy query", query: 1},
		} {
			t.Run(test.name+"/"+operation.name, func(t *testing.T) {
				recorder := tracetest.NewSpanRecorder()
				provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
				t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
				conn := newMockConn(test.err)
				var underlying driver.Conn = conn
				switch operation.name {
				case legacyExec:
					underlying = struct{ connExecer }{conn}
				case "legacy query":
					underlying = struct{ connQuery }{conn}
				}
				wrapped := newConn(underlying, newConfig(WithTracerProvider(provider)))
				var err error
				if operation.name == exec || operation.name == legacyExec {
					_, err = wrapped.ExecContext(context.Background(), "statement", nil)
				} else {
					var rows driver.Rows
					rows, err = wrapped.QueryContext(context.Background(), "statement", nil)
					require.Nil(t, rows)
				}
				require.True(t, err == test.err, "returned error must preserve identity")
				require.Equal(t, operation.exec, conn.execN)
				require.Equal(t, operation.execContext, conn.execContextN)
				require.Equal(t, operation.query, conn.queryN)
				require.Equal(t, operation.queryContext, conn.queryContextN)
				spans := recorder.Ended()
				require.Len(t, spans, 1)
				require.Equal(t, test.code, spans[0].Status().Code)
				if test.code == codes.Error {
					require.Equal(t, test.err.Error(), spans[0].Status().Description)
					require.Len(t, spans[0].Events(), 1)
					require.Equal(t, "exception", spans[0].Events()[0].Name)
				} else {
					require.Empty(t, spans[0].Events())
				}
			})
		}
	}
}

func TestErrSkipOutsideConnectionFastPathsIsAnError(t *testing.T) {
	for _, operation := range []string{"ping", "statement exec", "statement query"} {
		t.Run(operation, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			conf := newConfig(WithTracerProvider(provider))
			var err error
			switch operation {
			case "ping":
				err = newConn(newMockConn(driver.ErrSkip), conf).Ping(context.Background())
			case "statement exec":
				_, err = newStmt(newMockStmt(driver.ErrSkip), conf, "").ExecContext(context.Background(), nil)
			case "statement query":
				var rows driver.Rows
				rows, err = newStmt(newMockStmt(driver.ErrSkip), conf, "").QueryContext(context.Background(), nil)
				require.Nil(t, rows)
			}
			require.True(t, err == driver.ErrSkip, "returned error must preserve identity")
			spans := recorder.Ended()
			require.Len(t, spans, 1)
			require.Equal(t, codes.Error, spans[0].Status().Code)
			require.Equal(t, driver.ErrSkip.Error(), spans[0].Status().Description)
			require.Len(t, spans[0].Events(), 1)
			require.Equal(t, "exception", spans[0].Events()[0].Name)
		})
	}
}
