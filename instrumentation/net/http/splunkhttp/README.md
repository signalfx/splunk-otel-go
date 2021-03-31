# Splunk specific instrumentation for `net/http`

## Features

### Trace linkage between the APM and RUM products

`ServerTimingMiddleware` wraps the passed handler, functioning like middleware.
It adds trace context in [traceparent form](https://www.w3.org/TR/trace-context/#traceparent-header)
as [Server-Timing header](https://www.w3.org/TR/server-timing/) to the HTTP response.

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
	handler = splunkhttp.ServerTimingMiddleware(handler)
	handler = otelhttp.NewHandler(handler, "my-service")

	http.ListenAndServe(":9090", handler)
}
```
