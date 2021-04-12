# Splunk specific instrumentation for `net/http`

## Example

Simplified example:

```go
package main

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/signalfx/splunk-otel-go/distro"
	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
)

func main() {
	distro.Run()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})
	handler = splunkhttp.NewHandler(handler, "my-service")

	http.ListenAndServe(":9090", handler)
}
```

## Configuration

### Splunk distribution configuration

| Code                                                       | Environment variable                   | Default value  | Purpose                                         |
| ---------------------------------------------------------- | -------------------------------------- | -------------- | ----------------------------------------------- |
| `WithTraceResponseHeader`, `TraceResponseHeaderMiddleware` | `SPLUNK_TRACE_RESPONSE_HEADER_ENABLED` | `true`         | Adds `Server-Timing` header to HTTP responses.  |

## Features

### Trace linkage between the APM and RUM products

`TraceResponseHeaderMiddleware` wraps the passed handler, functioning like middleware.
It adds trace context in [traceparent form](https://www.w3.org/TR/trace-context/#traceparent-header)
as [Server-Timing header](https://www.w3.org/TR/server-timing/) to the HTTP response.
