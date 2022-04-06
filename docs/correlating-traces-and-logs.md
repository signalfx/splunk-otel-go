# Correlating Trace and Logs

> The official Splunk documentation for this page is
[Connect Go application trace data with logs](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/instrumentation/connect-traces-logs.html).
For instructions on how to contribute to the docs, see
[CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

The [OpenTelemetry trace API] can be used to extract trace metadata from a
context containing a span. That metadata can then be used to annotate log
events so they are correlated with the trace.

## Extracting Trace Metadata

To extract trace metadata, use the [`SpanContextFromContext`] function. This
extracts the trace metadata from an [`context.Context`] and returns it in
the form of a [`SpanContext`].

```go
spanContext := trace.SpanContextFromContext(ctx)
if !spanContext.IsValid() {
	// ctx does not contain a valid span.
	// There is no trace metadata to add.
	return
}
```

The OpenTelemetry [`SpanContext`] contains the trace and span ID, trace flags
that contain sampling information, and tracestate information which holds
vendor specific tracing-system context. All of this information can be added to
log events to enrich their context, but the trace and span ID are the metadata
that must be added to correlate with the trace.

## Annotating Log Events

Once you have a [`SpanContext`], you can use the trace metadata to
annotate log events. How this is done depends on the logging library used.

### Structured Logging

If you use a structured logger, add the trace metadata as logger fields.

For example, using the [zap] logging library:

```go
logger, _ := zap.NewProduction()
defer logger.Sync()
logger = logger.With(
	zap.String("trace_id", spanContext.TraceID().String()),
	zap.String("span_id", spanContext.SpanID().String()),
	zap.String("trace_flags", spanContext.TraceFlags().String()),
)
logger.Info("failed to fetch URL", zap.String("URL", url))
```

### Unstructured Logging

If using unstructured logging, the trace metadata can be added as a part of the
logged message.

For example using the standard library [`log`] package:

```go
log.Printf(
	"(trace_id: %s, span_id: %s, trace_flags: %s): failed to fetch URL: %s",
	spanContext.TraceID().String(),
	spanContext.SpanID().String(),
	spanContext.TraceFlags().String(),
	url,
)
```

> Make sure to add the metadata following an order you can parse later on.

## HTTP server example

The following is a complete example of an HTTP server using the [chi] framework that
extracts trace metadata and annotates log messages for handled requests.

```go
package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func withTraceMetadata(ctx context.Context, logger *zap.Logger) *zap.Logger {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		// ctx does not contain a valid span.
		// There is no trace metadata to add.
		return logger
	}
	return logger.With(
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
		zap.String("trace_flags", spanContext.TraceFlags().String()),
	)
}

func helloHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := withTraceMetadata(r.Context(), logger)

		n, err := w.Write([]byte("Hello World!\n"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.Error("failed to write request response", zap.Error(err))
		} else {
			l.Info("request handled", zap.Int("response_bytes", n))
		}
	}
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	router := chi.NewRouter()
	router.Use(splunkchi.Middleware())
	router.Get("/hello", helloHandler(logger))
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
```

[OpenTelemetry trace API]: https://pkg.go.dev/go.opentelemetry.io/otel/trace
[`SpanContextFromContext`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#SpanContextFromContext
[`context.Context`]: https://pkg.go.dev/context#Context
[`SpanContext`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#SpanContext
[zap]: https://github.com/uber-go/zap
[`log`]: https://pkg.go.dev/log
[chi]: https://github.com/go-chi/chi
