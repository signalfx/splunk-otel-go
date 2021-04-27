# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2021-04-27

The primary change of this release is updating the dependency of `go.opentelemetry.io/otel*` packages from [`v0.19.0`] to [`v0.20.0`].
This includes [a fix](https://github.com/open-telemetry/opentelemetry-go/pull/1830) in the Jaeger exporter.
This fix removes the duplicate batching that the exporter implemented.
Now the `BatchSpanProcessor` that `distro` configures by default will not experience an impedence mismatch with this duplicate batching.

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v0.19.0`] to [`v0.20.0`].

## [0.1.0] - 2021-04-08

### Added

- Add [`distro`](./distro) package providing functionality to quickly setup the OpenTelemetry Go implementation with useful Splunk defaults.
- Add [`splunkhttp`](./instrumentation/net/http/splunkhttp) module providing additional Splunk specific instrumentation for `net/http`.

[Unreleased]: https://github.com/signalfx/splunk-otel-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.2.0
[0.1.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.1.0

[`v0.20.0`]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.20.0
[`v0.19.0`]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.19.0
