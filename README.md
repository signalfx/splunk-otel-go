# Splunk distribution of OpenTelemetry Go

The Splunk distribution of [OpenTelemetry
Go](https://github.com/open-telemetry/opentelemetry-go) provides
multiple installable packages that automatically instruments your Go
application to capture and report distributed traces to Splunk APM.

This Splunk distribution comes with the following defaults:

- [B3 context propagation](https://github.com/openzipkin/b3-propagation).
- [Jaeger thrift
  exporter](https://opentelemetry-python.readthedocs.io/en/stable/exporter/jaeger/jaeger.html)
  configured to send spans to a locally running [SignalFx Smart
  Agent](https://docs.signalfx.com/en/latest/apm/apm-getting-started/apm-smart-agent.html)
  (`http://localhost:9080/v1/trace`).
- Unlimited default limits for [configuration options](#trace-configuration) to
  support full-fidelity traces.

> :construction: This project is currently in **BETA**.

## Getting Started

Supported libraries are listed
[here](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).

TODO

## Manually instrument an application

Documentation on how to manually instrument a Go application is available
[here](https://opentelemetry.io/docs/go/getting-started/).

## License and versioning

The Splunk distribution of OpenTelemetry Go is a
distribution of the [OpenTelemetry Go
project](https://github.com/open-telemetry/opentelemetry-go). It is
released under the terms of the Apache Software License version 2.0. See [the
license file](./LICENSE) for more details.
