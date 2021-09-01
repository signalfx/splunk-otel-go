package splunksql // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

import (
	"context"
	"database/sql/driver"
	"io"

	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type otelConn struct {
	driver.Conn

	config config
}

// Compile-time check otelConn implements database interfaces.
var (
	_ driver.Pinger             = (*otelConn)(nil)
	_ driver.ExecerContext      = (*otelConn)(nil)
	_ driver.QueryerContext     = (*otelConn)(nil)
	_ driver.Conn               = (*otelConn)(nil)
	_ driver.ConnPrepareContext = (*otelConn)(nil)
	_ driver.ConnBeginTx        = (*otelConn)(nil)
	_ driver.SessionResetter    = (*otelConn)(nil)
)

func newConn(conn driver.Conn, conf config) *otelConn {
	return &otelConn{Conn: conn, config: conf}
}

// Ping traces a ping to the connected database.
func (c *otelConn) Ping(ctx context.Context) error {
	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}

	var (
		err  error
		span trace.Span
	)
	ctx, span = c.config.Tracer(ctx).Start(ctx, "Ping", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		handleErr(span, err)
		span.End()
	}()

	err = pinger.Ping(ctx)
	return err
}

// ExecContext traces the call to the wrapped Connection ExecContext method.
func (c *otelConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	var span trace.Span
	ctx, span = c.config.Tracer(ctx).Start(
		ctx,
		"Exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBStatementKey.String(query)),
	)
	defer span.End()

	res, err := execer.ExecContext(ctx, query, args)
	handleErr(span, err)
	return res, err
}

// QueryContext traces the call to the wrapped Connection QueryContext method.
func (c *otelConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	// Split the context to make the query and returned rows spans siblings.
	qCtx, span := c.config.Tracer(ctx).Start(
		ctx,
		"Query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBStatementKey.String(query)),
	)
	defer span.End()

	rows, err := queryer.QueryContext(qCtx, query, args)
	if err != nil {
		handleErr(span, err)
		return nil, err
	}

	return newRows(ctx, rows, c.config), nil
}

// PrepareContext returns a prepared statement, bound to this traced connection.
func (c *otelConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	preparer, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	var span trace.Span
	ctx, span = c.config.Tracer(ctx).Start(
		ctx,
		"Prepare",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBStatementKey.String(query)),
	)
	defer span.End()

	stmt, err := preparer.PrepareContext(ctx, query)
	if err != nil {
		handleErr(span, err)
		return nil, err
	}
	return newStmt(stmt, c.config, query), nil
}

// BeginTx starts and returns a new traced transaction.
func (c *otelConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	transactor, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}

	// Split the context to make the begin and returned tx spans siblings.
	tCtx, span := c.config.Tracer(ctx).Start(ctx, "Begin", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	tx, err := transactor.BeginTx(tCtx, opts)
	if err != nil {
		handleErr(span, err)
		return nil, err
	}

	return newTx(ctx, tx, c.config), nil
}

// ResetSession traces the call to the wrapped Connection ResetSession method.
func (c *otelConn) ResetSession(ctx context.Context) error {
	resetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}

	var span trace.Span
	ctx, span = c.config.Tracer(ctx).Start(ctx, "Reset", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	err := resetter.ResetSession(ctx)
	handleErr(span, err)
	return err
}

// copied from stdlib database/sql package: src/database/sql/ctxutil.go
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	vArgs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("splunksql: driver does not support the use of Named Parameters")
		}
		vArgs[n] = param.Value
	}
	return vArgs, nil
}
