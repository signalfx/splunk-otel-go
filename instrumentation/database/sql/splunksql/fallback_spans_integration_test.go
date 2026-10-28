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

package splunksql_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
)

var fallbackSpanDriverID atomic.Uint64

func TestSuccessfulPreparedFallbackSpans(t *testing.T) {
	for _, operation := range []string{"exec", "query"} {
		t.Run(operation, func(t *testing.T) {
			driverName := fmt.Sprintf("splunktest-fallback-spans-%s-%d", operation, fallbackSpanDriverID.Add(1))
			s, err := NewSplunkSQLSuite(driverName, newSimpleMockDriver())
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, s.DB.Close())
				require.NoError(t, s.TracerProvider.Shutdown(context.Background()))
			})
			if operation == "exec" {
				_, err = s.DB.ExecContext(context.Background(), "statement")
				require.NoError(t, err)
			} else {
				rows, queryErr := s.DB.QueryContext(context.Background(), "statement")
				require.NoError(t, queryErr)
				defer func() { require.NoError(t, rows.Close()) }()
				require.NoError(t, rows.Err())
			}
			spans := s.SpanRecorder.Ended()
			require.NotEmpty(t, spans)
			for _, span := range spans {
				require.NotEqual(t, codes.Error, span.Status().Code, span.Name())
				require.Empty(t, span.Events(), span.Name())
			}
		})
	}
}
