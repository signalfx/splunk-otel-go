# Troubleshooting

## Enable Debug Logging

When things are not working, a good first step is to restart the program with
debug logging enabled. Do this by setting the `OTEL_LOG_LEVEL` environment
variable to `debug`.

```sh
export OTEL_LOG_LEVEL="debug"
```

Make sure to unset this logging level after the issue is resolved, it is quite
verbose and likely will overload systems if left on indefinitely.

## Missing Spans

There are many points were spans may be dropped. Follow these steps to ensure
they are not being dropped by the `SDK`.

1. [Enable debug logging](#enable-debug-logging). This should log the `SDK`
   configuration and all spans sent to registered `SpanProcessor`s. Verify the
   correct number of span are reported as exported in the log messages.
   Additionally, pay attention to any `total_dropped` count.
2. If an incorrect number of spans are being reported as exported, verify you
   are ending the spans you expect to see.
3. If spans are reported as dropped, the `BatchSpanProcessor` is dropping
   spans. Increase the capacity of the `BatchSpanProcessor`'s queue or the
   frequency it exports via [environment variables].

Following these steps should help understand if spans are being dropped in the
`SDK` or downstream.

## `transport: Error while dialing dial tcp: missing address`

Logged error message:

```log
2022/03/02 20:29:29 context deadline exceeded
2022/03/02 20:29:29 max retry time elapsed: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp: missing address"
...
```

This error is cause by the OTLP gRPC exporter not being able to connect with
the target endpoint.

1. Make sure the target endpoint is up and receiving connections.
2. Make sure the target endpoint is reachable from the connecting service.
3. Make sure target endpoint is correct if providing an override.
  
  For example, setting the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable
  to a URL without a prefix (i.e.
  `OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317`) needs to be updated to
  include a prefix (`OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317`).

If these steps do not resolve the problem please [open an
issue](#open-an-issue).

## Open An Issue

If you have not found a solution after going through this document, please
[open an Issue]. Be sure to include all the information included in the
template to help address your issue.

[open an Issue]: https://github.com/signalfx/splunk-otel-go/issues/new/choose
[environment variables]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.9.0/specification/sdk-environment-variables.md#batch-span-processor
