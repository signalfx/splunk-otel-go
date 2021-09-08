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

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"go.opentelemetry.io/otel/trace"
)

// otelRows traces driver.Rows functionality.
type otelRows struct {
	driver.Rows

	span   trace.Span
	config traceConfig
}

// Compile-time check otelRows implements driver.Rows.
var _ driver.Rows = (*otelRows)(nil)

func newRows(ctx context.Context, rows driver.Rows, c traceConfig) *otelRows {
	_, span := c.tracer(ctx).Start(ctx, moniker.Rows.String(), trace.WithSpanKind(trace.SpanKindClient))
	return &otelRows{
		Rows:   rows,
		span:   span,
		config: c,
	}
}

func (r otelRows) Close() error { // nolint: gocritic
	defer func() {
		if r.span != nil {
			r.span.End()
		}
	}()

	err := r.Rows.Close()
	handleErr(r.span, err)
	return err
}

func (r otelRows) Next(dest []driver.Value) error { // nolint: gocritic
	defer func() {
		if r.span != nil {
			r.span.AddEvent(moniker.Next.String())
		}
	}()

	err := r.Rows.Next(dest)
	handleErr(r.span, err)
	return err
}
