# Troubleshooting

> The official Splunk documentation for this page is
[Troubleshooting Go instrumentation](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/troubleshooting/common-go-troubleshooting.html).
For instructions on how to contribute to the docs, see
[CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

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

### Missing all spans from a service

If you see no spans in the Splunk Observability Cloud for your service
ensure the following first:

1. Make sure to wait a few minutes. There might be delays in the telemetry
   pipeline. Wait a few minutes and check again.
2. Ensure your service is named properly. Check that your service names are not
   showing up in the Splunk Observability Cloud interface with a prefix of
   `unknown_service:` (for example, `unknown_service:go` or
   `unknown_service:<your-programs-binary-name>`). If so, set the
   `OTEL_SERVICE_NAME` environment variable to your service's name and restart
   your application.

If you have taken these steps and still are not able to see spans, [enable
debug logging](#enable-debug-logging). This might produce a log message like
the following:

```text
debug	global/internal_logging.go:62	exporting spans	{"count": 154, "total_dropped": 0}
```

The `count` in the log message is the number of spans that were exported for a
batch by the `SDK`.

If `count` is non-zero that means the `SDK` is exporting span. If this is the
case, spans are being dropped downstream. Refer to [the collector
troubleshooting documentation].

If `count` is zero, that means the `SDK` is not exporting any spans. If this is
the case, ensure that all spans your application is creating are ended (for example,
by using `span.End()`).

The `total_dropped` value of the log message is the cumulative number of spans
the `SDK` has dropped. If this value is non-zero see the [missing some spans
from a service](#missing-some-spans-from-a-service) section for more
information on how to resolve this.

### Missing some spans from a service

If you see traces from your service in Splunk Observability Cloud that miss
spans, you might need to configure the `BatchSpanProcessor`. Verify whether
spans are being dropped by [enabling debug logging](#enable-debug-logging).
This should produce a log message like the following:

```text
debug	global/internal_logging.go:62	exporting spans	{"count": 364, "total_dropped": 1320}
```

The `total_dropped` value is the cumulative number of spans dropped by the
`SDK`. If the value is greater than zero, that means you need to reconfigured
the `BatchSpanProcessor`.

The `BatchSpanProcessor` has the following configuration parameters:

| Configuration Parameter                        | Default | Environment Variable             |
| ---------------------------------------------- | ------- | -------------------------------- |
| Delay interval between two consecutive exports | 5000    | `OTEL_BSP_SCHEDULE_DELAY`        |
| Maximum allowed time to export data            | 30000   | `OTEL_BSP_EXPORT_TIMEOUT`        |
| Maximum queue size                             | 2048    | `OTEL_BSP_MAX_QUEUE_SIZE`        |
| Maximum batch size                             | 512     | `OTEL_BSP_MAX_EXPORT_BATCH_SIZE` |

The `BatchSpanProcessor` drops new spans when the queue is full. There are two
reasons this will occur:

- Spans are being added faster than they can be exported.
- Exporting is taking so long the queue fills during the export.

If the value of `count` in the log messages is consistently equal to the
maximum batch size then instrumentation might be creating spans faster than
they can be exported.

One way to resolve this is to provide additional computational and network
resources. If your system has enough resources, increasing the batch size to
use more network bandwidth per export and increase the queue size to hold a
bigger buffer. For example:

```sh
export OTEL_BSP_MAX_EXPORT_BATCH_SIZE=1024
export OTEL_BSP_MAX_QUEUE_SIZE=20480
```

If the system has limited memory, do not increase the maximum queue size.

If the network has no bandwidth to spare, reduce your export batch size. For
example:

```sh
export OTEL_BSP_MAX_EXPORT_BATCH_SIZE=128
```

This might increase the export frequency and drain the queue faster.

If the value of `count` is not consistently equal to the maximum batch size,
the bottleneck might be the export process. The `SDK` might be taking so long
to export a batch that more spans than the queue can hold are added during the
process. This might be caused by an underlying network issue. Make sure you
have a stable network to the target and that you have adequate bandwidth. You
can also reduce export timeouts, decrease the export size and frequency, and
increase the queue size. For example:

```sh
# 5s export timeout.
export OTEL_BSP_EXPORT_TIMEOUT=5000
# 30s maximum time between exports.
export OTEL_BSP_SCHEDULE_DELAY=30000
export OTEL_BSP_MAX_QUEUE_SIZE=5120
export OTEL_BSP_MAX_EXPORT_BATCH_SIZE=128
```

Make sure to allocate enough memory resources on your system to accommodate the
increase in queue size. Changes in the export configuration might result in the
`SDK` dropping whole export batches that take too long.

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
[the collector troubleshooting documentation]: https://github.com/signalfx/splunk-otel-collector/blob/main/docs/troubleshooting.md
