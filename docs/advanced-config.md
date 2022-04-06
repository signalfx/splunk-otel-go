# Advanced configuration

> The official Splunk documentation for this page is
[Migrate from the SignalFx Tracing Library for Go](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/configuration/advanced-go-otel-configuration.html#advanced-go-otel-configuration).
For instructions on how to contribute to the docs, see
[CONTRIBUTING.md](./CONTRIBUTING.md#documentation).

## Splunk distribution configuration

<!-- markdownlint-disable MD013 -->
| Environment variable      | Option             | Default value  | Description |
| ------------------------- | -------------------| -------------- | ----------- |
| `SPLUNK_ACCESS_TOKEN`     | [`WithAccessToken`](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithAccessToken)  |                | The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html). [[1](#cfg1)] |
| `OTEL_RESOURCE_ATTRIBUTES` |                    |                | Comma-separated list of resource attributes added to every reported span. |
<!-- markdownlint-enable MD013 -->

[<a name="cfg1">1</a>]: The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html)
allows exporters sending data directly to the [Splunk Observability Cloud](https://dev.splunk.com/observability/docs/apibasics/api_list/).
To do so, the `OTEL_EXPORTER_JAEGER_ENDPOINT` must be set
or `distro.WithEndpoint` must be passed to `distro.Run`
with Splunk back-end ingest endpoint URL: `https://ingest.<REALM>.signalfx.com/v2/trace`.

## Trace configuration

<!-- markdownlint-disable MD013 -->
| Environment variable       | Option             | Default value  | Description |
| -------------------------- | -------------------| -------------- | ----------- |
| `OTEL_RESOURCE_ATTRIBUTES` |                    |                | Comma-separated list of resource attributes added to every reported span. |
<!-- markdownlint-enable MD013 -->

## Trace exporter configuration

<!-- markdownlint-disable MD013 -->
| Environment variable            | Option             | Default value  | Description |
| ------------------------------- | -------------------| -------------- | ----------- |
| `OTEL_EXPORTER_JAEGER_ENDPOINT` | [`WithEndpoint`](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithEndpoint)     | `http://localhost:14268/api/traces` | Jaeger Thrift HTTP endpoint for sending spans. |
| `OTEL_EXPORTER_JAEGER_USER`     |                    |                | Username to be used for HTTP basic authentication. |
| `OTEL_EXPORTER_JAEGER_PASSWORD` |                    |                | Password to be used for HTTP basic authentication. |
<!-- markdownlint-enable MD013 -->

## Trace propagation configuration

The trace propagator can be changed by using
[`otel.SetTextMapPropagator`](https://pkg.go.dev/go.opentelemetry.io/otel#SetTextMapPropagator)
after `distro.Run()` is invoked e.g.:

```go
distro.Run()
otel.SetTextMapPropagator(propagation.TraceContext{})
```
