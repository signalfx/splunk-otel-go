# Troubleshooting

## Enable debug logging

When things are not working, a good first step is to restart the program with
debug logging enabled. Do this by setting the `OTEL_LOG_LEVEL` environment
variable to `debug`.

```sh
export OTEL_LOG_LEVEL="debug"
```

Make sure to unset the environment variable after the issue is resolved, as its output
might overload systems if left on indefinitely.

## Missing spans

Spans might be dropped, or lost, due to several reasons. Follow these steps to ensure
spans are not being dropped by the `SDK`.

1. [Enable debug logging](#enable-debug-logging). This logs the `SDK`
   configuration and all spans sent to registered `SpanProcessor`s. Verify the
   that spans are exported in the log messages. Check also the`total_dropped` count.
2. If an incorrect number of spans are being reported as exported, verify you
   are ending the spans you expect to see.
3. If spans are reported as dropped, the `BatchSpanProcessor` is dropping
   spans. Increase the capacity of the `BatchSpanProcessor`'s queue or the
   frequency it exports via [environment variables].

Following the previous steps might help understand if spans are being dropped in the
`SDK` or downstream.

## `transport: Error while dialing dial tcp: missing address`

If you get the following logged error message:

```log
2022/03/02 20:29:29 context deadline exceeded
2022/03/02 20:29:29 max retry time elapsed: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp: missing address"
...
```

This error is caused by the OTLP gRPC exporter not being able to connect with
the target endpoint.

1. Make sure the target endpoint is up and receiving connections.
2. Make sure the target endpoint is reachable from the connecting service.
3. Make sure target endpoint is correct if providing an alternative value.
  
  For example, to set the endpoint URL to `otel-collector:4317` you must include
  a prefix: `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317`.

If the previous steps don't resolve the problem, [open an
issue](#open-an-issue).

## Open an issue

If you have not found a solution after going through this document,
[open an Issue]. Be sure to include all the required information.

[open an Issue]: https://github.com/signalfx/splunk-otel-go/issues/new/choose
[environment variables]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.9.0/specification/sdk-environment-variables.md#batch-span-processor
