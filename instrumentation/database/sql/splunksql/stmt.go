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

	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

// otelStmt is traced sql.Stmt.
type otelStmt struct {
	driver.Stmt

	config traceConfig
	query  string
}

// Compile-time check otelStmt implements database interfaces.
var (
	_ driver.Stmt             = (*otelStmt)(nil)
	_ driver.StmtExecContext  = (*otelStmt)(nil)
	_ driver.StmtQueryContext = (*otelStmt)(nil)
)

func newStmt(stmt driver.Stmt, c traceConfig, query string) *otelStmt {
	return &otelStmt{Stmt: stmt, config: c, query: query}
}

// ExecContext executes and traces a query that doesn't return rows, such as
// an INSERT or UPDATE.
func (s *otelStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	var (
		f   func(context.Context) error
		res driver.Result
	)
	if execer, ok := s.Stmt.(driver.StmtExecContext); ok {
		f = func(ctx context.Context) error {
			var err error
			res, err = execer.ExecContext(ctx, args)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Exec.
		vArgs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}
		f = func(ctx context.Context) error {
			var err error
			res, err = s.Exec(vArgs)
			return err
		}
	}

	err := s.config.withSpan(ctx, moniker.Exec, f, trace.WithAttributes(semconv.DBStatementKey.String(s.query)))
	return res, err
}

// QueryContext executes and traces a query that may return rows, such as a
// SELECT.
func (s *otelStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var (
		f    func(context.Context) error
		rows driver.Rows
	)
	if queryer, ok := s.Stmt.(driver.StmtQueryContext); ok {
		f = func(ctx context.Context) error {
			var err error
			rows, err = queryer.QueryContext(ctx, args)
			return err
		}
	} else {
		// Fallback to explicitly wrapping Query.
		vArgs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}
		f = func(ctx context.Context) error {
			var err error
			rows, err = s.Query(vArgs)
			return err
		}
	}

	err := s.config.withSpan(ctx, moniker.Query, f, trace.WithAttributes(semconv.DBStatementKey.String(s.query)))
	if err != nil {
		return nil, err
	}
	return rows, nil
}
