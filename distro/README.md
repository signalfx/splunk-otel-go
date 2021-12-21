# Package `github.com/signalfx/splunk-otel-go/distro`

This package provides a Splunk distribution of the OpenTelemetry Go SDK. It is
designed to provide an SDK properly configured to be used with the Splunk
platform out-of-the-box.

## Getting Started

The main entry point for the package is the [`Run`][] function. Use this
function to create an SDK that is ready to be used with OpenTelemetry and
forward all telemetry to Splunk. See [`example_test.go`](./example_test.go) for
a complete example.

## Configuration

The [`SDK`][] is configured with the following options.

| Option Name | Default Value | Environment Variable |
| ---| --- | --- |
| `WithAccessToken` | `""` | `SPLUNK_ACCESS_TOKEN` |
| `WithEndpoint` | `"localhost:4317"`(1) | none |
| `WithPropagator` | `tracecontext,baggage` | `OTEL_PROPAGATORS` |

(1): The default value depends on the exporter used. See the
[`WithEndpoint`](#withendpoint) section for more details.

Environment variable can be used to set related option values, but the value
set in code will take precedence. This is the same behavior the default
OpenTelemetry SDK has.

The following sections contain specific information for each option.

### `WithAccessToken`

`WithAccessToken` configures the authentication token used to authenticate
telemetry delivery requests to a Splunk back-end.

- Default value: empty (i.e. `""`)
- Environment variable: `SPLUNK_ACCESS_TOKEN`

### `WithEndpoint`

`WithEndpoint` configures the Splunk endpoint that telemetry is sent to.

- Default value: depends on the exporter used.
  - For the `otlp` over gRPC exporter: `"localhost:4317"`
  - For the `jaeger-thrift-splunk` exporter: `"http://127.0.0.1:9080/v1/trace"`

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
  - `"none"`: None, explicitly do not set a global propagator

  Values can be joined with a comma (`","`) to produce a composite
  `TextMapPropagator`.

[`Run`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#Run
[`SDK`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#SDK
