# Package `github.com/signalfx/splunk-otel-go/distro`

This package provides a Splunk distribution of the OpenTelemetry Go SDK. It is
designed to provide an SDK properly configured to be used with the Splunk
platform out-of-the-box.

## Getting Started

The main entry point for the package is the [`Run`] function. Use this
function to create an SDK that is ready to be used with OpenTelemetry and
forward all telemetry to Splunk. See [`example_test.go`](./example_test.go) for
a complete example.

Read the official documentation for this distribution in the
[Splunk Docs site](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/get-started.html).
