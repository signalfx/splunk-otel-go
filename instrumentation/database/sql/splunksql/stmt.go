package splunksql

import (
	"context"
	"database/sql/driver"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// otelStmt is traced sql.Stmt.
type otelStmt struct {
	driver.Stmt

	config config
	query  string
}

// Compile-time check otelStmt implements database interfaces.
var (
	_ driver.Stmt             = (*otelStmt)(nil)
	_ driver.StmtExecContext  = (*otelStmt)(nil)
	_ driver.StmtQueryContext = (*otelStmt)(nil)
)

func newStmt(stmt driver.Stmt, c config, query string) *otelStmt {
	return &otelStmt{Stmt: stmt, config: c, query: query}
}

// ExecContext executes and traces a query that doesn't return rows, such as
// an INSERT or UPDATE.
func (s *otelStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := s.Stmt.(driver.StmtExecContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	var span trace.Span
	ctx, span = s.config.Tracer(ctx).Start(
		ctx,
		"Exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBStatementKey.String(s.query)),
	)
	defer span.End()

	res, err := execer.ExecContext(ctx, args)
	handleErr(span, err)
	return res, err
}

// QueryContext executes and traces a query that may return rows, such as a
// SELECT.
func (s *otelStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := s.Stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	// Split the context to make the query and returned rows spans siblings.
	qCtx, span := s.config.Tracer(ctx).Start(
		ctx,
		"Query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBStatementKey.String(s.query)),
	)
	defer span.End()

	rows, err := queryer.QueryContext(qCtx, args)
	if err != nil {
		handleErr(span, err)
		return nil, err
	}

	return newRows(ctx, rows, s.config), nil
}
