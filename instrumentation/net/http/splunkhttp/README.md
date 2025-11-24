# Splunk specific instrumentation for `net/http`

## Example

Simplified example:

```go
package main

import (
	"net/http"

	"github.com/signalfx/splunk-otel-go/distro"
	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	distro.Run()

	var handler http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello"))
		}
	)
	handler = splunkhttp.NewHandler(handler)
	handler = otelhttp.NewHandler(handler, "my-service")

	http.ListenAndServe(":9090", handler)
}
```

## Configuration

### Splunk distribution configuration

<!-- markdownlint-disable MD013 -->
| Environment variable                   | Default value  | Purpose                                                                                                |
| -------------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------ |
| `SPLUNK_TRACE_RESPONSE_HEADER_ENABLED` | `true`         | Adds `Server-Timing` header to HTTP responses. [More](#trace-linkage-between-the-apm-and-rum-products) |
<!-- markdownlint-enable MD013 -->

## Features

### Trace linkage between the APM and RUM products

`NewHandler` wraps the passed handler, functioning like middleware.
It adds trace context in [traceparent form](https://www.w3.org/TR/trace-context/#traceparent-header)
as [Server-Timing header](https://www.w3.org/TR/server-timing/) to the HTTP response:

```HTTP
Access-Control-Expose-Headers: Server-Timing
Server-Timing: traceparent;desc="00-<serverTraceId>-<serverSpanId>-01"
```

This information can be later consumed by the [splunk-otel-js-web](https://github.com/signalfx/splunk-otel-js-web)
library.
