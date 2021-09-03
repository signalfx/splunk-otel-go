package splunksql

import (
	"database/sql/driver"
	"io"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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
