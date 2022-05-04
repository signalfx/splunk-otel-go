# Splunk Distribution of OpenTelemetry Go

[![Splunk GDI Specification](https://img.shields.io/badge/GDI-1.2.0-blue)](https://github.com/signalfx/gdi-specification/releases/tag/v1.2.0)
[![GitHub Release](https://img.shields.io/github/v/release/signalfx/splunk-otel-go?include_prereleases)](https://github.com/signalfx/splunk-otel-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/signalfx/splunk-otel-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/signalfx/splunk-otel-go)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/signalfx/splunk-otel-go/ci)](https://github.com/signalfx/splunk-otel-go/actions?query=branch%3Amain)
[![codecov](https://codecov.io/gh/signalfx/splunk-otel-go/branch/main/graph/badge.svg)](https://codecov.io/gh/signalfx/splunk-otel-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/signalfx/splunk-otel-go)](https://goreportcard.com/report/github.com/signalfx/splunk-otel-go)

The Splunk distribution of [OpenTelemetry
Go](https://github.com/open-telemetry/opentelemetry-go) provides
multiple packages that instrument your Go
application to capture and report distributed traces to Splunk APM.

> :construction: This project is currently in **BETA**.
> It is **officially supported** by Splunk.
> However, breaking changes **MAY** be introduced.

## Documentation

Read the official documentation for this distribution in the
[Splunk Docs site](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/get-started.html).

### Examples

You can find our official "user-facing" examples
[here](https://github.com/signalfx/tracing-examples/tree/main/opentelemetry-tracing/opentelemetry-go).

The [example directory](./example) contains smaller, developer focused, examples.
It is meant to be used for experimenting and exploratory end-to-end testing.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md)
before creating an issue or a pull request.

## License

The Splunk distribution of OpenTelemetry Go is a
distribution of the [OpenTelemetry Go
project](https://github.com/open-telemetry/opentelemetry-go). It is
released under the terms of the Apache Software License version 2.0. See [the
license file](./LICENSE) for more details.

>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk
SignalFx](https://www.splunk.com/en_us/investor-relations/acquisitions/signalfx.html)
for more information.
