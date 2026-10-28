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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestFastPathSpanErrorStatus(t *testing.T) {
	const legacyExec = "legacy exec"
	for _, test := range []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "success", code: codes.Unset},
		{name: "skip", err: driver.ErrSkip, code: codes.Unset},
		{name: "driver error", err: errTest, code: codes.Error},
		{name: "wrapped skip is an error", err: fmt.Errorf("driver failed: %w", driver.ErrSkip), code: codes.Error},
	} {
		for _, operation := range []string{"exec", "query", legacyExec, "legacy query"} {
			t.Run(test.name+"/"+operation, func(t *testing.T) {
				recorder := tracetest.NewSpanRecorder()
				provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
				t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
				conn := newMockConn(test.err)
				var underlying driver.Conn = conn
				switch operation {
				case legacyExec:
					underlying = struct{ connExecer }{conn}
				case "legacy query":
					underlying = struct{ connQuery }{conn}
				}
				wrapped := newConn(underlying, newConfig(WithTracerProvider(provider)))
				var err error
				if operation == "exec" || operation == legacyExec {
					_, err = wrapped.ExecContext(context.Background(), "statement", nil)
				} else {
					var rows driver.Rows
					rows, err = wrapped.QueryContext(context.Background(), "statement", nil)
					require.Nil(t, rows)
				}
				require.Equal(t, test.err, err)
				spans := recorder.Ended()
				require.Len(t, spans, 1)
				require.Equal(t, test.code, spans[0].Status().Code)
				if test.code == codes.Error {
					require.Len(t, spans[0].Events(), 1)
				} else {
					require.Empty(t, spans[0].Events())
				}
			})
		}
	}
}
