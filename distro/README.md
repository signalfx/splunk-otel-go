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

By default, the [`SDK`] logs relevant information, warnings, and errors. You
can set the default logging level by setting the `OTEL_LOG_LEVEL` environment
variable to one of the following values:

- `"error"`: Log errors
- `"warn"`: Log errors and warnings
- `"info"`: Log information, warnings, and errors
- `"debug"`: Log debugging and operation information, warnings, and errors

Logging can be explicitly configured in the code that creates the [`SDK`] using
the `WithLogger` option.
