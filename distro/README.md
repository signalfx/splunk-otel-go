# Package `github.com/signalfx/splunk-otel-go/distro`

This package provides a Splunk distribution of the OpenTelemetry Go SDK. It is
designed to provide an SDK properly configured to be used with the Splunk
platform out-of-the-box.

## Getting Started

The main entry point for the package is the [`Run`] function. Use this
function to create an SDK that is ready to be used with OpenTelemetry and
forward all telemetry to Splunk. See [`example_test.go`](./example_test.go) for
a complete example.

## Logging

By default, the [`SDK`] logs relevant information, warnings, and errors. 
You can set the default logging level by setting the `OTEL_LOG_LEVEL` 
environment variable to one of the following values:

- `"error"`: Log errors
- `"warn"`: Log errors and warnings
- `"info"`: Log information, warnings, and errors
- `"debug"`: Log debugging and operation information, warnings, and errors

Logging can be explicitly configured in the code that creates the [`SDK`] 
using the `WithLogger` option.
## Configuration

The [`SDK`] is configured with the following options.

| Option Name | Default Value | Environment Variable |
| ---| --- | --- |
| `WithAccessToken` | `""` | `SPLUNK_ACCESS_TOKEN` |
| `WithEndpoint` | (1) | (2) |
| `WithPropagator` | `tracecontext,baggage` | `OTEL_PROPAGATORS` |
| `WithTraceExporter` | `otlp` | `OTEL_TRACES_EXPORTER` |
| `WithTLSConfig` | none | n/a |
| `WithLogger` | [`zapr`] | n/a |

(1): The default value depends on the exporter used. See the
[`WithEndpoint`](#withendpoint) section for more details.
(2): The environment variable depends on the exporter used. See the
[`WithEndpoint`](#withendpoint) section for more details.

Environment variable can be used to set related option values, but the value
set in code will take precedence. This is the same behavior the default
OpenTelemetry SDK has.

The following sections contain specific information for each option.

### `WithAccessToken`

`WithAccessToken` configures the authentication token used to authenticate
telemetry sent directly to Splunk Observability Cloud.

- Default value: empty (i.e. `""`)
- Environment variable: `SPLUNK_ACCESS_TOKEN`

### `WithEndpoint`

`WithEndpoint` configures the Splunk endpoint that telemetry is sent to.

- Default value: depends on the exporter used.
  - For the `otlp` over gRPC exporter: `"localhost:4317"`
  - For the `jaeger-thrift-splunk` exporter: `"http://127.0.0.1:9080/v1/trace"`
- Environment variable: depends on the exporter used.
  - For the `otlp` over gRPC exporter:
    - `OTEL_EXPORTER_OTLP_ENDPOINT`
    - `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`
  - For the `jaeger-thrift-splunk` exporter: `OTEL_EXPORTER_JAEGER_ENDPOINT`

### `WithPropagator`

`WithPropagator` configures the OpenTelemetry `TextMapPropagator` set as the
global `TextMapPropagator`. Setting to `nil` will prevent any global
`TextMapPropagator` from being set.

- Default value: A W3C tracecontext and baggage `TextMapPropagator`
- Environment variable: `OTEL_PROPAGATORS`

  The environment variable values are restricted to the following.
  - `"tracecontext"`: W3C tracecontext
  - `"baggage"`: W3C baggage
  - `"b3"`: B3 single-header format
  - `"b3multi"`: B3 multi-header format
  - `"jaeger"`: Jaeger
  - `"xray"`: AWS X-Ray
  - `"ottrace"`: OpenTracing
  - `"none"`: None; explicitly do not set any global propagator

  Values can be joined with a comma (`","`) to produce a composite
  `TextMapPropagator`.

### `WithTraceExporter`

`WithTraceExporter` configures the OpenTelemetry trace SpanExporter used to
deliver telemetry. This exporter is registered with the OpenTelemetry SDK using
a batch span processor.

- Default value: an OTLP gRPC SpanExporter.
- Environment variable: `OTEL_TRACES_EXPORTER`

  The environment variable values are restricted to the following.
  - `"otlp"`: OTLP gRPC exporter.
  - `"jaeger-thrift-splunk"`: Jaeger thrift exporter.
  - `"none"`: None, explicitly do not set an exporter.

  Any other value will be ignored and the default used instead.

### `WithTLSConfig`

`WithTLSConfig` configures the TLS configuration used by the exporter.

- Default value: none; an non-TLS connection is used.
- Environment variable: n/a (only configurable in code).

### `WithLogger`

`WithLogger` configures the logger used by this distro.

Configure [`logr.Logger`] with a verbosity value that allows to emit
 [`Info`] logs of the desired level. The SDK uses the following log level 
 to verbosity value correspondences:

| Log level | Verbosity value |
|---|---|
| warning | `0` |
| info | `1` |
| debug | `2` or higher |

The default value, [`zapr`], is configured for `info` logging. This setting is configurable only in code.

[`Run`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#Run
[`SDK`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#SDK
[`zapr`]: https://pkg.go.dev/github.com/go-logr/zapr
[`logr.Logger`]: https://pkg.go.dev/github.com/go-logr/logr#Logger
[`Info`]: https://pkg.go.dev/github.com/go-logr/logr#Logger.Info
