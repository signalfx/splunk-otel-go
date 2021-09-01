package splunksql

import (
	"database/sql/driver"
	"io"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type spanName string

const (
	querySpan    spanName = "Query"
	pingSpan              = "Ping"
	prepareSpan           = "Prepare"
	execSpan              = "Exec"
	beginSpan             = "Begin"
	resetSpan             = "Reset"
	closeSpan             = "Close"
	commitSpan            = "Commit"
	rollbackSpan          = "Rollback"
)

func (n spanName) String() string { return string(n) }

func handleErr(span trace.Span, err error) {
	if span == nil {
		return
	}

	switch err {
	case nil:
		// Everything Okay.
	case io.EOF:
		// Expected at end of iteration, do not record these.
	case driver.ErrSkip:
		// Expected if method not implemented, do not record these.
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
